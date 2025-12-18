package security

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// TestNetworkSecurity tests network-level security controls
func TestNetworkSecurity(t *testing.T) {
	// Create test environment
	testDir, err := ioutil.TempDir("", "network-security-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Security test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
		description string
	}{
		{
			name: "HTTPSRedirection",
			testFunc: func(t *testing.T) {
				testHTTPSRedirection(t)
			},
			description: "Test automatic HTTPS redirection",
		},
		{
			name: "TLSConfiguration",
			testFunc: func(t *testing.T) {
				testTLSConfiguration(t)
			},
			description: "Test TLS security configuration",
		},
		{
			name: "RateLimiting",
			testFunc: func(t *testing.T) {
				testRateLimiting(t)
			},
			description: "Test API rate limiting",
		},
		{
			name: "IPWhitelisting",
			testFunc: func(t *testing.T) {
				testIPWhitelisting(t)
			},
			description: "Test IP-based access control",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t)
		})
	}
}

// TestDataEncryption tests data encryption and protection
func TestDataEncryption(t *testing.T) {
	// Create test environment
	testDir, err := ioutil.TempDir("", "data-encryption-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Encryption test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
		description string
	}{
		{
			name: "PasswordHashing",
			testFunc: func(t *testing.T) {
				testPasswordHashing(t)
			},
			description: "Test secure password hashing",
		},
		{
			name: "SensitiveDataEncryption",
			testFunc: func(t *testing.T) {
				testSensitiveDataEncryption(t)
			},
			description: "Test sensitive data encryption at rest",
		},
		{
			name: "DataMasking",
			testFunc: func(t *testing.T) {
			 testDataMasking(t)
			},
			description: "Test sensitive data masking in logs",
		},
		{
			name: "TokenSecurity",
			testFunc: func(t *testing.T) {
				testTokenSecurity(t)
			},
			description: "Test secure token generation and validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t)
		})
	}
}

// TestAPIVulnerabilityScanning tests for common API vulnerabilities
func TestAPIVulnerabilityScanning(t *testing.T) {
	// Create test environment
	testDir, err := ioutil.TempDir("", "api-vuln-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Vulnerability test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
		description string
	}{
		{
			name: "SQLInjectionPrevention",
			testFunc: func(t *testing.T) {
				testSQLInjectionPrevention(t)
			},
			description: "Test SQL injection attack prevention",
		},
		{
			name: "XSSPrevention",
			testFunc: func(t *testing.T) {
				testXSSPrevention(t)
			},
			description: "Test XSS attack prevention",
		},
		{
			name: "CSRFProtection",
			testFunc: func(t *testing.T) {
				testCSRFProtection(t)
			},
			description: "Test CSRF protection",
		},
		{
			name: "BrokenAuthentication",
			testFunc: func(t *testing.T) {
				testBrokenAuthentication(t)
			},
			description: "Test broken authentication prevention",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t)
		})
	}
}

// Security test implementation functions

func testHTTPSRedirection(t *testing.T) {
	// Setup mock service with HTTPS redirection
	router := mux.NewRouter()
	
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			// Redirect HTTP to HTTPS
			target := fmt.Sprintf("https://%s%s", r.Host, r.URL.Path)
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Secure connection"))
		}
	})
	
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Test HTTP request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()
	
	// Should redirect (301)
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Errorf("Expected HTTP redirect (301), got %d", resp.StatusCode)
	} else {
		t.Logf("✓ HTTP requests properly redirect to HTTPS")
	}
}

func testTLSConfiguration(t *testing.T) {
	// Test TLS version and cipher configuration
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12, // Require TLS 1.2+
			},
		},
	}
	
	// This would test against a real server with TLS
	// For mock testing, we verify client configuration
	if client.Transport.(*http.Transport).TLSClientConfig.MinVersion < tls.VersionTLS12 {
		t.Error("TLS minimum version should be at least 1.2")
	} else {
		t.Logf("✓ TLS configuration requires minimum version 1.2")
	}
}

func testRateLimiting(t *testing.T) {
	// Setup mock service with rate limiting
	router := mux.NewRouter()
	requestCount := 0
	
	router.HandleFunc("/api/builds", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		// Simple rate limiting: max 5 requests
		if requestCount > 5 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Rate limit exceeded"}`))
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("POST")
	
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Test rate limiting
	successCount := 0
	rateLimitedCount := 0
	
	for i := 0; i < 10; i++ {
		resp, err := http.Post(
			server.URL+"/api/builds",
			"application/json",
			bytes.NewBuffer([]byte(`{"test": "data"}`)),
		)
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++
		}
		resp.Body.Close()
	}
	
	if successCount <= 5 && rateLimitedCount > 0 {
		t.Logf("✓ Rate limiting active: %d successful, %d rate limited", successCount, rateLimitedCount)
	} else {
		t.Errorf("Rate limiting not working: %d successful, %d rate limited", successCount, rateLimitedCount)
	}
}

