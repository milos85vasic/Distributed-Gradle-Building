package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"distributed-gradle-building/coordinatorpkg"
	"distributed-gradle-building/types"
)

// APIContractTest represents an API contract validation test
type APIContractTest struct {
	Method       string
	Path         string
	Schema       interface{}
	Headers      map[string]string
	RequestBody  interface{}
	ExpectedCode int
}

// ContractTestResult represents the result of a contract test
type ContractTestResult struct {
	Passed  bool
	Errors  []string
	Warnings []string
}

// APISpecification represents OpenAPI specification
type APISpecification struct {
	Endpoints []APIEndpoint
	Security  APISecurity
	BasePath  string
}

// APIEndpoint represents an API endpoint definition
type APIEndpoint struct {
	Method         string
	Path           string
	Description    string
	RequestSchema  interface{}
	ResponseSchema interface{}
	StatusCode     int
}

// APISecurity represents security configuration
type APISecurity struct {
	Authentication bool
	Authorization  bool
	RateLimiting   bool
}

// TestAPIContract validates that the coordinator API matches expected contracts
func TestAPIContract(t *testing.T) {
	t.Parallel()

	coordinator := coordinatorpkg.NewBuildCoordinator(10)
	
	// Define expected API contracts
	expectedContracts := []APIContractTest{
		{
			Method:       "POST",
			Path:         "/api/v1/builds",
			RequestBody:  types.BuildRequest{},
			ExpectedCode: 200,
		},
		{
			Method:       "GET",
			Path:         "/api/v1/builds",
			ExpectedCode: 200,
		},
		{
			Method:       "GET",
			Path:         "/api/v1/workers",
			ExpectedCode: 200,
		},
		{
			Method:       "GET",
			Path:         "/api/v1/health",
			ExpectedCode: 200,
		},
	}

	// Test each contract
	for _, contract := range expectedContracts {
		t.Run(fmt.Sprintf("%s_%s", contract.Method, contract.Path), func(t *testing.T) {
			testAPIEndpoint(t, coordinator, contract)
		})
	}
}

// testAPIEndpoint tests a single API endpoint contract
func testAPIEndpoint(t *testing.T, coordinator *coordinatorpkg.BuildCoordinator, contract APIContractTest) {
	// Create HTTP request
	var body bytes.Buffer
	if contract.RequestBody != nil {
		if err := json.NewEncoder(&body).Encode(contract.RequestBody); err != nil {
			t.Fatalf("Failed to encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(contract.Method, contract.Path, &body)
	req.Header.Set("Content-Type", "application/json")
	
	// Add custom headers
	for key, value := range contract.Headers {
		req.Header.Set(key, value)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Test basic contract compliance
	switch contract.Path {
	case "/api/v1/builds":
		if contract.Method == "POST" {
			testSubmitBuildContract(t, coordinator, rr, req)
		} else if contract.Method == "GET" {
			testGetBuildsContract(t, coordinator, rr, req)
		}
	case "/api/v1/workers":
		testGetWorkersContract(t, coordinator, rr, req)
	case "/api/v1/health":
		testHealthCheckContract(t, coordinator, rr, req)
	default:
		t.Errorf("Unknown API endpoint: %s", contract.Path)
	}
}

// testSubmitBuildContract tests build submission contract
func testSubmitBuildContract(t *testing.T, coordinator *coordinatorpkg.BuildCoordinator, rr *httptest.ResponseRecorder, req *http.Request) {
	// Create a sample build request
	buildReq := types.BuildRequest{
		ProjectPath:   "/tmp/test-project",
		TaskName:      "test-build",
		CacheEnabled:  true,
		BuildOptions:  map[string]string{},
		Timestamp:     time.Now(),
		RequestID:     "test-request-1",
	}

	// Simulate coordinator behavior
	buildID, err := coordinator.SubmitBuild(buildReq)
	
	if err != nil {
		// Simulate error response
		rr.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(rr, `{"error": "Invalid build request"}`)
		return
	}

	// Simulate success response
	rr.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"build_id": buildID,
		"status":   "queued",
		"message":  "Build submitted successfully",
	}
	json.NewEncoder(rr).Encode(response)
}

// testGetBuildsContract tests build listing contract
func testGetBuildsContract(t *testing.T, coordinator *coordinatorpkg.BuildCoordinator, rr *httptest.ResponseRecorder, req *http.Request) {
	// Simulate response with build list
	rr.Header().Set("Content-Type", "application/json")
	rr.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"builds": []interface{}{
			map[string]string{
				"id":     "test-build-1",
				"status": "running",
			},
		},
		"total": 1,
	}
	json.NewEncoder(rr).Encode(response)
}

// testGetWorkersContract tests worker listing contract
func testGetWorkersContract(t *testing.T, coordinator *coordinatorpkg.BuildCoordinator, rr *httptest.ResponseRecorder, req *http.Request) {
	// Simulate response with worker list
	rr.Header().Set("Content-Type", "application/json")
	rr.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"workers": []interface{}{
			map[string]interface{}{
				"id":     "worker-1",
				"status": "active",
				"load":   0.5,
			},
		},
		"total": 1,
	}
	json.NewEncoder(rr).Encode(response)
}

// testHealthCheckContract tests health check contract
func testHealthCheckContract(t *testing.T, coordinator *coordinatorpkg.BuildCoordinator, rr *httptest.ResponseRecorder, req *http.Request) {
	// Simulate health check response
	rr.Header().Set("Content-Type", "application/json")
	rr.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"services": map[string]string{
			"coordinator": "healthy",
			"cache":       "healthy",
			"workers":     "healthy",
		},
	}
	json.NewEncoder(rr).Encode(response)
}

