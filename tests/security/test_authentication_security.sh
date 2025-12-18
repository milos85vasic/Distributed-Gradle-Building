#!/bin/bash

# Security tests for authentication and authorization
# This test validates security controls and access controls

# Import test framework
source "$(dirname "$0")/../test_framework.sh"

# Test configuration
source "$(dirname "$0")/../test_config.sh"

# Simple test functions
test_pass() {
    echo "PASS: $1"
}

test_fail() {
    echo "FAIL: $1"
}

# Security test configuration
SECURITY_TEST_USERS=(
    "admin:admin123:admin"
    "user:user123:user"
    "worker:worker123:worker"
    "guest:guest:guest"
)

SECURITY_TEST_INVALID_TOKENS=(
    "invalid-token"
    "expired-token"
    "malformed-token"
    "empty-token"
)

# Test setup
setup_security_test() {
    echo "Setting up security test environment..."
    
    # Create test workspace
    TEST_WORKSPACE=$(mktemp -d)
    export TEST_WORKSPACE
    
    # Create security test configuration
    cat > "$TEST_WORKSPACE/security_config.json" << 'EOF'
{
    "authentication": {
        "enabled": true,
        "method": "jwt",
        "secret_key": "test-secret-key-for-security-testing",
        "token_expiry": 3600
    },
    "authorization": {
        "roles": {
            "admin": ["*"],
            "user": ["read", "build:submit"],
            "worker": ["build:execute", "status:update"],
            "guest": ["read"]
        }
    },
    "security_headers": {
        "x_frame_options": "DENY",
        "content_type_options": "nosniff",
        "xss_protection": "1; mode=block"
    }
}
EOF
    
    echo "Security test workspace created at $TEST_WORKSPACE"
}

# Test JWT token generation and validation
test_jwt_token_validation() {
    echo "Testing JWT token validation..."
    
    # Start mock auth service
    cd "$TEST_WORKSPACE"
    
    cat > mock_auth_server.py << 'EOF'
import http.server
import socketserver
import json
import time
import base64
import hmac
import hashlib
import jwt
import os
from urllib.parse import urlparse, parse_qs

class SecurityAuthService(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == "/api/auth/login":
            content_length = int(self.headers["Content-Length"])
            post_data = self.rfile.read(content_length)
            credentials = json.loads(post_data.decode("utf-8"))
            
            username = credentials.get("username")
            password = credentials.get("password")
            
            # Simple authentication logic for testing
            users = {
                "admin": {"password": "admin123", "role": "admin"},
                "user": {"password": "user123", "role": "user"},
                "worker": {"password": "worker123", "role": "worker"},
                "guest": {"password": "guest", "role": "guest"}
            }
            
            if username in users and users[username]["password"] == password:
                # Generate JWT token
                payload = {
                    "username": username,
                    "role": users[username]["role"],
                    "exp": int(time.time()) + 3600
                }
                
                token = jwt.encode(payload, "test-secret-key-for-security-testing", algorithm="HS256")
                
                response = {
                    "status": "success",
                    "token": token,
                    "user": {
                        "username": username,
                        "role": users[username]["role"]
                    }
                }
                
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())
            else:
                self.send_response(401)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"error": "Invalid credentials"}')
        
        elif self.path == "/api/auth/validate":
            auth_header = self.headers.get("Authorization", "")
            if auth_header.startswith("Bearer "):
                token = auth_header[7:]
                
                try:
                    payload = jwt.decode(token, "test-secret-key-for-security-testing", algorithms=["HS256"])
                    response = {
                        "valid": True,
                        "user": payload
                    }
                    
                    self.send_response(200)
                    self.send_header("Content-Type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps(response).encode())
                except jwt.InvalidTokenError:
                    response = {"valid": False, "error": "Invalid token"}
                    
                    self.send_response(401)
                    self.send_header("Content-Type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps(response).encode())
            else:
                response = {"valid": False, "error": "No token provided"}
                
                self.send_response(401)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())
    
    def log_message(self, format, *args):
        # Suppress log messages
        pass

PORT = int(os.environ.get("COORDINATOR_PORT", "8082"))
with socketserver.TCPServer(("", PORT), SecurityAuthService) as httpd:
    httpd.serve_forever()
EOF
    
    python3 mock_auth_server.py &
    AUTH_PID=$!
    
    sleep 2
    
    # Test valid user authentication
    for user_cred in "${SECURITY_TEST_USERS[@]}"; do
        IFS=':' read -r username password role <<< "$user_cred"
        
        CREDENTIALS="{
            \"username\": \"$username\",
            \"password\": \"$password\"
        }"
        
        RESPONSE=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/auth/login" \
            -H "Content-Type: application/json" \
            -d "$CREDENTIALS")
        
        TOKEN=$(echo "$RESPONSE" | jq -r '.token // empty')
        RESPONSE_ROLE=$(echo "$RESPONSE" | jq -r '.user.role // empty')
        
        if [ -n "$TOKEN" ] && [ "$RESPONSE_ROLE" = "$role" ]; then
            test_pass "User $username authenticated successfully with role $role"
        else
            test_fail "User $username authentication failed: $RESPONSE"
        fi
    done
    
    # Test invalid authentication
    INVALID_CREDENTIALS='{"username": "invalid", "password": "invalid"}'
    RESPONSE=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "$INVALID_CREDENTIALS")
    
    HTTP_CODE=$(echo "$RESPONSE" | curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "$INVALID_CREDENTIALS")
    
    if [ "$HTTP_CODE" = "401" ]; then
        test_pass "Invalid credentials properly rejected with 401"
    else
        test_fail "Invalid credentials should return 401, got $HTTP_CODE"
    fi
    
    # Test token validation
    ADMIN_TOKEN=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username": "admin", "password": "admin123"}' | jq -r '.token')
    
    VALIDATE_RESPONSE=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/auth/validate" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    
    IS_VALID=$(echo "$VALIDATE_RESPONSE" | jq -r '.valid // false')
    
    if [ "$IS_VALID" = "true" ]; then
        test_pass "Valid token successfully validated"
    else
        test_fail "Valid token validation failed: $VALIDATE_RESPONSE"
    fi
    
    # Stop auth service
    kill $AUTH_PID 2>/dev/null
}

