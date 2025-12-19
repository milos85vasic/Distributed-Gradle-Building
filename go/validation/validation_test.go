package validation

import (
	"strings"
	"testing"
	"time"

	"distributed-gradle-building/types"
)

func TestValidateBuildRequest(t *testing.T) {
	validRequest := &types.BuildRequest{
		ProjectPath: "/projects/myapp",
		TaskName:    "build",
		RequestID:   "req-123",
		BuildOptions: map[string]string{
			"clean": "true",
		},
	}

	// Test valid request
	if err := ValidateBuildRequest(validRequest); err != nil {
		t.Errorf("Expected valid request, got error: %v", err)
	}

	// Test missing project path
	invalidRequest := &types.BuildRequest{
		TaskName:  "build",
		RequestID: "req-123",
	}
	if err := ValidateBuildRequest(invalidRequest); err == nil {
		t.Error("Expected error for missing project path")
	}

	// Test missing task name
	invalidRequest = &types.BuildRequest{
		ProjectPath: "/projects/myapp",
		RequestID:   "req-123",
	}
	if err := ValidateBuildRequest(invalidRequest); err == nil {
		t.Error("Expected error for missing task name")
	}

	// Test missing request ID
	invalidRequest = &types.BuildRequest{
		ProjectPath: "/projects/myapp",
		TaskName:    "build",
	}
	if err := ValidateBuildRequest(invalidRequest); err == nil {
		t.Error("Expected error for missing request ID")
	}
}

func TestValidateProjectPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string // empty string means no error expected
	}{
		{"Valid absolute path", "/projects/myapp", ""},
		{"Valid nested path", "/home/user/projects/build.gradle", ""},
		{"Empty path", "", "project path cannot be empty"},
		{"Relative path", "projects/myapp", "project path must be absolute"},
		{"Path traversal", "/projects/../etc/passwd", "path traversal"},
		{"Path with dangerous char <", "/projects/<script>", "dangerous character"},
		{"Path with dangerous char >", "/projects/>", "dangerous character"},
		{"Path with dangerous char |", "/projects/|", "dangerous character"},
		{"Path with dangerous char &", "/projects/&", "dangerous character"},
		{"Path with dangerous char ;", "/projects/;", "dangerous character"},
		{"Path with dangerous char `", "/projects/`", "dangerous character"},
		{"Path with dangerous char $", "/projects/$", "dangerous character"},
		{"Path with dangerous char (", "/projects/(", "dangerous character"},
		{"Path with dangerous char )", "/projects/)", "dangerous character"},
		{"Path with dangerous char {", "/projects/{", "dangerous character"},
		{"Path with dangerous char }", "/projects/}", "dangerous character"},
		{"Too long path", "/" + string(make([]byte, 4097)), "project path too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectPath(tt.path)
			if tt.expected == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expected)
				} else if err.Error() != "" && !contains(err.Error(), tt.expected) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expected, err)
				}
			}
		})
	}
}

func TestValidateTaskName(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		expected string // empty string means no error expected
	}{
		{"Valid task name", "build", ""},
		{"Valid task with colon", "test:unit", ""},
		{"Valid task with underscore", "build_test", ""},
		{"Valid task with hyphen", "build-test", ""},
		{"Empty task name", "", "task name cannot be empty"},
		{"Task with space", "build test", "invalid task name format"},
		{"Task with dot", "build.test", "invalid task name format"},
		{"Task with script tag", "<script>alert()</script>", "dangerous pattern"},
		{"Task with javascript", "javascript:alert()", "dangerous pattern"},
		{"Task with data URI", "data:text/html", "dangerous pattern"},
		{"Task with command separator", "build;rm -rf", "dangerous pattern"},
		{"Task with command chaining", "build&&rm", "dangerous pattern"},
		{"Task with pipe", "build|cat", "dangerous pattern"},
		{"Task with backtick", "build`rm`", "dangerous pattern"},
		{"Task with dollar", "build$", "dangerous pattern"},
		{"Task with subshell", "build$(rm)", "dangerous pattern"},
		{"Too long task name", strings.Repeat("a", 257), "task name too long:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskName(tt.taskName)
			if tt.expected == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expected)
				} else if err.Error() != "" && !contains(err.Error(), tt.expected) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expected, err)
				}
			}
		})
	}
}

