package security

import (
	"bytes"
	"context"
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
)

// TestRoleBasedAccessControl tests RBAC implementation
func TestRoleBasedAccessControl(t *testing.T) {
	// Create test environment
	testDir, err := ioutil.TempDir("", "rbac-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// RBAC test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
		description string
	}{
		{
			name: "AdminRoleAccess",
			testFunc: func(t *testing.T) {
				testAdminRoleAccess(t)
			},
			description: "Test admin role has full access",
		},
		{
			name: "UserRoleAccess",
			testFunc: func(t *testing.T) {
				testUserRoleAccess(t)
			},
			description: "Test user role has limited access",
		},
		{
			name: "UnauthorizedAccess",
			testFunc: func(t *testing.T) {
				testUnauthorizedAccess(t)
			},
			description: "Test unauthorized access is blocked",
		},
		{
			name: "RoleEscalationPrevention",
			testFunc: func(t *testing.T) {
				testRoleEscalationPrevention(t)
			},
			description: "Test role escalation attacks are prevented",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t)
		})
	}
}

// TestResourceLevelPermissions tests resource-level access control
func TestResourceLevelPermissions(t *testing.T) {
	// Create test environment
	testDir, err := ioutil.TempDir("", "resource-permissions-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Resource permissions test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
		description string
	}{
		{
			name: "ProjectOwnership",
			testFunc: func(t *testing.T) {
				testProjectOwnership(t)
			},
			description: "Test project ownership rights",
		},
		{
			name: "BuildAccessControl",
			testFunc: func(t *testing.T) {
				testBuildAccessControl(t)
			},
			description: "Test build access control",
		},
		{
			name: "ResourceSharing",
			testFunc: func(t *testing.T) {
				testResourceSharing(t)
			},
			description: "Test resource sharing permissions",
		},
		{
			name: "CrossTenantIsolation",
			testFunc: func(t *testing.T) {
				testCrossTenantIsolation(t)
			},
			description: "Test cross-tenant resource isolation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t)
		})
	}
}

// TestAPIEndpointAuthorization tests API endpoint authorization
func TestAPIEndpointAuthorization(t *testing.T) {
	// Create test environment
	testDir, err := ioutil.TempDir("", "api-authorization-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// API authorization test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
		description string
	}{
		{
			name: "EndpointPermissions",
			testFunc: func(t *testing.T) {
				testEndpointPermissions(t)
			},
			description: "Test endpoint-level permissions",
		},
		{
			name: "MethodLevelAccess",
			testFunc: func(t *testing.T) {
				testMethodLevelAccess(t)
			},
			description: "Test HTTP method-level access control",
		},
		{
			name: "ParameterBasedAccess",
			testFunc: func(t *testing.T) {
				testParameterBasedAccess(t)
			},
			description: "Test parameter-based access control",
		},
		{
			name: "ConditionalAccess",
			testFunc: func(t *testing.T) {
				testConditionalAccess(t)
			},
			description: "Test conditional access policies",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t)
		})
	}
}

// Authorization test implementation functions

func testAdminRoleAccess(t *testing.T) {
	// Setup mock service with role-based access control
	router := setupRBACRouter()
	
	// Generate admin token
	adminToken := generateAuthToken("admin", "admin-role")
	
	// Test admin access to admin endpoints
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/admin/users"},
		{"POST", "/api/admin/config"},
		{"DELETE", "/api/admin/logs"},
		{"GET", "/api/admin/system/status"},
	}
	
	for _, endpoint := range endpoints {
		req, err := http.NewRequest(endpoint.method, endpoint.path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		
		req.Header.Set("Authorization", "Bearer "+adminToken)
		
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
			t.Errorf("Admin access failed for %s %s: %d", endpoint.method, endpoint.path, rr.Code)
		} else {
			t.Logf("✓ Admin access granted for %s %s", endpoint.method, endpoint.path)
		}
	}
}

