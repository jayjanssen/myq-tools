# Local Docker Integration Tests Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enable running integration tests locally against MySQL 8.0, 8.4, and 9.5 in Docker containers.

**Architecture:** Two bash scripts manage container lifecycle and test execution. Makefile targets provide the user interface. Containers use random ports discovered at runtime.

**Tech Stack:** Bash, Docker, Make, Go test with `-tags=integration`

---

### Task 1: Create scripts directory

**Files:**
- Create: `scripts/.gitkeep` (placeholder, will be replaced)

**Step 1: Create directory**

```bash
mkdir -p scripts
```

**Step 2: Verify**

Run: `ls -la scripts/`
Expected: Empty directory exists

**Step 3: Commit**

```bash
git add scripts
git commit -m "chore: add scripts directory"
```

---

### Task 2: Create docker-mysql.sh script

**Files:**
- Create: `scripts/docker-mysql.sh`

**Step 1: Write the script**

```bash
#!/bin/bash
set -e

# MySQL versions to test (must match .github/workflows/test.yml)
MYSQL_VERSIONS=("8.0" "8.4" "9.5")
CONTAINER_PREFIX="myq-mysql"

wait_for_mysql() {
    echo "Waiting for MySQL containers to be ready..."

    for version in "${MYSQL_VERSIONS[@]}"; do
        name="${CONTAINER_PREFIX}-${version}"
        port=$(docker port "$name" 3306 2>/dev/null | cut -d: -f2)

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

up() {
    for version in "${MYSQL_VERSIONS[@]}"; do
        name="${CONTAINER_PREFIX}-${version}"

        # Skip if already running
        if docker ps -q -f name="^${name}$" | grep -q .; then
            port=$(docker port "$name" 3306 | cut -d: -f2)
            echo "Container $name already running on port $port"
            continue
        fi

        # Remove stopped container if exists
        docker rm -f "$name" 2>/dev/null || true

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
        docker rm -f "$name" 2>/dev/null && echo "Removed $name" || true
    done
}

status() {
    for version in "${MYSQL_VERSIONS[@]}"; do
        name="${CONTAINER_PREFIX}-${version}"
        if docker ps -q -f name="^${name}$" | grep -q .; then
            port=$(docker port "$name" 3306 | cut -d: -f2)
            echo "$name: running on port $port"
        else
            echo "$name: not running"
        fi
    done
}

case "${1:-}" in
    up)     up ;;
    down)   down ;;
    status) status ;;
    *)      echo "Usage: $0 {up|down|status}"; exit 1 ;;
esac
```

**Step 2: Make executable**

```bash
chmod +x scripts/docker-mysql.sh
```

**Step 3: Test up command (requires Docker)**

Run: `./scripts/docker-mysql.sh up`
Expected: Three containers start, waits for ready, shows ports

**Step 4: Test status command**

Run: `./scripts/docker-mysql.sh status`
Expected: Shows all three containers running with ports

**Step 5: Test down command**

Run: `./scripts/docker-mysql.sh down`
Expected: All three containers removed

**Step 6: Commit**

```bash
git add scripts/docker-mysql.sh
git commit -m "feat: add docker-mysql.sh for container lifecycle"
```

---

### Task 3: Create run-integration-tests.sh script

**Files:**
- Create: `scripts/run-integration-tests.sh`

**Step 1: Write the script**

```bash
#!/bin/bash
set -e

MYSQL_VERSIONS=("8.0" "8.4" "9.5")
CONTAINER_PREFIX="myq-mysql"
OVERALL_EXIT=0

# Check if containers are running
for version in "${MYSQL_VERSIONS[@]}"; do
    name="${CONTAINER_PREFIX}-${version}"
    if ! docker ps -q -f name="^${name}$" | grep -q .; then
        echo "ERROR: Container $name is not running"
        echo "Run 'make docker-mysql-up' first"
        exit 1
    fi
done

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
    else
        echo ""
        echo "PASSED: MySQL $version"
    fi
done

echo ""
echo "========================================"
if [ $OVERALL_EXIT -eq 0 ]; then
    echo "All MySQL versions passed"
else
    echo "Some MySQL versions failed"
fi
echo "========================================"

exit $OVERALL_EXIT
```

**Step 2: Make executable**

```bash
chmod +x scripts/run-integration-tests.sh
```

**Step 3: Commit**

```bash
git add scripts/run-integration-tests.sh
git commit -m "feat: add run-integration-tests.sh for multi-version testing"
```

---

### Task 4: Add Makefile targets

**Files:**
- Modify: `Makefile`

**Step 1: Add new targets after existing test-integration target**

Add these lines after the `test-integration` target:

```makefile
# Start MySQL containers for integration tests (8.0, 8.4, 9.5)
docker-mysql-up:
	@./scripts/docker-mysql.sh up

# Stop and remove MySQL containers
docker-mysql-down:
	@./scripts/docker-mysql.sh down

# Show MySQL container status
docker-mysql-status:
	@./scripts/docker-mysql.sh status

# Run integration tests against all Docker MySQL versions
test-integration-docker: docker-mysql-up
	@./scripts/run-integration-tests.sh
```

**Step 2: Update .PHONY line**

Add new targets to the `.PHONY` line at the top of the Makefile:

```makefile
.PHONY: install test test-race test-verbose test-coverage test-integration docker-mysql-up docker-mysql-down docker-mysql-status test-integration-docker benchmark benchmark-ci build clean
```

**Step 3: Commit**

```bash
git add Makefile
git commit -m "feat: add Makefile targets for Docker integration tests"
```

---

### Task 5: End-to-end verification

**Step 1: Run full workflow**

```bash
make test-integration-docker
```

Expected:
- Containers start (or report already running)
- Tests run against 8.0, 8.4, 9.5 sequentially
- Clear headers for each version
- Final summary shows pass/fail

**Step 2: Verify teardown**

```bash
make docker-mysql-down
make docker-mysql-status
```

Expected: All containers removed, status shows "not running"

**Step 3: Final commit if any adjustments needed**

---

## Summary

| Target | Purpose |
|--------|---------|
| `make docker-mysql-up` | Start MySQL 8.0, 8.4, 9.5 containers |
| `make docker-mysql-down` | Stop and remove containers |
| `make docker-mysql-status` | Show container status and ports |
| `make test-integration-docker` | Run tests against all versions |
| `make test-integration` | Existing: test against env vars or localhost:3306 |