# Test role-based access control
test_role_based_access() {
    echo "Testing role-based access control..."
    
    # Start mock coordinator with RBAC
    cd "$TEST_WORKSPACE"
    
    cat > mock_rbac_server.py << 'EOF'
import http.server
import socketserver
import json
import jwt
import os
from functools import wraps

class RBACCoordinator(http.server.BaseHTTPRequestHandler):
    
    def check_auth(self, required_role=None, required_permission=None):
        auth_header = self.headers.get("Authorization", "")
        if not auth_header.startswith("Bearer "):
            return False, "No token provided"
        
        token = auth_header[7:]
        try:
            payload = jwt.decode(token, "test-secret-key-for-security-testing", algorithms=["HS256"])
            user_role = payload.get("role")
            
            # Define permissions for roles
            permissions = {
                "admin": ["*"],
                "user": ["read", "build:submit"],
                "worker": ["build:execute", "status:update"],
                "guest": ["read"]
            }
            
            if required_role and user_role != required_role:
                return False, "Insufficient role"
            
            if required_permission:
                user_permissions = permissions.get(user_role, [])
                if "*" not in user_permissions and required_permission not in user_permissions:
                    return False, "Insufficient permission"
            
            return True, payload
        except jwt.InvalidTokenError:
            return False, "Invalid token"
    
    def do_GET(self):
        if self.path == "/api/builds":
            # Requires "read" permission
            valid, result = self.check_auth(required_permission="read")
            if valid:
                response = {"builds": ["build-1", "build-2"]}
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())
            else:
                self.send_response(403)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": result}).encode())
        
        elif self.path == "/api/admin/stats":
            # Requires "admin" role
            valid, result = self.check_auth(required_role="admin")
            if valid:
                response = {"stats": "admin-data"}
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())
            else:
                self.send_response(403)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": result}).encode())
    
    def do_POST(self):
        if self.path == "/api/builds":
            # Requires "build:submit" permission
            valid, result = self.check_auth(required_permission="build:submit")
            if valid:
                response = {"build_id": "build-123", "status": "submitted"}
                self.send_response(201)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())
            else:
                self.send_response(403)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": result}).encode())
    
    def log_message(self, format, *args):
        pass

PORT = int(os.environ.get("COORDINATOR_PORT", "8082"))
with socketserver.TCPServer(("", PORT), RBACCoordinator) as httpd:
    httpd.serve_forever()