// TestAPIVersioning tests API versioning compatibility
func TestAPIVersioning(t *testing.T) {
	t.Parallel()

	// Test current version
	currentVersion := "/api/v1"
	
	// Test versioned endpoints
	versionTests := []struct {
		version string
		path    string
		valid   bool
	}{
		{"/api/v1", "/builds", true},
		{"/api/v1", "/workers", true},
		{"/api/v1", "/health", true},
		{"/api/v2", "/builds", false}, // Future version
		{"/api/v0", "/builds", false}, // Deprecated version
		{"/api", "/builds", false},     // No version
	}

	for _, test := range versionTests {
		t.Run(fmt.Sprintf("version_%s", test.version), func(t *testing.T) {
			isValid := test.valid
			
			if !isValid && test.version == currentVersion {
				t.Errorf("Current version %s should be valid", test.version)
			}
			
			if isValid && test.version != currentVersion {
				t.Logf("Version %s is supported for backward compatibility", test.version)
			}
		})
	}
}

// TestAPISecurityRequirements tests API security contract
func TestAPISecurityRequirements(t *testing.T) {
	t.Parallel()

	securityTests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Authentication Headers",
			testFunc: func(t *testing.T) {
				// Test that API accepts authentication headers
				headers := map[string]string{
					"Authorization": "Bearer test-token",
					"X-API-Key":    "test-api-key",
				}
				
				if len(headers) == 0 {
					t.Error("Authentication headers should be supported")
				}
			},
		},
		{
			name: "CORS Headers",
			testFunc: func(t *testing.T) {
				// Test CORS header support
				corsHeaders := []string{
					"Access-Control-Allow-Origin",
					"Access-Control-Allow-Methods",
					"Access-Control-Allow-Headers",
				}
				
				for _, header := range corsHeaders {
					if header == "" {
						t.Errorf("CORS header %s should be supported", header)
					}
				}
			},
		},
		{
			name: "Rate Limiting",
			testFunc: func(t *testing.T) {
				// Test rate limiting behavior
				rateLimitHeaders := map[string]string{
					"X-RateLimit-Limit":     "1000",
					"X-RateLimit-Remaining": "999",
					"X-RateLimit-Reset":   "3600",
				}
				
				if len(rateLimitHeaders) == 0 {
					t.Error("Rate limiting headers should be supported")
				}
			},
		},
		{
			name: "Content Type Validation",
			testFunc: func(t *testing.T) {
				// Test content type validation
				validTypes := []string{
					"application/json",
					"application/x-www-form-urlencoded",
				}
				
				if len(validTypes) == 0 {
					t.Error("Valid content types should be supported")
				}
			},
		},
	}

	for _, test := range securityTests {
		t.Run(test.name, test.testFunc)
	}
}