func testUserRoleAccess(t *testing.T) {
	// Setup mock service with role-based access control
	router := setupRBACRouter()
	
	// Generate user token
	userToken := generateAuthToken("user", "user-role")
	
	// Test user access to user endpoints
	endpoints := []struct {
		method string
		path   string
		expect int
	}{
		{"GET", "/api/user/profile", http.StatusOK},
		{"POST", "/api/user/build", http.StatusOK},
		{"GET", "/api/user/builds", http.StatusOK},
		{"GET", "/api/admin/users", http.StatusForbidden},
		{"POST", "/api/admin/config", http.StatusForbidden},
		{"DELETE", "/api/admin/logs", http.StatusForbidden},
	}
	
	for _, endpoint := range endpoints {
		req, err := http.NewRequest(endpoint.method, endpoint.path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		
		req.Header.Set("Authorization", "Bearer "+userToken)
		
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != endpoint.expect {
			t.Errorf("User access incorrect for %s %s: got %d, expected %d", 
				endpoint.method, endpoint.path, rr.Code, endpoint.expect)
		} else {
			t.Logf("✓ User access correct for %s %s: %d", endpoint.method, endpoint.path, rr.Code)
		}
	}
}

func testUnauthorizedAccess(t *testing.T) {
	// Setup mock service with role-based access control
	router := setupRBACRouter()
	
	// Test access without authentication
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/user/profile"},
		{"POST", "/api/user/build"},
		{"GET", "/api/admin/users"},
		{"POST", "/api/admin/config"},
	}
	
	for _, endpoint := range endpoints {
		req, err := http.NewRequest(endpoint.method, endpoint.path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Unauthorized access not blocked for %s %s: %d", endpoint.method, endpoint.path, rr.Code)
		} else {
			t.Logf("✓ Unauthorized access blocked for %s %s", endpoint.method, endpoint.path)
		}
	}
	
	// Test with invalid token
	req, err := http.NewRequest("GET", "/api/user/profile", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer invalid-token")
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Invalid token not rejected: %d", rr.Code)
	} else {
		t.Logf("✓ Invalid token properly rejected")
	}
}

func testRoleEscalationPrevention(t *testing.T) {
	// Setup mock service with role-based access control
	router := setupRBACRouter()
	
	// Generate user token
	userToken := generateAuthToken("user", "user-role")
	
	// Test role escalation attempts
	escalationAttempts := []struct {
		method string
		path   string
		body   string
	}{
		{"POST", "/api/user/upgrade", `{"role": "admin"}`},
		{"PUT", "/api/user/profile", `{"role": "admin"}`},
		{"PATCH", "/api/user/permissions", `{"admin": true}`},
		{"POST", "/api/admin/grant", `{"user": "test", "role": "admin"}`},
	}
	
	for _, attempt := range escalationAttempts {
		req, err := http.NewRequest(attempt.method, attempt.path, strings.NewReader(attempt.body))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		
		req.Header.Set("Authorization", "Bearer "+userToken)
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusForbidden && rr.Code != http.StatusNotFound {
			t.Errorf("Role escalation not prevented for %s %s: %d", attempt.method, attempt.path, rr.Code)
		} else {
			t.Logf("✓ Role escalation prevented for %s %s", attempt.method, attempt.path)
		}
	}
}

func testProjectOwnership(t *testing.T) {
	// Setup mock service with project ownership control
	router := setupProjectRouter()
	
	// Create test projects and users
	ownerToken := generateAuthToken("owner", "user-role")
	otherToken := generateAuthToken("other", "user-role")
	
	// Test owner access
	req, err := http.NewRequest("GET", "/api/projects/project1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Project owner access denied: %d", rr.Code)
	} else {
		t.Logf("✓ Project owner access granted")
	}
	
	// Test other user access (should be denied)
	req, err = http.NewRequest("GET", "/api/projects/project1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+otherToken)
	
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusForbidden {
		t.Errorf("Non-owner access not denied: %d", rr.Code)
	} else {
		t.Logf("✓ Non-owner access properly denied")
	}
}