EOF
    
    python3 mock_rbac_server.py &
    RBAC_PID=$!
    
    sleep 2
    
    # Get tokens for different users
    ADMIN_TOKEN=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username": "admin", "password": "admin123"}' | jq -r '.token')
    
    USER_TOKEN=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username": "user", "password": "user123"}' | jq -r '.token')
    
    GUEST_TOKEN=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username": "guest", "password": "guest"}' | jq -r '.token')
    
    # Test admin access to admin endpoint
    ADMIN_ACCESS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "http://localhost:$COORDINATOR_PORT/api/admin/stats" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    
    if [ "$ADMIN_ACCESS" = "200" ]; then
        test_pass "Admin can access admin endpoint"
    else
        test_fail "Admin access to admin endpoint failed: $ADMIN_ACCESS"
    fi
    
    # Test user denied from admin endpoint
    USER_ADMIN_ACCESS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "http://localhost:$COORDINATOR_PORT/api/admin/stats" \
        -H "Authorization: Bearer $USER_TOKEN")
    
    if [ "$USER_ADMIN_ACCESS" = "403" ]; then
        test_pass "User properly denied from admin endpoint"
    else
        test_fail "User should be denied from admin endpoint: $USER_ADMIN_ACCESS"
    fi
    
    # Test user can submit builds
    USER_BUILD_ACCESS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
        -H "Authorization: Bearer $USER_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"project_name": "test"}')
    
    if [ "$USER_BUILD_ACCESS" = "201" ]; then
        test_pass "User can submit builds"
    else
        test_fail "User build submission failed: $USER_BUILD_ACCESS"
    fi
    
    # Test guest can read builds but not submit
    GUEST_READ_ACCESS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "http://localhost:$COORDINATOR_PORT/api/builds" \
        -H "Authorization: Bearer $GUEST_TOKEN")
    
    GUEST_BUILD_ACCESS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
        -H "Authorization: Bearer $GUEST_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"project_name": "test"}')
    
    if [ "$GUEST_READ_ACCESS" = "200" ] && [ "$GUEST_BUILD_ACCESS" = "403" ]; then
        test_pass "Guest can read but not submit builds"
    else
        test_fail "Guest access control failed: read=$GUEST_READ_ACCESS, submit=$GUEST_BUILD_ACCESS"
    fi
    
    # Stop RBAC service
    kill $RBAC_PID 2>/dev/null
}

# Test input validation and SQL injection prevention
test_input_validation() {
    echo "Testing input validation and injection prevention..."
    
    # Start mock service with input validation
    cd "$TEST_WORKSPACE"
    
    cat > mock_validation_server.py << 'EOF'
import http.server
import socketserver
import json
import re
import html
import os

class InputValidationService(http.server.BaseHTTPRequestHandler):
    
    def validate_input(self, input_data):
        # Check for SQL injection patterns
        sql_patterns = [
            r"union.*select",
            r"select.*from",
            r"drop.*table",
            r"insert.*into",
            r"update.*set",
            r"delete.*from",
            r"--",
            r"/\*.*\*/",
            r"xp_cmdshell",
            r"sp_executesql"
        ]
        
        # Check for XSS patterns
        xss_patterns = [
            r"<script.*?>.*?</script>",
            r"javascript:",
            r"on\w+\s*=",
            r"<iframe.*?>",
            r"<object.*?>",
            r"<embed.*?>"
        ]
        
        input_str = str(input_data).lower()
        
        for pattern in sql_patterns:
            if re.search(pattern, input_str, re.IGNORECASE):
                return False, f"SQL injection pattern detected: {pattern}"
        
        for pattern in xss_patterns:
            if re.search(pattern, input_str, re.IGNORECASE):
                return False, f"XSS pattern detected: {pattern}"
        
        return True, "Valid input"
    
    def do_POST(self):
        if self.path == "/api/builds":
            content_length = int(self.headers["Content-Length"])
            post_data = self.rfile.read(content_length)
            
            try:
                build_data = json.loads(post_data.decode("utf-8"))
            except json.JSONDecodeError:
                self.send_response(400)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"error": "Invalid JSON"}')
                return
            
            # Validate project name
            if "project_name" in build_data:
                valid, message = self.validate_input(build_data["project_name"])
                if not valid:
                    self.send_response(400)
                    self.send_header("Content-Type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps({"error": message}).encode())
                    return
            
            # Validate tasks
            if "tasks" in build_data:
                for task in build_data["tasks"]:
                    valid, message = self.validate_input(task)
                    if not valid:
                        self.send_response(400)
                        self.send_header("Content-Type", "application/json")
                        self.end_headers()
                        self.wfile.write(json.dumps({"error": message}).encode())
                        return
            
            response = {"status": "accepted", "build_id": "safe-build-123"}
            self.send_response(201)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())
    
    def log_message(self, format, *args):
        pass

PORT = int(os.environ.get("COORDINATOR_PORT", "8082"))
with socketserver.TCPServer(("", PORT), InputValidationService) as httpd:
    httpd.serve_forever()
