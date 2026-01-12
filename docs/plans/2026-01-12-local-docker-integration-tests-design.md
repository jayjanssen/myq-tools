# Local Docker Integration Tests Design

## Overview

Run integration tests locally against multiple MySQL versions using Docker, matching the versions tested in CI.

## Goals

- Test against MySQL 8.0, 8.4, and 9.5 locally before pushing
- Match exactly what GitHub Actions tests
- Support both local MySQL and Docker-based testing workflows

## Design

### Makefile Targets

```makefile
# Start MySQL containers for all test versions (8.0, 8.4, 9.5)
docker-mysql-up:
	@./scripts/docker-mysql.sh up

# Stop and remove MySQL containers
docker-mysql-down:
	@./scripts/docker-mysql.sh down

# Run integration tests against all Docker MySQL versions
test-integration-docker:
	@./scripts/docker-mysql.sh up
	@./scripts/run-integration-tests.sh
```

Existing `test-integration` target unchanged (uses env vars or localhost:3306).

### Container Management (`scripts/docker-mysql.sh`)

```bash
#!/bin/bash
# MySQL versions to test (must match .github/workflows/test.yml)
MYSQL_VERSIONS=("8.0" "8.4" "9.5")
CONTAINER_PREFIX="myq-mysql"

up() {
    for version in "${MYSQL_VERSIONS[@]}"; do
        name="${CONTAINER_PREFIX}-${version}"

        # Skip if already running
        if docker ps -q -f name="$name" | grep -q .; then
            echo "Container $name already running"
            continue
        fi

        echo "Starting MySQL $version..."
        docker run -d --name "$name" \
            -e MYSQL_ROOT_PASSWORD=testpass \
            -e MYSQL_DATABASE=testdb \
            -P \
            mysql:"$version"
    done

    wait_for_mysql
}

down() {
    for version in "${MYSQL_VERSIONS[@]}"; do
        name="${CONTAINER_PREFIX}-${version}"
        docker rm -f "$name" 2>/dev/null && echo "Removed $name"
    done
}

wait_for_mysql() {
    echo "Waiting for MySQL containers to be ready..."

    for version in "${MYSQL_VERSIONS[@]}"; do
        name="${CONTAINER_PREFIX}-${version}"
        port=$(docker port "$name" 3306 | cut -d: -f2)

        for i in {1..30}; do
            if docker exec "$name" mysqladmin ping -h localhost --silent 2>/dev/null; then
                echo "MySQL $version ready on port $port"
                break
            fi

            if [ $i -eq 30 ]; then
                echo "ERROR: MySQL $version failed to start"
                exit 1
            fi

            sleep 2
        done
    done
}

case "$1" in
    up)   up ;;
    down) down ;;
    *)    echo "Usage: $0 {up|down}"; exit 1 ;;
esac
```

### Test Runner (`scripts/run-integration-tests.sh`)

```bash
#!/bin/bash
MYSQL_VERSIONS=("8.0" "8.4" "9.5")
CONTAINER_PREFIX="myq-mysql"
OVERALL_EXIT=0

for version in "${MYSQL_VERSIONS[@]}"; do
    name="${CONTAINER_PREFIX}-${version}"

    # Discover the random port Docker assigned
    port=$(docker port "$name" 3306 | cut -d: -f2)

    echo ""
    echo "========================================"
    echo "Testing MySQL $version (port $port)"
    echo "========================================"

    export MYSQL_HOST=127.0.0.1
    export MYSQL_PORT=$port
    export MYSQL_USER=root
    export MYSQL_PASSWORD=testpass

    if ! go test -tags=integration ./...; then
        OVERALL_EXIT=1
        echo ""
        echo "FAILED: MySQL $version"
        echo "To debug, connect with:"
        echo "  mysql -h 127.0.0.1 -P $port -u root -ptestpass"
    fi
done

exit $OVERALL_EXIT
```

## File Structure

```
scripts/
├── docker-mysql.sh          # Container lifecycle (up/down/wait)
└── run-integration-tests.sh # Sequential test runner
```

## Workflow

```bash
# Run tests against all MySQL versions
make test-integration-docker

# Containers stay running for repeated test runs
make test-integration-docker

# Tear down when done
make docker-mysql-down

# Quick test against local MySQL (existing behavior)
make test-integration
```

## Decisions

- **Long-running containers**: Stay up between test runs for faster iteration
- **Random ports**: Docker assigns ports to avoid conflicts; discovered at runtime
- **Sequential execution**: Clear headers per version, easy to read output
- **Failure handling**: Shows mysql connection command for debugging, continues testing remaining versions
- **Version list hardcoded**: In both workflow and scripts; simple, versions rarely change
