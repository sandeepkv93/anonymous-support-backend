// Load test for feed endpoint using k6
// Run with: k6 run tests/load/feed_test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '1m', target: 50 },  // Ramp up to 50 users
    { duration: '3m', target: 50 },  // Stay at 50 users
    { duration: '1m', target: 100 }, // Ramp up to 100 users
    { duration: '3m', target: 100 }, // Stay at 100 users
    { duration: '1m', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.01'],   // Error rate must be less than 1%
    errors: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test data
const testUsers = [
  { email: 'loadtest1@example.com', password: 'TestPass123!' },
  { email: 'loadtest2@example.com', password: 'TestPass123!' },
  { email: 'loadtest3@example.com', password: 'TestPass123!' },
];

function login() {
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];

  const loginRes = http.post(`${BASE_URL}/api.auth.v1.AuthService/Login`, JSON.stringify({
    email: user.email,
    password: user.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(loginRes, {
    'login successful': (r) => r.status === 200,
  });

  if (!success) {
    errorRate.add(1);
    return null;
  }

  const body = JSON.parse(loginRes.body);
  return body.accessToken;
}

export default function () {
  // Login and get token
  const token = login();
  if (!token) {
    sleep(1);
    return;
  }

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // Test 1: Get feed
  const feedRes = http.get(
    `${BASE_URL}/api.post.v1.PostService/GetFeed?limit=20&offset=0`,
    { headers }
  );

  check(feedRes, {
    'feed status is 200': (r) => r.status === 200,
    'feed response time OK': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);

  sleep(1);

  // Test 2: Get feed with category filter
  const categorizedFeedRes = http.get(
    `${BASE_URL}/api.post.v1.PostService/GetFeed?limit=20&offset=0&category=alcohol`,
    { headers }
  );

  check(categorizedFeedRes, {
    'categorized feed status is 200': (r) => r.status === 200,
    'categorized feed response time OK': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);

  sleep(1);

  // Test 3: Create a post (10% of requests)
  if (Math.random() < 0.1) {
    const postRes = http.post(
      `${BASE_URL}/api.post.v1.PostService/CreatePost`,
      JSON.stringify({
        type: 'CheckIn',
        content: 'Load test post',
        categories: ['test'],
        urgencyLevel: 1,
      }),
      { headers }
    );

    check(postRes, {
      'create post status is 200': (r) => r.status === 200,
    }) || errorRate.add(1);
  }

  sleep(2);
}

export function handleSummary(data) {
  return {
    'load-test-results.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, options) {
  // Simple text summary
  let summary = '\n';
  summary += '================== Load Test Summary ==================\n';
  summary += `Total Requests: ${data.metrics.http_reqs.values.count}\n`;
  summary += `Failed Requests: ${data.metrics.http_req_failed.values.rate * 100}%\n`;
  summary += `Avg Duration: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  summary += `P95 Duration: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += `P99 Duration: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n`;
  summary += '======================================================\n';
  return summary;
}
