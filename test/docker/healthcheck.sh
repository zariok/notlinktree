#!/bin/sh
set -e

# Check if the server is responding
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${NLT_PORT:-8080}/api/config)
if [ "$response" != "200" ]; then
    echo "Health check: Server not responding, got HTTP $response"
    exit 1
fi

# Check if config file exists
if [ ! -f "/config/config.yaml" ]; then
    echo "Health check: Config file missing"
    exit 1
fi

# Check if we can read the config file
if ! cat "/config/config.yaml" > /dev/null 2>&1; then
    echo "Health check: Cannot read config file"
    exit 1
fi

# All checks passed
echo "Health check: All checks passed"
exit 0 