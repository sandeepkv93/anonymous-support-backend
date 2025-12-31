// +build load

package load

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/anonymous-support/internal/app"
	"github.com/yourorg/anonymous-support/internal/domain"
)

const (
	numConcurrentUsers = 100
	testDuration       = 30 * time.Second
)

type LoadTestMetrics struct {
	totalRequests     int64
	successfulReqs    int64
	failedReqs        int64
	totalLatency      int64
	maxLatency        int64
	minLatency        int64
	errors            []error
	mu                sync.Mutex
}

func (m *LoadTestMetrics) recordRequest(latency time.Duration, err error) {
	atomic.AddInt64(&m.totalRequests, 1)
	latencyMs := latency.Milliseconds()
	atomic.AddInt64(&m.totalLatency, latencyMs)

	// Update max latency
	for {
		current := atomic.LoadInt64(&m.maxLatency)
		if latencyMs <= current || atomic.CompareAndSwapInt64(&m.maxLatency, current, latencyMs) {
			break
		}
	}

	// Update min latency (initialize to max value first)
	if m.minLatency == 0 {
		atomic.StoreInt64(&m.minLatency, latencyMs)
	}
	for {
		current := atomic.LoadInt64(&m.minLatency)
		if latencyMs >= current || atomic.CompareAndSwapInt64(&m.minLatency, current, latencyMs) {
			break
		}
	}

	if err != nil {
		atomic.AddInt64(&m.failedReqs, 1)
		m.mu.Lock()
		m.errors = append(m.errors, err)
		m.mu.Unlock()
	} else {
		atomic.AddInt64(&m.successfulReqs, 1)
	}
}

func (m *LoadTestMetrics) report(t *testing.T) {
	totalReqs := atomic.LoadInt64(&m.totalRequests)
	successReqs := atomic.LoadInt64(&m.successfulReqs)
	failedReqs := atomic.LoadInt64(&m.failedReqs)
	totalLat := atomic.LoadInt64(&m.totalLatency)
	maxLat := atomic.LoadInt64(&m.maxLatency)
	minLat := atomic.LoadInt64(&m.minLatency)

	avgLatency := float64(0)
	if totalReqs > 0 {
		avgLatency = float64(totalLat) / float64(totalReqs)
	}

	successRate := float64(successReqs) / float64(totalReqs) * 100
	throughput := float64(totalReqs) / testDuration.Seconds()

	t.Logf("\n=== Load Test Results ===")
	t.Logf("Total Requests: %d", totalReqs)
	t.Logf("Successful: %d (%.2f%%)", successReqs, successRate)
	t.Logf("Failed: %d", failedReqs)
	t.Logf("Throughput: %.2f req/s", throughput)
	t.Logf("Average Latency: %.2f ms", avgLatency)
	t.Logf("Min Latency: %d ms", minLat)
	t.Logf("Max Latency: %d ms", maxLat)

	if len(m.errors) > 0 {
		t.Logf("Sample errors (first 5):")
		for i := 0; i < len(m.errors) && i < 5; i++ {
			t.Logf("  - %v", m.errors[i])
		}
	}
}

func TestLoadUserRegistration(t *testing.T) {
	cfg := &app.Config{
		Environment: "load_test",
		PostgresDSN: "postgres://localhost:5432/anonymous_support_test?sslmode=disable",
		RedisDSN:    "localhost:6379",
		MongoDSN:    "mongodb://localhost:27017/anonymous_support_test",
	}

	testApp, err := app.NewApp(cfg)
	assert.NoError(t, err)
	defer testApp.Close()

	metrics := &LoadTestMetrics{
		minLatency: 0,
	}

	ctx := context.Background()
	startTime := time.Now()
	var wg sync.WaitGroup

	// Spawn concurrent users
	for i := 0; i < numConcurrentUsers; i++ {
		wg.Add(1)
		go func(userNum int) {
			defer wg.Done()

			counter := 0
			for time.Since(startTime) < testDuration {
				username := fmt.Sprintf("load_test_user_%d_%d", userNum, counter)
				counter++

				reqStart := time.Now()
				_, err := testApp.AuthService.RegisterAnonymous(ctx, username)
				latency := time.Since(reqStart)

				metrics.recordRequest(latency, err)
			}
		}(i)
	}

	wg.Wait()
	metrics.report(t)

	// Assert success rate is above threshold
	successRate := float64(metrics.successfulReqs) / float64(metrics.totalRequests) * 100
	assert.Greater(t, successRate, 95.0, "Success rate should be above 95%")
}