func TestValidateBuildOptions(t *testing.T) {
	// Test valid options
	validOptions := map[string]string{
		"clean":       "true",
		"parallel":    "4",
		"build-cache": "enabled",
	}
	if err := ValidateBuildOptions(validOptions); err != nil {
		t.Errorf("Expected valid options, got error: %v", err)
	}

	// Test nil options (should be valid)
	if err := ValidateBuildOptions(nil); err != nil {
		t.Errorf("Expected nil options to be valid, got error: %v", err)
	}

	// Test empty options (should be valid)
	if err := ValidateBuildOptions(map[string]string{}); err != nil {
		t.Errorf("Expected empty options to be valid, got error: %v", err)
	}

	// Test options too large
	largeOptions := make(map[string]string)
	largeKey := string(make([]byte, 5000))
	largeOptions[largeKey] = "value"
	if err := ValidateBuildOptions(largeOptions); err == nil {
		t.Error("Expected error for options too large")
	}

	// Test invalid key
	invalidKeyOptions := map[string]string{
		"invalid key!": "value",
	}
	if err := ValidateBuildOptions(invalidKeyOptions); err == nil {
		t.Error("Expected error for invalid key")
	}

	// Test invalid value
	invalidValueOptions := map[string]string{
		"option": "<script>alert()</script>",
	}
	if err := ValidateBuildOptions(invalidValueOptions); err == nil {
		t.Error("Expected error for invalid value")
	}
}

func TestValidateWorkerConfig(t *testing.T) {
	validConfig := &types.WorkerConfig{
		ID:                  "worker-1",
		CoordinatorURL:      "http://localhost:8080",
		HTTPPort:            8080,
		RPCPort:             8081,
		MaxConcurrentBuilds: 5,
	}

	// Test valid config
	if err := ValidateWorkerConfig(validConfig); err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}

	// Test missing ID
	invalidConfig := &types.WorkerConfig{
		CoordinatorURL:      "http://localhost:8080",
		HTTPPort:            8080,
		RPCPort:             8081,
		MaxConcurrentBuilds: 5,
	}
	if err := ValidateWorkerConfig(invalidConfig); err == nil {
		t.Error("Expected error for missing ID")
	}

	// Test missing coordinator URL
	invalidConfig = &types.WorkerConfig{
		ID:                  "worker-1",
		HTTPPort:            8080,
		RPCPort:             8081,
		MaxConcurrentBuilds: 5,
	}
	if err := ValidateWorkerConfig(invalidConfig); err == nil {
		t.Error("Expected error for missing coordinator URL")
	}

	// Test invalid URL
	invalidConfig = &types.WorkerConfig{
		ID:                  "worker-1",
		CoordinatorURL:      "://invalid-url",
		HTTPPort:            8080,
		RPCPort:             8081,
		MaxConcurrentBuilds: 5,
	}
	if err := ValidateWorkerConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid URL")
	}

	// Test invalid HTTP port
	invalidConfig = &types.WorkerConfig{
		ID:                  "worker-1",
		CoordinatorURL:      "http://localhost:8080",
		HTTPPort:            0,
		RPCPort:             8081,
		MaxConcurrentBuilds: 5,
	}
	if err := ValidateWorkerConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid HTTP port")
	}

	// Test invalid RPC port
	invalidConfig = &types.WorkerConfig{
		ID:                  "worker-1",
		CoordinatorURL:      "http://localhost:8080",
		HTTPPort:            8080,
		RPCPort:             70000,
		MaxConcurrentBuilds: 5,
	}
	if err := ValidateWorkerConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid RPC port")
	}

	// Test invalid max concurrent builds
	invalidConfig = &types.WorkerConfig{
		ID:                  "worker-1",
		CoordinatorURL:      "http://localhost:8080",
		HTTPPort:            8080,
		RPCPort:             8081,
		MaxConcurrentBuilds: 0,
	}
	if err := ValidateWorkerConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid max concurrent builds")
	}

	// Test max concurrent builds too high
	invalidConfig = &types.WorkerConfig{
		ID:                  "worker-1",
		CoordinatorURL:      "http://localhost:8080",
		HTTPPort:            8080,
		RPCPort:             8081,
		MaxConcurrentBuilds: 101,
	}
	if err := ValidateWorkerConfig(invalidConfig); err == nil {
		t.Error("Expected error for max concurrent builds too high")
	}
}