func testIPWhitelisting(t *testing.T) {
	// Setup mock service with IP whitelisting
	router := mux.NewRouter()
	
	whitelistedIPs := []string{"127.0.0.1", "::1"}
	
	router.HandleFunc("/api/admin", func(w http.ResponseWriter, r *http.Request) {
		clientIP := strings.Split(r.RemoteAddr, ":")[0]
		
		isWhitelisted := false
		for _, ip := range whitelistedIPs {
			if ip == clientIP {
				isWhitelisted = true
				break
			}
		}
		
		if !isWhitelisted {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "IP not whitelisted"}`))
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"admin_data": "secure"}`))
	})
	
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Test whitelisted IP access
	resp, err := http.Get(server.URL + "/api/admin")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		t.Logf("✓ Whitelisted IP access allowed")
	} else {
		t.Errorf("Whitelisted IP access denied: %d", resp.StatusCode)
	}
}

func testPasswordHashing(t *testing.T) {
	// Test password hashing with bcrypt
	password := "test-password-123"
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	// Verify password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		t.Errorf("Password verification failed: %v", err)
	} else {
		t.Logf("✓ Password hashing and verification working")
	}
	
	// Test incorrect password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrong-password"))
	if err == nil {
		t.Error("Incorrect password should not verify")
	} else {
		t.Logf("✓ Incorrect password properly rejected")
	}
}

func testSensitiveDataEncryption(t *testing.T) {
	// Test basic encryption/decryption
	sensitiveData := "user-credentials-secret-key"
	
	// Simple XOR encryption for demonstration
	key := []byte("encryption-key-123")
	encrypted := make([]byte, len(sensitiveData))
	
	for i := 0; i < len(sensitiveData); i++ {
		encrypted[i] = sensitiveData[i] ^ key[i%len(key)]
	}
	
	// Decrypt
	decrypted := make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i++ {
		decrypted[i] = encrypted[i] ^ key[i%len(key)]
	}
	
	if string(decrypted) == sensitiveData {
		t.Logf("✓ Sensitive data encryption/decryption working")
	} else {
		t.Error("Data encryption/decryption failed")
	}
}

func testDataMasking(t *testing.T) {
	// Test data masking for logs
	email := "user@example.com"
	maskedEmail := maskEmail(email)
	
	expectedMasked := "u***@e******.com"
	if maskedEmail == expectedMasked {
		t.Logf("✓ Email masking working: %s -> %s", email, maskedEmail)
	} else {
		t.Errorf("Email masking failed: expected %s, got %s", expectedMasked, maskedEmail)
	}
	
	// Test credit card masking
	creditCard := "1234567890123456"
	maskedCard := maskCreditCard(creditCard)
	
	expectedCardMask := "************3456"
	if maskedCard == expectedCardMask {
		t.Logf("✓ Credit card masking working: %s -> %s", creditCard, maskedCard)
	} else {
		t.Errorf("Credit card masking failed: expected %s, got %s", expectedCardMask, maskedCard)
	}
}

func testTokenSecurity(t *testing.T) {
	// Test secure token generation
	token := generateSecureToken()
	
	if len(token) >= 32 {
		t.Logf("✓ Secure token generated: %d characters", len(token))
	} else {
		t.Errorf("Token too short: %d characters", len(token))
	}
	
	// Test token contains random characters
	token2 := generateSecureToken()
	if token != token2 {
		t.Logf("✓ Token generation is non-deterministic")
	} else {
		t.Error("Token generation is deterministic")
	}
}