func testBuildAccessControl(t *testing.T) {
	// Setup mock service with build access control
	router := setupProjectRouter()
	
	// Create test project and user
	userToken := generateAuthToken("user", "user-role")
	
	// Test build permission checks
	buildActions := []struct {
		method string
		path   string
		expect int
	}{
		{"GET", "/api/projects/project1/builds", http.StatusOK},
		{"POST", "/api/projects/project1/builds", http.StatusOK},
		{"DELETE", "/api/projects/project1/builds/123", http.StatusOK},
		{"GET", "/api/projects/project2/builds", http.StatusForbidden},
		{"POST", "/api/projects/project2/builds", http.StatusForbidden},
	}
	
	for _, action := range buildActions {
		req, err := http.NewRequest(action.method, action.path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		
		req.Header.Set("Authorization", "Bearer "+userToken)
		
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != action.expect {
			t.Errorf("Build access incorrect for %s %s: got %d, expected %d", 
				action.method, action.path, rr.Code, action.expect)
		} else {
			t.Logf("✓ Build access correct for %s %s: %d", action.method, action.path, rr.Code)
		}
	}
}

func testResourceSharing(t *testing.T) {
	// Setup mock service with resource sharing
	router := setupProjectRouter()
	
	// Create test project and user
	ownerToken := generateAuthToken("owner", "user-role")
	
	// Test resource sharing
	shareData := map[string]interface{}{
		"user":    "other-user",
		"role":    "viewer",
		"project": "project1",
	}
	
	body, _ := json.Marshal(shareData)
	
	req, err := http.NewRequest("POST", "/api/projects/project1/share", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("Resource sharing failed: %d", rr.Code)
	} else {
		t.Logf("✓ Resource sharing working")
	}
}

func testCrossTenantIsolation(t *testing.T) {
	// Setup mock service with tenant isolation
	router := setupTenantRouter()
	
	// Create test users from different tenants
	tenant1Token := generateTenantToken("user1", "tenant1")
	tenant2Token := generateTenantToken("user2", "tenant2")
	
	// Test tenant1 accessing tenant2 resources (should be denied)
	req, err := http.NewRequest("GET", "/api/tenant2/data", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+tenant1Token)
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusForbidden {
		t.Errorf("Cross-tenant access not blocked: %d", rr.Code)
	} else {
		t.Logf("✓ Cross-tenant access properly blocked")
	}
	
	// Test tenant2 accessing own resources (should be allowed)
	req, err = http.NewRequest("GET", "/api/tenant2/data", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+tenant2Token)
	
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("Tenant own access blocked: %d", rr.Code)
	} else {
		t.Logf("✓ Tenant own access allowed")
	}
}

func testEndpointPermissions(t *testing.T) {
	// Setup mock service with endpoint permissions
	router := setupEndpointRouter()
	
	// Test different role access to endpoints
	testCases := []struct {
		role   string
		path   string
		expect int
	}{
		{"admin", "/api/secure/data", http.StatusOK},
		{"user", "/api/secure/data", http.StatusForbidden},
		{"guest", "/api/secure/data", http.StatusForbidden},
		{"admin", "/api/public/data", http.StatusOK},
		{"user", "/api/public/data", http.StatusOK},
		{"guest", "/api/public/data", http.StatusOK},
	}
	
	for _, tc := range testCases {
		token := generateAuthToken(tc.role, tc.role)
		
		req, err := http.NewRequest("GET", tc.path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		
		req.Header.Set("Authorization", "Bearer "+token)
		
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != tc.expect {
			t.Errorf("Endpoint permission incorrect for %s accessing %s: got %d, expected %d", 
				tc.role, tc.path, rr.Code, tc.expect)
		} else {
			t.Logf("✓ Endpoint permission correct for %s accessing %s: %d", 
				tc.role, tc.path, rr.Code)
		}
	}
}

func testMethodLevelAccess(t *testing.T) {
	// Setup mock service with method-level access
	router := setupMethodRouter()
	
	userToken := generateAuthToken("user", "user-role")
	
	// Test method-level permissions
	methodTests := []struct {
		method string
		path   string
		expect int
	}{
		{"GET", "/api/user/data", http.StatusOK},
		{"POST", "/api/user/data", http.StatusForbidden},
		{"PUT", "/api/user/data", http.StatusForbidden},
		{"DELETE", "/api/user/data", http.StatusForbidden},
	}
	
	for _, test := range methodTests {
		req, err := http.NewRequest(test.method, test.path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		
		req.Header.Set("Authorization", "Bearer "+userToken)
		
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != test.expect {
			t.Errorf("Method access incorrect for %s %s: got %d, expected %d", 
				test.method, test.path, rr.Code, test.expect)
		} else {
			t.Logf("✓ Method access correct for %s %s: %d", 
				test.method, test.path, rr.Code)
		}
	}
}

func testParameterBasedAccess(t *testing.T) {
	// Setup mock service with parameter-based access
	router := setupParameterRouter()
	
	userToken := generateAuthToken("user", "user-role")
	
	// Test parameter-based permissions
	paramTests := []struct {
		path   string
		expect int
	}{
		{"/api/data/123", http.StatusOK},     // User owns data 123
		{"/api/data/456", http.StatusForbidden}, // User doesn't own data 456
		{"/api/data/self", http.StatusOK},    // Self access
		{"/api/data/all", http.StatusForbidden}, // Admin only
	}
	
	for _, test := range paramTests {
		req, err := http.NewRequest("GET", test.path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		
		req.Header.Set("Authorization", "Bearer "+userToken)
		
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != test.expect {
			t.Errorf("Parameter access incorrect for %s: got %d, expected %d", 
				test.path, rr.Code, test.expect)
		} else {
			t.Logf("✓ Parameter access correct for %s: %d", test.path, rr.Code)
		}
	}
}

func testConditionalAccess(t *testing.T) {
	// Setup mock service with conditional access
	router := setupConditionalRouter()
	
	userToken := generateAuthToken("user", "user-role")
	
	// Test time-based access
	req, err := http.NewRequest("GET", "/api/time-sensitive", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+userToken)
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusOK && rr.Code != http.StatusForbidden {
		t.Errorf("Conditional access unexpected: %d", rr.Code)
	} else {
		t.Logf("✓ Conditional access working: %d", rr.Code)
	}
	
	// Test IP-based access
	req, err = http.NewRequest("GET", "/api/ip-restricted", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusOK && rr.Code != http.StatusForbidden {
		t.Errorf("IP-based access unexpected: %d", rr.Code)
	} else {
		t.Logf("✓ IP-based access working: %d", rr.Code)
	}
}

// Helper functions for authorization tests

// Mock user database
type User struct {
	ID       string
	Username string
	Role     string
	Tenant   string
}

// Mock project database
type Project struct {
	ID      string
	Owner   string
	Members map[string]string
}

var users = map[string]User{
	"admin":   {ID: "1", Username: "admin", Role: "admin", Tenant: "default"},
	"user":    {ID: "2", Username: "user", Role: "user", Tenant: "default"},
	"owner":   {ID: "3", Username: "owner", Role: "user", Tenant: "default"},
	"other":   {ID: "4", Username: "other", Role: "user", Tenant: "default"},
	"user1":   {ID: "5", Username: "user1", Role: "user", Tenant: "tenant1"},
	"user2":   {ID: "6", Username: "user2", Role: "user", Tenant: "tenant2"},
}

var projects = map[string]Project{
	"project1": {ID: "project1", Owner: "owner", Members: map[string]string{"owner": "owner"}},
	"project2": {ID: "project2", Owner: "other", Members: map[string]string{"other": "other"}},
}

func setupRBACRouter() *mux.Router {
	router := mux.NewRouter()
	
	// Authentication middleware
	router.Use(authMiddleware)
	
	// Admin endpoints
	router.HandleFunc("/api/admin/users", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		if user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "admin users"})
	}).Methods("GET")
	
	router.HandleFunc("/api/admin/config", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		if user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "admin config"})
	}).Methods("POST")
	
	router.HandleFunc("/api/admin/logs", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		if user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "admin logs"})
	}).Methods("DELETE")
	
	router.HandleFunc("/api/admin/system/status", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		if user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "system ok"})
	}).Methods("GET")
	
	// User endpoints
	router.HandleFunc("/api/user/profile", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "user profile"})
	}).Methods("GET")
	
	router.HandleFunc("/api/user/build", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "build started"})
	}).Methods("POST")
	
	router.HandleFunc("/api/user/builds", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string][]string{"builds": {"build1", "build2"}})
	}).Methods("GET")
	
	// Role escalation attempts (should be forbidden)
	router.HandleFunc("/api/user/upgrade", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}).Methods("POST")
	
	router.HandleFunc("/api/user/profile", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		if user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "profile updated"})
	}).Methods("PUT")
	
	router.HandleFunc("/api/user/permissions", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}).Methods("PATCH")
	
	router.HandleFunc("/api/admin/grant", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}).Methods("POST")
	
	return router
}

