package validation

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"distributed-gradle-building/types"
)

// Validator interface for different validation types
type Validator interface {
	Validate() error
}

// ValidateBuildRequest validates a build request
func ValidateBuildRequest(req *types.BuildRequest) error {
	if err := ValidateProjectPath(req.ProjectPath); err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}

	if err := ValidateTaskName(req.TaskName); err != nil {
		return fmt.Errorf("invalid task name: %w", err)
	}

	if err := ValidateBuildOptions(req.BuildOptions); err != nil {
		return fmt.Errorf("invalid build options: %w", err)
	}

	if req.RequestID == "" {
		return fmt.Errorf("request ID is required")
	}

	return nil
}

// ValidateProjectPath validates project path for security
func ValidateProjectPath(path string) error {
	if path == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Validate absolute path format
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("project path must be absolute: %s", path)
	}

	// Check for dangerous characters
	dangerousChars := []string{"<", ">", "|", "&", ";", "`", "$", "(", ")", "{", "}"}
	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("dangerous character '%s' in path: %s", char, path)
		}
	}

	// Validate path length
	if len(path) > 4096 {
		return fmt.Errorf("project path too long: %d characters", len(path))
	}

	return nil
}

// ValidateTaskName validates Gradle task name
func ValidateTaskName(taskName string) error {
	if taskName == "" {
		return fmt.Errorf("task name cannot be empty")
	}

	// Gradle task name regex (alphanumeric, hyphens, underscores, colons)
	validTaskName := regexp.MustCompile(`^[a-zA-Z0-9:_-]+$`)
	if !validTaskName.MatchString(taskName) {
		return fmt.Errorf("invalid task name format: %s", taskName)
	}

	// Check for dangerous characters or patterns
	dangerousPatterns := []string{
		"<script", "</script>", "javascript:", "data:", "vbscript:",
		"&&", "||", "|", ";", "&", "`", "$", "$(", "${",
	}

	lowerTaskName := strings.ToLower(taskName)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerTaskName, pattern) {
			return fmt.Errorf("dangerous pattern '%s' in task name", pattern)
		}
	}

	// Validate length
	if len(taskName) > 256 {
		return fmt.Errorf("task name too long: %d characters", len(taskName))
	}

	return nil
}

// ValidateBuildOptions validates build options map
func ValidateBuildOptions(options map[string]string) error {
	if options == nil {
		return nil
	}

	// Validate total size
	totalSize := 0
	for k, v := range options {
		totalSize += len(k) + len(v)
	}

	if totalSize > 10240 { // 10KB limit
		return fmt.Errorf("build options too large: %d bytes", totalSize)
	}

	// Validate each option
	for key, value := range options {
		if err := validateOptionKey(key); err != nil {
			return fmt.Errorf("invalid option key '%s': %w", key, err)
		}

		if err := validateOptionValue(value); err != nil {
			return fmt.Errorf("invalid option value for key '%s': %w", key, err)
		}
	}

	return nil
}

// validateOptionKey validates build option key
func validateOptionKey(key string) error {
	if key == "" {
		return fmt.Errorf("option key cannot be empty")
	}

	// Allowed characters: alphanumeric, hyphens, underscores
	validKey := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validKey.MatchString(key) {
		return fmt.Errorf("invalid key format: %s", key)
	}

	if len(key) > 100 {
		return fmt.Errorf("key too long: %d characters", len(key))
	}

	return nil
}

// validateOptionValue validates build option value
func validateOptionValue(value string) error {
	if len(value) > 1000 {
		return fmt.Errorf("value too long: %d characters", len(value))
	}

	// Check for script injection patterns
	dangerousPatterns := []string{
		"<script", "</script>", "javascript:", "data:", "vbscript:",
		"&&", "||", "|", ";", "&", "`", "$", "$(", "${",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerValue, pattern) {
			return fmt.Errorf("dangerous pattern '%s' in value", pattern)
		}
	}

	return nil
}

// ValidateWorkerConfig validates worker configuration
func ValidateWorkerConfig(config *types.WorkerConfig) error {
	if config.ID == "" {
		return fmt.Errorf("worker ID is required")
	}

	if config.CoordinatorURL == "" {
		return fmt.Errorf("coordinator URL is required")
	}

	// Validate URL format
	if _, err := url.Parse(config.CoordinatorURL); err != nil {
		return fmt.Errorf("invalid coordinator URL: %w", err)
	}

	if config.HTTPPort < 1 || config.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", config.HTTPPort)
	}

	if config.RPCPort < 1 || config.RPCPort > 65535 {
		return fmt.Errorf("invalid RPC port: %d", config.RPCPort)
	}

	if config.MaxConcurrentBuilds < 1 || config.MaxConcurrentBuilds > 100 {
		return fmt.Errorf("invalid max concurrent builds: %d (must be 1-100)", config.MaxConcurrentBuilds)
	}

	return nil
}

// ValidateCoordinatorConfig validates coordinator configuration
func ValidateCoordinatorConfig(config *types.CoordinatorConfig) error {
	if config.HTTPPort < 1 || config.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", config.HTTPPort)
	}

	if config.RPCPort < 1 || config.RPCPort > 65535 {
		return fmt.Errorf("invalid RPC port: %d", config.RPCPort)
	}

	if config.MaxWorkers < 1 || config.MaxWorkers > 1000 {
		return fmt.Errorf("invalid max workers: %d (must be 1-1000)", config.MaxWorkers)
	}

	if config.QueueSize < 1 || config.QueueSize > 10000 {
		return fmt.Errorf("invalid queue size: %d (must be 1-10000)", config.QueueSize)
	}

	if config.HeartbeatTimeout < time.Second || config.HeartbeatTimeout > time.Hour {
		return fmt.Errorf("invalid heartbeat timeout: %v (must be 1s-1h)", config.HeartbeatTimeout)
	}

	return nil
}

// SanitizeInput sanitizes user input for logging
func SanitizeInput(input string) string {
	// Remove or replace dangerous characters
	sanitized := strings.ReplaceAll(input, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\t", "")

	// Limit length for logging
	if len(sanitized) > 200 {
		sanitized = sanitized[:200] + "..."
	}

	return sanitized
}
