#!/bin/sh
set -e

echo "Password change tests: Starting password change integration tests..."

# Container runtime will be detected when needed

# Function to detect container runtime
detect_container_runtime() {
    if command -v docker >/dev/null 2>&1 && docker ps >/dev/null 2>&1; then
        echo "docker"
    elif command -v podman >/dev/null 2>&1 && podman ps >/dev/null 2>&1; then
        echo "podman"
    else
        echo ""
    fi
}

# Test configuration - passwords must meet validation requirements
ORIGINAL_PASSWORD="testpass123"
NEW_PASSWORD="newtestpass456"

# Determine if we're running inside a container (integration test) or standalone
if [ -n "$CONTAINER_NAME" ]; then
    # Running inside integration test container
    CONTAINER_NAME="notlinktree_password_test"
    TEST_PORT="8080"
    HOST="notlinktree_password_test"
elif [ -n "$INTEGRATION_TEST" ]; then
    # Running from integration test container
    CONTAINER_NAME="notlinktree_password_test"
    TEST_PORT="8080"
    HOST="notlinktree_password_test"
else
    # Running standalone test
    CONTAINER_NAME="test-password"
    TEST_PORT="8083"
    HOST="localhost"
fi

# Function to test login with given password
test_login() {
    local password="$1"
    local expected_status="$2"
    local test_name="$3"
    
    echo "Testing $test_name with password: $password"
    
    response=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "{\"password\":\"$password\"}" \
        http://$HOST:$TEST_PORT/api/admin/login || echo "000")
    
    echo "Got response code: $response (expected: $expected_status)"
    
    if [ "$response" != "$expected_status" ]; then
        echo "FAIL: $test_name - Expected HTTP $expected_status, got $response"
        return 1
    fi
    
    echo "PASS: $test_name"
    return 0
}

# Function to change password using docker exec or API
change_password() {
    local new_password="$1"
    local container_name="$2"
    local current_password="$3"  # Add current password parameter
    
    echo "Changing password to: $new_password"
    
    # Detect container runtime
    CONTAINER_RUNTIME=$(detect_container_runtime)
    
    # Check if we're running from integration test container (no container runtime available)
    if [ -n "$INTEGRATION_TEST" ] && [ -z "$CONTAINER_RUNTIME" ]; then
        echo "Running from integration test container - testing password change via API"
        
        # Use current password if provided, otherwise fall back to original
        local login_password="${current_password:-$ORIGINAL_PASSWORD}"
        
        # First, we need to login to get a JWT token
        echo "Logging in to get authentication token..."
        login_response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "{\"password\":\"$login_password\"}" \
            http://$HOST:$TEST_PORT/api/admin/login)
        
        if ! echo "$login_response" | jq -e ".success" > /dev/null; then
            echo "FAIL: Login failed - cannot get authentication token"
            return 1
        fi
        
        # Extract the token
        token=$(echo "$login_response" | jq -r ".data.token")
        if [ "$token" = "null" ] || [ -z "$token" ]; then
            echo "FAIL: No authentication token received"
            return 1
        fi
        
        echo "Got authentication token, changing password via API..."
        
        # Change password via API
        password_response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "{\"password\":\"$new_password\"}" \
            http://$HOST:$TEST_PORT/api/admin/password)
        
        if ! echo "$password_response" | jq -e ".success" > /dev/null; then
            echo "FAIL: Password change via API failed"
            echo "Response: $password_response"
            return 1
        fi
        
        echo "PASS: Password change via API executed successfully"
        return 0
    fi
    
    # Use container runtime exec to change password
    if ! $CONTAINER_RUNTIME exec $container_name sh -c "cd /config && notlinktree -setadminpw '$new_password'"; then
        echo "FAIL: Password change command failed"
        return 1
    fi
    
    echo "PASS: Password change command executed successfully"
    return 0
}

# Function to wait for service to be ready
wait_for_service() {
    local max_attempts=30
    local attempt=1
    
    echo "Waiting for service to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f http://$HOST:$TEST_PORT/api/config > /dev/null 2>&1; then
            echo "Service is ready after $attempt attempts"
            return 0
        fi
        
        echo "Attempt $attempt/$max_attempts: Service not ready yet, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "FAIL: Service did not become ready after $max_attempts attempts"
    return 1
}

# Main test execution
echo "Starting password change tests..."

# Set initial password for testing
echo "Setting initial password for testing..."
CONTAINER_RUNTIME=$(detect_container_runtime)

# Check if we're running from integration test container (no container runtime available)
if [ -n "$INTEGRATION_TEST" ] && [ -z "$CONTAINER_RUNTIME" ]; then
    echo "Running from integration test container - setting initial password via API"
    
    # First, we need to login to get a JWT token
    echo "Logging in to get authentication token..."
    login_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"password\":\"$ORIGINAL_PASSWORD\"}" \
        http://$HOST:$TEST_PORT/api/admin/login)

    if ! echo "$login_response" | jq -e ".success" > /dev/null; then
        echo "FAIL: Initial login failed - cannot get authentication token"
        exit 1
    fi

    echo "Got authentication token for initial setup"
