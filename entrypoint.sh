#!/bin/sh
# entrypoint.sh

set -e

WSL_IP=$(ip route | awk '/^default via/{print $3}')
export APP_HOST=$WSL_IP

if [ "$RUN_INTEGRATION_TESTS" = "true" ]; then
  echo "Running integration tests..."
  go test -v -tags integration -coverprofile=coverage.out -coverpkg=./src/... ./test/integration
  go tool cover -html=coverage.out -o coverage.html
  exit 0
fi

echo "Starting the application..."
exec ./account-service -port=3000 -host=localhost