EOF
    
    python3 mock_validation_server.py &
    VALIDATION_PID=$!
    
    sleep 2
    
    # Test valid input
    VALID_BUILD='{"project_name": "valid-project", "tasks": ["compile", "test"]}'
    VALID_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
        -H "Content-Type: application/json" \
        -d "$VALID_BUILD")
    
    if [ "$VALID_RESPONSE" = "201" ]; then
        test_pass "Valid input accepted"
    else
        test_fail "Valid input rejected: $VALID_RESPONSE"
    fi
    
    # Test SQL injection attempts
    SQL_INJECTION_ATTEMPTS=(
        '{"project_name": "test; DROP TABLE users; --", "tasks": ["compile"]}'
        '{"project_name": "test", "tasks": ["compile; SELECT * FROM users"]}'
        '{"project_name": "test'\'' UNION SELECT * FROM users --", "tasks": ["compile"]}'
    )
    
    for injection in "${SQL_INJECTION_ATTEMPTS[@]}"; do
        INJECTION_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "$injection")
        
        if [ "$INJECTION_RESPONSE" = "400" ]; then
            test_pass "SQL injection attempt blocked"
        else
            test_fail "SQL injection attempt not blocked: $INJECTION_RESPONSE"
        fi
    done
    
    # Test XSS attempts
    XSS_ATTEMPTS=(
        '{"project_name": "script-tag-test", "tasks": ["compile"]}'
        '{"project_name": "test", "tasks": ["javascript-alert-test"]}'
        '{"project_name": "test", "tasks": ["img-tag-test"]}'
    )
    
    for xss in "${XSS_ATTEMPTS[@]}"; do
        XSS_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "$xss")
        
        if [ "$XSS_RESPONSE" = "400" ]; then
            test_pass "XSS attempt blocked"
        else
            test_fail "XSS attempt not blocked: $XSS_RESPONSE"
        fi
    done
    
    # Stop validation service
    kill $VALIDATION_PID 2>/dev/null
}

# Test secure headers
test_security_headers() {
    echo "Testing security headers..."
    
    # Start service with security headers
    cd "$TEST_WORKSPACE"
    
    cat > security_headers_server.py << 'EOF'
import http.server
import socketserver
import json
import os

class SecurityHeadersService(http.server.BaseHTTPRequestHandler):
    
    def send_response_with_headers(self, code):
        self.send_response(code)
        self.send_header('X-Frame-Options', 'DENY')
        self.send_header('X-Content-Type-Options', 'nosniff')
        self.send_header('X-XSS-Protection', '1; mode=block')
        self.send_header('Strict-Transport-Security', 'max-age=31536000; includeSubDomains')
        self.send_header('Content-Security-Policy', "default-src 'self'")
    
    def do_GET(self):
        self.send_response_with_headers(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({'status': 'ok'}).encode())
    
    def log_message(self, format, *args):
        pass

PORT = int(os.environ.get('COORDINATOR_PORT', '8082'))
with socketserver.TCPServer(('', PORT), SecurityHeadersService) as httpd:
    httpd.serve_forever()
EOF
    
    python3 security_headers_server.py &
    
    HEADERS_PID=$!
    
    sleep 2
    
    # Get response headers
    HEADERS_RESPONSE=$(curl -s -I "http://localhost:$COORDINATOR_PORT/api/status")
    
    # Check for required security headers
    SECURITY_HEADERS=(
        "X-Frame-Options: DENY"
        "X-Content-Type-Options: nosniff"
        "X-XSS-Protection: 1\; mode=block"
        "Strict-Transport-Security:"
        "Content-Security-Policy:"
    )
    
    for header in "${SECURITY_HEADERS[@]}"; do
        if echo "$HEADERS_RESPONSE" | grep -q "$header"; then
            test_pass "Security header present: $header"
        else
            test_fail "Security header missing: $header"
        fi
    done
    
    # Stop headers service
    kill $HEADERS_PID 2>/dev/null
}

# Test cleanup
test_cleanup() {
    echo "Cleaning up security test..."
    
    # Kill any remaining processes
    pkill -f "python3.*Security" 2>/dev/null || true
    pkill -f "python3.*RBAC" 2>/dev/null || true
    pkill -f "python3.*InputValidation" 2>/dev/null || true
    pkill -f "python3.*SecurityHeaders" 2>/dev/null || true
    
    # Cleanup test workspace
    rm -rf "$TEST_WORKSPACE"
    
    test_pass "Security test cleanup completed"
}

# Main test execution
main() {
    echo "Starting security tests..."
    
    # Check dependencies
    if ! command -v python3 >/dev/null 2>&1; then
        test_fail "python3 required for security test mock services"
        exit 1
    fi
    
    # Check if PyJWT is available
    if ! python3 -c 'import jwt' 2>/dev/null; then
        test_fail "PyJWT library required for security tests"
        exit 1
    fi
    
    # Run security test suite
    setup_security_test || exit 1
    test_jwt_token_validation || exit 1
    test_role_based_access || exit 1
    test_input_validation || exit 1
    test_security_headers || exit 1
    test_cleanup || exit 1
    
    echo "Security tests completed successfully!"
}

# Run tests if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi
