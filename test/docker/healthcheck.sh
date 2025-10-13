#!/bin/sh
set -e

echo "Health check: Starting health check on port ${NLT_PORT:-8080}"

# Check if the server is responding
echo "Health check: Testing server response..."
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${NLT_PORT:-8080}/api/config || echo "000")
echo "Health check: Got response code: $response"

if [ "$response" != "200" ]; then
    echo "Health check: Server not responding, got HTTP $response"
    echo "Health check: Trying to get more details..."
    curl -v http://localhost:${NLT_PORT:-8080}/api/config || echo "Health check: curl failed completely"
    exit 1
fi

# Check if config file exists
echo "Health check: Checking config file..."
if [ ! -f "/config/config.yaml" ]; then
    echo "Health check: Config file missing at /config/config.yaml"
    echo "Health check: Contents of /config:"
    ls -la /config/ || echo "Health check: Cannot list /config"
    exit 1
fi

# Check if we can read the config file
echo "Health check: Testing config file readability..."
if ! cat "/config/config.yaml" > /dev/null 2>&1; then
    echo "Health check: Cannot read config file"
    exit 1
fi

# All checks passed
echo "Health check: All checks passed"
exit 0 