func TestLoadPostCreation(t *testing.T) {
	cfg := &app.Config{
		Environment: "load_test",
		PostgresDSN: "postgres://localhost:5432/anonymous_support_test?sslmode=disable",
		RedisDSN:    "localhost:6379",
		MongoDSN:    "mongodb://localhost:27017/anonymous_support_test",
	}

	testApp, err := app.NewApp(cfg)
	assert.NoError(t, err)
	defer testApp.Close()

	// Pre-create test users
	ctx := context.Background()
	users := make([]string, numConcurrentUsers)
	for i := 0; i < numConcurrentUsers; i++ {
		authResp, err := testApp.AuthService.RegisterAnonymous(ctx, fmt.Sprintf("post_load_user_%d", i))
		assert.NoError(t, err)
		users[i] = authResp.User.ID
	}

	metrics := &LoadTestMetrics{
		minLatency: 0,
	}

	startTime := time.Now()
	var wg sync.WaitGroup

	// Concurrent post creation
	for i, userID := range users {
		wg.Add(1)
		go func(userID string, userNum int) {
			defer wg.Done()

			counter := 0
			for time.Since(startTime) < testDuration {
				content := fmt.Sprintf("Load test post %d from user %d", counter, userNum)
				counter++

				reqStart := time.Now()
				_, err := testApp.PostService.CreatePost(
					ctx,
					userID,
					fmt.Sprintf("post_load_user_%d", userNum),
					domain.PostTypeSOS,
					content,
					[]string{"load-test"},
					5,
					"afternoon",
					5,
					[]string{},
					"public",
					nil,
				)
				latency := time.Since(reqStart)

				metrics.recordRequest(latency, err)
			}
		}(userID, i)
	}

	wg.Wait()
	metrics.report(t)

	// Assert reasonable performance
	avgLatency := float64(metrics.totalLatency) / float64(metrics.totalRequests)
	assert.Less(t, avgLatency, 200.0, "Average latency should be under 200ms")

	successRate := float64(metrics.successfulReqs) / float64(metrics.totalRequests) * 100
	assert.Greater(t, successRate, 90.0, "Success rate should be above 90%")
}

func TestLoadFeedRetrieval(t *testing.T) {
	cfg := &app.Config{
		Environment: "load_test",
		PostgresDSN: "postgres://localhost:5432/anonymous_support_test?sslmode=disable",
		RedisDSN:    "localhost:6379",
		MongoDSN:    "mongodb://localhost:27017/anonymous_support_test",
	}

	testApp, err := app.NewApp(cfg)
	assert.NoError(t, err)
	defer testApp.Close()

	ctx := context.Background()

	// Pre-populate with some posts
	authResp, _ := testApp.AuthService.RegisterAnonymous(ctx, "feed_load_user")
	for i := 0; i < 50; i++ {
		testApp.PostService.CreatePost(
			ctx,
			authResp.User.ID,
			"feed_load_user",
			domain.PostTypeSOS,
			fmt.Sprintf("Pre-populated post %d", i),
			[]string{"feed-test"},
			5,
			"afternoon",
			5,
			[]string{},
			"public",
			nil,
		)
	}

	metrics := &LoadTestMetrics{
		minLatency: 0,
	}

	startTime := time.Now()
	var wg sync.WaitGroup

	// Concurrent feed reads
	for i := 0; i < numConcurrentUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for time.Since(startTime) < testDuration {
				reqStart := time.Now()
				_, err := testApp.PostService.GetFeed(ctx, []string{"feed-test"}, nil, nil, 20, 0)
				latency := time.Since(reqStart)

				metrics.recordRequest(latency, err)
			}
		}()
	}

	wg.Wait()
	metrics.report(t)

	// Feed reads should be very fast with caching
	avgLatency := float64(metrics.totalLatency) / float64(metrics.totalRequests)
	assert.Less(t, avgLatency, 100.0, "Average feed latency should be under 100ms with caching")
}
