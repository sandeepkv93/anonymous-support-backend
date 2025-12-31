FROM mcr.microsoft.com/devcontainers/go:1-1.21-bullseye

# Install additional tools
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    postgresql-client \
    mongodb-clients \
    redis-tools \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install Go tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install golang.org/x/vuln/cmd/govulncheck@latest \
    && go install github.com/bufbuild/buf/cmd/buf@latest \
    && go install github.com/go-task/task/v3/cmd/task@latest

USER vscode