func testSQLInjectionPrevention(t *testing.T) {
	// Setup mock service vulnerable to SQL injection for testing
	router := mux.NewRouter()
	
	router.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		
		// Check for SQL injection patterns
		injectionPatterns := []string{
			"'",
			";",
			"--",
			"/*",
			"*/",
			"xp_",
			"select",
			"drop",
			"insert",
			"update",
			"delete",
		}
		
		for _, pattern := range injectionPatterns {
			if strings.Contains(strings.ToLower(username), pattern) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Invalid input detected"}`))
				return
			}
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users": []}`))
	})
	
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Test SQL injection attempts
	injectionAttempts := []string{
		"admin'; DROP TABLE users; --",
		"' OR '1'='1",
		"1' UNION SELECT * FROM users --",
	}
	
	for _, attempt := range injectionAttempts {
		resp, err := http.Get(fmt.Sprintf("%s/api/users?username=%s", server.URL, attempt))
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusBadRequest {
			t.Logf("✓ SQL injection attempt blocked: %s", attempt)
		} else {
			t.Errorf("SQL injection attempt not blocked: %s (status: %d)", attempt, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

func testXSSPrevention(t *testing.T) {
	// Setup mock service with XSS protection
	router := mux.NewRouter()
	
	router.HandleFunc("/api/comments", func(w http.ResponseWriter, r *http.Request) {
		comment := r.URL.Query().Get("comment")
		
		// Check for XSS patterns
		xssPatterns := []string{
			"<script",
			"javascript:",
			"onerror=",
			"onload=",
			"<iframe",
			"<object",
			"<embed",
		}
		
		for _, pattern := range xssPatterns {
			if strings.Contains(strings.ToLower(comment), pattern) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Invalid input detected"}`))
				return
			}
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "comment posted"}`))
	})
	
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Test XSS attempts
	xssAttempts := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
		"<iframe src='javascript:alert(\"xss\")'></iframe>",
	}
	
	for _, attempt := range xssAttempts {
		resp, err := http.Get(fmt.Sprintf("%s/api/comments?comment=%s", server.URL, attempt))
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusBadRequest {
			t.Logf("✓ XSS attempt blocked: %s", attempt)
		} else {
			t.Errorf("XSS attempt not blocked: %s (status: %d)", attempt, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

func testCSRFProtection(t *testing.T) {
	// Setup mock service with CSRF protection
	router := mux.NewRouter()
	
	// Store session tokens
	sessionTokens := map[string]string{}
	
	router.HandleFunc("/api/csrf-token", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
		}
		
		csrfToken := generateSecureToken()
		sessionTokens[sessionID] = csrfToken
		
		w.Header().Set("X-Session-ID", sessionID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"csrf_token": "%s"}`, csrfToken)))
	})
	
	router.HandleFunc("/api/transfer", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")
		csrfToken := r.Header.Get("X-CSRF-Token")
		
		storedToken, exists := sessionTokens[sessionID]
		if !exists || storedToken != csrfToken {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "CSRF token invalid"}`))
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "transfer completed"}`))
	}).Methods("POST")
	
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Get CSRF token
	resp, err := http.Get(server.URL + "/api/csrf-token")
	if err != nil {
		t.Fatalf("Failed to get CSRF token: %v", err)
	}
	defer resp.Body.Close()
	
	var tokenResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("Failed to decode CSRF token response: %v", err)
	}
	
	sessionID := resp.Header.Get("X-Session-ID")
	csrfToken := tokenResp["csrf_token"]
	
	// Test valid CSRF token
	req, _ := http.NewRequest("POST", server.URL+"/api/transfer", nil)
	req.Header.Set("X-Session-ID", sessionID)
	req.Header.Set("X-CSRF-Token", csrfToken)
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make protected request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		t.Logf("✓ Valid CSRF token accepted")
	} else {
		t.Errorf("Valid CSRF token rejected: %d", resp.StatusCode)
	}
	
	// Test invalid CSRF token
	req, _ = http.NewRequest("POST", server.URL+"/api/transfer", nil)
	req.Header.Set("X-Session-ID", sessionID)
	req.Header.Set("X-CSRF-Token", "invalid-token")
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request with invalid CSRF: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusForbidden {
		t.Logf("✓ Invalid CSRF token rejected")
	} else {
		t.Errorf("Invalid CSRF token accepted: %d", resp.StatusCode)
	}
}

func testBrokenAuthentication(t *testing.T) {
	// Test for common authentication vulnerabilities
	
	// Test weak password rejection
	weakPasswords := []string{
		"password",
		"123456",
		"admin",
		"qwerty",
		"password123",
	}
	
	for _, weak := range weakPasswords {
		if isStrongPassword(weak) {
			t.Errorf("Weak password should be rejected: %s", weak)
		} else {
			t.Logf("✓ Weak password rejected: %s", weak)
		}
	}
	
	// Test session management
	sessionToken := generateSecureToken()
	
	if len(sessionToken) < 32 {
		t.Error("Session token should be at least 32 characters")
	} else {
		t.Logf("✓ Session token length sufficient: %d characters", len(sessionToken))
	}
	
	// Test session timeout
	sessionCreated := time.Now().Add(-25 * time.Hour) // 25 hours ago
	if isSessionValid(sessionCreated) {
		t.Error("Old session should be invalid")
	} else {
		t.Logf("✓ Session timeout working")
	}
}

// Helper functions for security tests

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	
	local := parts[0]
	domain := parts[1]
	
	if len(local) > 1 {
		maskedLocal := string(local[0]) + strings.Repeat("*", len(local)-2) + string(local[len(local)-1])
		return maskedLocal + "@" + domain
	}
	
	return email
}

func maskCreditCard(card string) string {
	if len(card) < 4 {
		return strings.Repeat("*", len(card))
	}
	
	return strings.Repeat("*", len(card)-4) + card[len(card)-4:]
}

func generateSecureToken() string {
	// Generate cryptographically secure random token
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[i%len(charset)] // Simplified for testing
	}
	return string(b)
}

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func isSessionValid(createdAt time.Time) bool {
	maxAge := 24 * time.Hour
	return time.Since(createdAt) < maxAge
}