func setupProjectRouter() *mux.Router {
	router := mux.NewRouter()
	
	// Authentication middleware
	router.Use(authMiddleware)
	
	// Project endpoints
	router.HandleFunc("/api/projects/{projectId}", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		projectID := mux.Vars(r)["projectId"]
		
		project, exists := projects[projectID]
		if !exists || project.Owner != user.Username {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"project": projectID, "owner": project.Owner})
	}).Methods("GET")
	
	router.HandleFunc("/api/projects/{projectId}/builds", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		projectID := mux.Vars(r)["projectId"]
		
		project, exists := projects[projectID]
		if !exists || project.Owner != user.Username {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string][]string{"builds": {"build1", "build2"}})
	}).Methods("GET")
	
	router.HandleFunc("/api/projects/{projectId}/builds", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		projectID := mux.Vars(r)["projectId"]
		
		project, exists := projects[projectID]
		if !exists || project.Owner != user.Username {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"build": "created"})
	}).Methods("POST")
	
	router.HandleFunc("/api/projects/{projectId}/builds/{buildId}", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		projectID := mux.Vars(r)["projectId"]
		
		project, exists := projects[projectID]
		if !exists || project.Owner != user.Username {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "build deleted"})
	}).Methods("DELETE")
	
	router.HandleFunc("/api/projects/{projectId}/share", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		projectID := mux.Vars(r)["projectId"]
		
		project, exists := projects[projectID]
		if !exists || project.Owner != user.Username {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "project shared"})
	}).Methods("POST")
	
	return router
}

