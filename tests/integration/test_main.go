package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

const dockerComposeEnv = "INTEGRATION_DOCKER_COMPOSE"

func TestMain(m *testing.M) {
	if os.Getenv(dockerComposeEnv) == "1" {
		if err := runDockerCompose(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to run docker-compose: %v\n", err)
			os.Exit(1)
		}
		if err := waitForReady(integrationBaseURL()); err != nil {
			fmt.Fprintf(os.Stderr, "server not ready: %v\n", err)
			_ = runDockerComposeDown()
			os.Exit(1)
		}
	}

	code := m.Run()

	if os.Getenv(dockerComposeEnv) == "1" {
		_ = runDockerComposeDown()
	}

	os.Exit(code)
}

func runDockerCompose() error {
	cmd := exec.Command("docker-compose", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runDockerComposeDown() error {
	cmd := exec.Command("docker-compose", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForReady(baseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 2 * time.Second}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/ready", nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err == nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		time.Sleep(1 * time.Second)
	}
}