func TestValidateCoordinatorConfig(t *testing.T) {
	validConfig := &types.CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          8081,
		MaxWorkers:       10,
		QueueSize:        100,
		HeartbeatTimeout: 30 * time.Second,
	}

	// Test valid config
	if err := ValidateCoordinatorConfig(validConfig); err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}

	// Test invalid HTTP port
	invalidConfig := &types.CoordinatorConfig{
		HTTPPort:         0,
		RPCPort:          8081,
		MaxWorkers:       10,
		QueueSize:        100,
		HeartbeatTimeout: 30 * time.Second,
	}
	if err := ValidateCoordinatorConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid HTTP port")
	}

	// Test invalid RPC port
	invalidConfig = &types.CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          0,
		MaxWorkers:       10,
		QueueSize:        100,
		HeartbeatTimeout: 30 * time.Second,
	}
	if err := ValidateCoordinatorConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid RPC port")
	}

	// Test invalid max workers
	invalidConfig = &types.CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          8081,
		MaxWorkers:       0,
		QueueSize:        100,
		HeartbeatTimeout: 30 * time.Second,
	}
	if err := ValidateCoordinatorConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid max workers")
	}

	// Test max workers too high
	invalidConfig = &types.CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          8081,
		MaxWorkers:       1001,
		QueueSize:        100,
		HeartbeatTimeout: 30 * time.Second,
	}
	if err := ValidateCoordinatorConfig(invalidConfig); err == nil {
		t.Error("Expected error for max workers too high")
	}

	// Test invalid queue size
	invalidConfig = &types.CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          8081,
		MaxWorkers:       10,
		QueueSize:        0,
		HeartbeatTimeout: 30 * time.Second,
	}
	if err := ValidateCoordinatorConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid queue size")
	}

	// Test queue size too high
	invalidConfig = &types.CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          8081,
		MaxWorkers:       10,
		QueueSize:        10001,
		HeartbeatTimeout: 30 * time.Second,
	}
	if err := ValidateCoordinatorConfig(invalidConfig); err == nil {
		t.Error("Expected error for queue size too high")
	}

	// Test invalid heartbeat timeout
	invalidConfig = &types.CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          8081,
		MaxWorkers:       10,
		QueueSize:        100,
		HeartbeatTimeout: 500 * time.Millisecond,
	}
	if err := ValidateCoordinatorConfig(invalidConfig); err == nil {
		t.Error("Expected error for heartbeat timeout too short")
	}

	// Test heartbeat timeout too long
	invalidConfig = &types.CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          8081,
		MaxWorkers:       10,
		QueueSize:        100,
		HeartbeatTimeout: 2 * time.Hour,
	}
	if err := ValidateCoordinatorConfig(invalidConfig); err == nil {
		t.Error("Expected error for heartbeat timeout too long")
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal input", "hello world", "hello world"},
		{"Input with newline", "hello\nworld", "helloworld"},
		{"Input with carriage return", "hello\rworld", "helloworld"},
		{"Input with tab", "hello\tworld", "helloworld"},
		{"Input with multiple special chars", "hello\n\r\tworld", "helloworld"},
		{"Long input", string(make([]byte, 250)), string(make([]byte, 200)) + "..."},
		{"Exactly 200 chars", string(make([]byte, 200)), string(make([]byte, 200))},
		{"Just over 200 chars", string(make([]byte, 201)), string(make([]byte, 200)) + "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidateOptionKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string // empty string means no error expected
	}{
		{"Valid key", "clean", ""},
		{"Valid key with underscore", "build_cache", ""},
		{"Valid key with hyphen", "build-cache", ""},
		{"Empty key", "", "option key cannot be empty"},
		{"Key with space", "build cache", "invalid key format"},
		{"Key with dot", "build.cache", "invalid key format"},
		{"Key with special char", "build@cache", "invalid key format"},
		{"Too long key", strings.Repeat("a", 101), "key too long:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptionKey(tt.key)
			if tt.expected == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expected)
				} else if err.Error() != "" && !contains(err.Error(), tt.expected) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expected, err)
				}
			}
		})
	}
}

func TestValidateOptionValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string // empty string means no error expected
	}{
		{"Valid value", "true", ""},
		{"Valid long value", string(make([]byte, 1000)), ""},
		{"Too long value", string(make([]byte, 1001)), "value too long"},
		{"Value with script tag", "<script>alert()</script>", "dangerous pattern"},
		{"Value with javascript", "javascript:alert()", "dangerous pattern"},
		{"Value with command separator", "value;rm -rf", "dangerous pattern"},
		{"Value with command chaining", "value&&rm", "dangerous pattern"},
		{"Value with pipe", "value|cat", "dangerous pattern"},
		{"Value with backtick", "value`rm`", "dangerous pattern"},
		{"Value with dollar", "value$", "dangerous pattern"},
		{"Value with subshell", "value$(rm)", "dangerous pattern"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptionValue(tt.value)
			if tt.expected == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expected)
				} else if err.Error() != "" && !contains(err.Error(), tt.expected) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expected, err)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
