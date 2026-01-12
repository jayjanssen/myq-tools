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