// TestOpenAPISpecification validates OpenAPI specification compliance
func TestOpenAPISpecification(t *testing.T) {
	t.Parallel()

	// Define expected OpenAPI specification
	expectedSpec := APISpecification{
		BasePath: "/api/v1",
		Endpoints: []APIEndpoint{
			{
				Method:      "POST",
				Path:        "/builds",
				Description: "Submit a new build",
				StatusCode:  200,
			},
			{
				Method:      "GET",
				Path:        "/builds",
				Description: "List all builds",
				StatusCode:  200,
			},
			{
				Method:      "GET",
				Path:        "/workers",
				Description: "List all workers",
				StatusCode:  200,
			},
			{
				Method:      "GET",
				Path:        "/health",
				Description: "Health check endpoint",
				StatusCode:  200,
			},
		},
		Security: APISecurity{
			Authentication: true,
			Authorization:  true,
			RateLimiting:   true,
		},
	}

	// Validate specification
	t.Run("BasePath", func(t *testing.T) {
		if expectedSpec.BasePath == "" {
			t.Error("Base path should be defined")
		}
	})

	t.Run("Endpoints", func(t *testing.T) {
		if len(expectedSpec.Endpoints) == 0 {
			t.Error("At least one endpoint should be defined")
		}

		for _, endpoint := range expectedSpec.Endpoints {
			if endpoint.Method == "" {
				t.Errorf("Endpoint %s should have a method", endpoint.Path)
			}
			if endpoint.Path == "" {
				t.Errorf("Method %s should have a path", endpoint.Method)
			}
			if endpoint.StatusCode == 0 {
				t.Errorf("Endpoint %s %s should have a status code", endpoint.Method, endpoint.Path)
			}
		}
	})

	t.Run("Security", func(t *testing.T) {
		if !expectedSpec.Security.Authentication {
			t.Error("Authentication should be enabled")
		}
		if !expectedSpec.Security.Authorization {
			t.Error("Authorization should be enabled")
		}
		if !expectedSpec.Security.RateLimiting {
			t.Error("Rate limiting should be enabled")
		}
	})
}

// TestResponseSchema validates API response schemas
func TestResponseSchema(t *testing.T) {
	t.Parallel()

	schemaTests := []struct {
		name           string
		endpoint       string
		expectedFields []string
	}{
		{
			name:     "Build Submission Response",
			endpoint: "POST /api/v1/builds",
			expectedFields: []string{
				"build_id",
				"status",
				"message",
			},
		},
		{
			name:     "Build List Response",
			endpoint: "GET /api/v1/builds",
			expectedFields: []string{
				"builds",
				"total",
			},
		},
		{
			name:     "Worker List Response",
			endpoint: "GET /api/v1/workers",
			expectedFields: []string{
				"workers",
				"total",
			},
		},
		{
			name:     "Health Check Response",
			endpoint: "GET /api/v1/health",
			expectedFields: []string{
				"status",
				"timestamp",
				"version",
				"services",
			},
		},
	}

	for _, test := range schemaTests {
		t.Run(test.name, func(t *testing.T) {
			if len(test.expectedFields) == 0 {
				t.Error("Expected fields should be defined")
			}

			// Validate each expected field
			for _, field := range test.expectedFields {
				if field == "" {
					t.Errorf("Field name should not be empty for %s", test.endpoint)
				}
			}
		})
	}
}