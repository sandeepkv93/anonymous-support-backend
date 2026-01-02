# Tests

## Integration tests

Integration tests live under `tests/integration` and exercise the running API via Connect RPC.

### Run with a locally running stack

1. Start the services:

```bash
docker-compose up -d
```

2. Run the integration tests:

```bash
go test -v -run Integration ./tests/integration/...
```

If your API is running somewhere other than `http://localhost:8080`, set:

```bash
INTEGRATION_BASE_URL=http://localhost:8080
```

### Let the tests start docker-compose

Set `INTEGRATION_DOCKER_COMPOSE=1` to have the tests spin up the docker-compose stack automatically:

```bash
INTEGRATION_DOCKER_COMPOSE=1 go test -v -run Integration ./tests/integration/...
```
