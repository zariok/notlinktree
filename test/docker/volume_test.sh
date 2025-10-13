#!/bin/sh
set -e

# Test that the main service is responding
test_main_service() {
    echo "Testing main service (notlinktree)..."
    
    # Test basic connectivity
    echo "Attempting to connect to notlinktree:8080/api/config..."
    response=$(curl -s -o /dev/null -w "%{http_code}" http://notlinktree:8080/api/config || echo "000")
    echo "Got response code: $response"
    if [ "$response" != "200" ]; then
        echo "Main service not responding, got HTTP $response"
        echo "Trying to get more details..."
        curl -v http://notlinktree:8080/api/config || echo "curl failed completely"
        exit 1
    fi
    
    # Test that we get valid JSON
    echo "Testing JSON response..."
    response=$(curl -s http://notlinktree:8080/api/config)
    echo "Response: $response"
    if ! echo "$response" | jq -e ".data.title" > /dev/null; then
        echo "Main service not returning valid JSON with expected structure"
        exit 1
    fi
    
    echo "Main service is working correctly"
}

# Test that the local service is responding
test_local_service() {
    echo "Testing local service (notlinktree_local)..."
    
    # Test basic connectivity
    echo "Attempting to connect to notlinktree_local:8080/api/config..."
    response=$(curl -s -o /dev/null -w "%{http_code}" http://notlinktree_local:8080/api/config || echo "000")
    echo "Got response code: $response"
    if [ "$response" != "200" ]; then
        echo "Local service not responding, got HTTP $response"
        echo "Trying to get more details..."
        curl -v http://notlinktree_local:8080/api/config || echo "curl failed completely"
        exit 1
    fi
    
    # Test that we get valid JSON
    echo "Testing JSON response..."
    response=$(curl -s http://notlinktree_local:8080/api/config)
    echo "Response: $response"
    if ! echo "$response" | jq -e ".data.title" > /dev/null; then
        echo "Local service not returning valid JSON with expected structure"
        exit 1
    fi
    
    echo "Local service is working correctly"
}

# Test that we can read the test config
test_config_file() {
    echo "Testing config file access..."
    
    if [ ! -f "/test_config/config.yaml" ]; then
        echo "Test config file not found"
        exit 1
    fi
    
    if ! cat "/test_config/config.yaml" > /dev/null 2>&1; then
        echo "Cannot read test config file"
        exit 1
    fi
    
    echo "Config file is accessible"
}

# Run all tests
echo "Starting integration tests..."
echo "Current working directory: $(pwd)"
echo "Contents of /test_config:"
ls -la /test_config/ || echo "Cannot list /test_config"
echo "Contents of /test:"
ls -la /test/ || echo "Cannot list /test"
echo "Environment variables:"
env | grep -E "(NLT_|PATH)" || echo "No NLT_ variables found"

test_config_file
test_main_service
test_local_service

# Test password change functionality
echo ""
echo "=== Testing Password Change Functionality ==="
echo "Running password change tests on notlinktree_password_test service..."
if ! INTEGRATION_TEST=1 /test/password_test.sh; then
    echo "Password change tests failed"
    exit 1
fi

echo "All integration tests passed!" 