else
    # Use container runtime exec to set initial password
    if ! $CONTAINER_RUNTIME exec $CONTAINER_NAME sh -c "cd /config && notlinktree -setadminpw '$ORIGINAL_PASSWORD'"; then
        echo "Failed to set initial password"
        exit 1
    fi
fi

# Wait for service to be ready
if ! wait_for_service; then
    echo "Service readiness test failed"
    exit 1
fi

# Test 1: Verify original password works
echo ""
echo "=== Test 1: Verify original password works ==="
if ! test_login "$ORIGINAL_PASSWORD" "200" "original password login"; then
    echo "Original password test failed - cannot proceed with password change tests"
    exit 1
fi

# Test 2: Change password using docker exec
echo ""
echo "=== Test 2: Change password using docker exec ==="
if ! change_password "$NEW_PASSWORD" "$CONTAINER_NAME" "$ORIGINAL_PASSWORD"; then
    echo "Password change test failed"
    exit 1
fi

# Give the service a moment to reload config
echo "Waiting for config reload..."
sleep 3

# Test 3: Verify old password no longer works
echo ""
echo "=== Test 3: Verify old password no longer works ==="
if ! test_login "$ORIGINAL_PASSWORD" "401" "old password rejection"; then
    echo "Old password rejection test failed"
    exit 1
fi

# Test 4: Verify new password works
echo ""
echo "=== Test 4: Verify new password works ==="
if ! test_login "$NEW_PASSWORD" "200" "new password login"; then
    echo "New password test failed"
    exit 1
fi

# Test 5: Test password change with empty password (should fail)
echo ""
echo "=== Test 5: Test password change with empty password (should fail) ==="
if change_password "" "$CONTAINER_NAME" "$NEW_PASSWORD"; then
    echo "FAIL: Empty password change should have failed but succeeded"
    exit 1
fi
echo "PASS: Empty password change correctly rejected"

# Test 5b: Test password change with short password (should fail)
echo ""
echo "=== Test 5b: Test password change with short password (should fail) ==="
if change_password "abc123" "$CONTAINER_NAME" "$NEW_PASSWORD"; then
    echo "FAIL: Short password change should have failed but succeeded"
    exit 1
fi
echo "PASS: Short password change correctly rejected"

# Test 5c: Test password change with previously blocked common password (should succeed)
echo ""
echo "=== Test 5c: Test password change with previously blocked common password (should succeed) ==="
if ! change_password "password123" "$CONTAINER_NAME" "$NEW_PASSWORD"; then
    echo "FAIL: Previously blocked common password change should have succeeded but failed"
    exit 1
fi
echo "PASS: Previously blocked common password change now allowed"

# Give the service a moment to reload config
echo "Waiting for config reload..."
sleep 3

# Test 5d: Verify the previously blocked common password works
echo ""
echo "=== Test 5d: Verify the previously blocked common password works ==="
if ! test_login "password123" "200" "previously blocked common password login"; then
    echo "Previously blocked common password login test failed"
    exit 1
fi

# Test 6: Test password change with special characters
echo ""
echo "=== Test 6: Test password change with special characters ==="
SPECIAL_PASSWORD="test@pass#123\$%^&*()"
if ! change_password "$SPECIAL_PASSWORD" "$CONTAINER_NAME" "password123"; then
    echo "Special character password change test failed"
    exit 1
fi

# Give the service a moment to reload config
echo "Waiting for config reload..."
sleep 3

# Test 7: Verify special character password works
echo ""
echo "=== Test 7: Verify special character password works ==="
if ! test_login "$SPECIAL_PASSWORD" "200" "special character password login"; then
    echo "Special character password test failed"
    exit 1
fi

# Test 8: Test password change back to original
echo ""
echo "=== Test 8: Test password change back to original ==="
if ! change_password "$ORIGINAL_PASSWORD" "$CONTAINER_NAME" "$SPECIAL_PASSWORD"; then
    echo "Password change back to original failed"
    exit 1
fi

# Give the service a moment to reload config
echo "Waiting for config reload..."
sleep 3

# Test 9: Verify we can login with original password again
echo ""
echo "=== Test 9: Verify we can login with original password again ==="
if ! test_login "$ORIGINAL_PASSWORD" "200" "original password login after change back"; then
    echo "Original password login after change back failed"
    exit 1
fi

echo ""
echo "All password change tests passed! ✅"
echo "Summary:"
echo "- Original password login: PASS"
echo "- Password change via container exec: PASS"
echo "- Old password rejection: PASS"
echo "- New password login: PASS"
echo "- Empty password rejection: PASS"
echo "- Short password rejection: PASS"
echo "- Previously blocked common password now allowed: PASS"
echo "- Previously blocked common password login: PASS"
echo "- Special character password: PASS"
echo "- Password change back to original: PASS"
echo "- Final login verification: PASS"
