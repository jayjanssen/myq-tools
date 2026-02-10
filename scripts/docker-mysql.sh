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
