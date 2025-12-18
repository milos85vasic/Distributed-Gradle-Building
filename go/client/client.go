package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GradleBuildClient provides a Go client for the distributed build system
type GradleBuildClient struct {
	BaseURL     string
	HTTPClient  *http.Client
	AuthToken   string
}

// NewClient creates a new client for the distributed build system
func NewClient(baseURL string) *GradleBuildClient {
	return &GradleBuildClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewAuthenticatedClient creates a new authenticated client
func NewAuthenticatedClient(baseURL, authToken string) *GradleBuildClient {
	client := NewClient(baseURL)
	client.AuthToken = authToken
	return client
}

// BuildRequest represents a build request
type BuildRequest struct {
	ProjectPath  string            `json:"project_path"`
	TaskName     string            `json:"task_name"`
	WorkerID     string            `json:"worker_id,omitempty"`
	CacheEnabled bool              `json:"cache_enabled"`
	BuildOptions map[string]string `json:"build_options,omitempty"`
}

// BuildResponse represents the response to a build request
type BuildResponse struct {
	BuildID string `json:"build_id"`
	Status  string `json:"status"`
}

// BuildStatus represents the status of a build
type BuildStatus struct {
	BuildID       string        `json:"build_id"`
	WorkerID      string        `json:"worker_id"`
	ProjectPath   string        `json:"project_path"`
	TaskName      string        `json:"task_name"`
	Status        string        `json:"status"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      time.Duration `json:"duration"`
	Success       bool          `json:"success"`
	CacheHitRate  float64       `json:"cache_hit_rate"`
	Artifacts     []string      `json:"artifacts"`
	ErrorMessage  string        `json:"error_message,omitempty"`
}

// WorkerInfo represents information about a worker
type WorkerInfo struct {
	ID              string    `json:"id"`
	Host            string    `json:"host"`
	Status          string    `json:"status"`
	CPU             float64   `json:"cpu"`
	Memory          uint64    `json:"memory"`
	DiskUsage       uint64    `json:"disk_usage"`
	BuildCount      int       `json:"build_count"`
	LastBuildTime   time.Time `json:"last_build_time"`
	AverageBuildTime time.Duration `json:"average_build_time"`
	LastCheckin     time.Time `json:"last_checkin"`
	ResponseTime    time.Duration `json:"response_time"`
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	Timestamp     time.Time `json:"timestamp"`
	WorkerCount   int       `json:"worker_count"`
	QueueLength   int       `json:"queue_length"`
	ActiveBuilds  int       `json:"active_builds"`
}

// SubmitBuild submits a build request to the coordinator
func (c *GradleBuildClient) SubmitBuild(req BuildRequest) (*BuildResponse, error) {
	url := fmt.Sprintf("%s/api/build", c.BaseURL)
	
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	if c.AuthToken != "" {
		httpReq.Header.Set("X-Auth-Token", c.AuthToken)
	}
	
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	var buildResp BuildResponse
	if err := json.NewDecoder(resp.Body).Decode(&buildResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	return &buildResp, nil
}

// GetBuildStatus retrieves the status of a specific build
func (c *GradleBuildClient) GetBuildStatus(buildID string) (*BuildStatus, error) {
	url := fmt.Sprintf("%s/api/build/%s", c.BaseURL, buildID)
	
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	if c.AuthToken != "" {
		httpReq.Header.Set("X-Auth-Token", c.AuthToken)
	}
	
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	var status BuildStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	return &status, nil
}

// GetWorkers retrieves information about all workers
func (c *GradleBuildClient) GetWorkers() (map[string]*WorkerInfo, error) {
	url := fmt.Sprintf("%s/api/workers", c.BaseURL)
	
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	if c.AuthToken != "" {
		httpReq.Header.Set("X-Auth-Token", c.AuthToken)
	}
	
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	var workers map[string]*WorkerInfo
	if err := json.NewDecoder(resp.Body).Decode(&workers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	return workers, nil
}

// GetSystemStatus retrieves the overall system status
func (c *GradleBuildClient) GetSystemStatus() (*SystemStatus, error) {
	url := fmt.Sprintf("%s/api/status", c.BaseURL)
	
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	if c.AuthToken != "" {
		httpReq.Header.Set("X-Auth-Token", c.AuthToken)
	}
	
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	var status SystemStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	return &status, nil
}

// WaitForBuild waits for a build to complete and returns its final status
func (c *GradleBuildClient) WaitForBuild(buildID string, timeout time.Duration) (*BuildStatus, error) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		status, err := c.GetBuildStatus(buildID)
		if err != nil {
			return nil, err
		}
		
		if status.Status == "completed" || status.Status == "failed" {
			return status, nil
		}
		
		time.Sleep(5 * time.Second)
	}
	
	return nil, fmt.Errorf("timeout waiting for build %s to complete", buildID)
}

// HealthCheck performs a health check on the system
func (c *GradleBuildClient) HealthCheck() error {
	url := fmt.Sprintf("%s/health", c.BaseURL)
	
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	if c.AuthToken != "" {
		httpReq.Header.Set("X-Auth-Token", c.AuthToken)
	}
	
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
	}
	
	return nil
}