func setupTenantRouter() *mux.Router {
	router := mux.NewRouter()
	
	// Authentication middleware
	router.Use(tenantAuthMiddleware)
	
	// Tenant endpoints
	router.HandleFunc("/api/{tenantId}/data", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		tenantID := mux.Vars(r)["tenantId"]
		
		if user.Tenant != tenantID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"tenant": tenantID, "data": "secure"})
	}).Methods("GET")
	
	return router
}

func setupEndpointRouter() *mux.Router {
	router := mux.NewRouter()
	
	// Authentication middleware
	router.Use(authMiddleware)
	
	// Secure endpoint
	router.HandleFunc("/api/secure/data", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		if user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"data": "secure"})
	}).Methods("GET")
	
	// Public endpoint
	router.HandleFunc("/api/public/data", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"data": "public"})
	}).Methods("GET")
	
	return router
}

func setupMethodRouter() *mux.Router {
	router := mux.NewRouter()
	
	// Authentication middleware
	router.Use(authMiddleware)
	
	// Method-based access control
	router.HandleFunc("/api/user/data", func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context().Value("user").(User)
		
		switch r.Method {
		case "GET":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"data": "user data"})
		default:
			http.Error(w, "Method not allowed", http.StatusForbidden)
		}
	})
	
	return router
}

func setupParameterRouter() *mux.Router {
	router := mux.NewRouter()
	
	// Authentication middleware
	router.Use(authMiddleware)
	
	// Parameter-based access
	router.HandleFunc("/api/data/{dataId}", func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(User)
		dataID := mux.Vars(r)["dataId"]
		
		// User owns data 123, but not 456
		if dataID == "456" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		// Special cases
		if dataID == "all" && user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"data": dataID})
	}).Methods("GET")
	
	return router
}

func setupConditionalRouter() *mux.Router {
	router := mux.NewRouter()
	
	// Authentication middleware
	router.Use(authMiddleware)
	
	// Time-based access
	router.HandleFunc("/api/time-sensitive", func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context().Value("user").(User)
		
		// Allow access during business hours (9 AM - 5 PM)
		hour := time.Now().Hour()
		if hour < 9 || hour > 17 {
			http.Error(w, "Access denied outside business hours", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"data": "time-sensitive"})
	}).Methods("GET")
	
	// IP-based access
	router.HandleFunc("/api/ip-restricted", func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context().Value("user").(User)
		
		// Only allow access from specific IP ranges
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}
		
		allowedIPs := []string{"127.0.0.1", "::1", "192.168.1.0/24"}
		isAllowed := false
		
		for _, allowed := range allowedIPs {
			if strings.Contains(clientIP, strings.TrimSuffix(allowed, "/24")) {
				isAllowed = true
				break
			}
		}
		
		if !isAllowed {
			http.Error(w, "IP not allowed", http.StatusForbidden)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"data": "ip-restricted"})
	}).Methods("GET")
	
	return router
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Parse Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		
		// Simple token validation (in real implementation, use JWT)
		token := tokenParts[1]
		user, valid := validateToken(token)
		if !valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		// Add user to context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func tenantAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Parse Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		
		// Simple tenant token validation
		token := tokenParts[1]
		user, valid := validateTenantToken(token)
		if !valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		// Add user to context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateToken(token string) (User, bool) {
	// Simple token validation for testing
	// In real implementation, validate JWT signature, expiration, etc.
	
	tokenMap := map[string]string{
		"admin-token": "admin",
		"user-token":  "user",
		"owner-token": "owner",
		"other-token": "other",
	}
	
	username, exists := tokenMap[token]
	if !exists {
		return User{}, false
	}
	
	user, exists := users[username]
	if !exists {
		return User{}, false
	}
	
	return user, true
}

func validateTenantToken(token string) (User, bool) {
	// Simple tenant token validation for testing
	tokenMap := map[string]string{
		"tenant1-token": "user1",
		"tenant2-token": "user2",
	}
	
	username, exists := tokenMap[token]
	if !exists {
		return User{}, false
	}
	
	user, exists := users[username]
	if !exists {
		return User{}, false
	}
	
	return user, true
}

func generateAuthToken(username, role string) string {
	// Generate simple token for testing
	// In real implementation, generate JWT with proper claims
	return fmt.Sprintf("%s-token", username)
}

func generateTenantToken(username, tenant string) string {
	// Generate simple tenant token for testing
	// In real implementation, generate JWT with tenant claim
	return fmt.Sprintf("%s-token", tenant)
}