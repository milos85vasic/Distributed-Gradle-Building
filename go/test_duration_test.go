package main_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type Config struct {
	MetricsInterval time.Duration `json:"metrics_interval"`
}

func TestDurationParsing(t *testing.T) {
	data := `{"metrics_interval": "60s"}`
	var config Config
	err := json.Unmarshal([]byte(data), &config)
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Config: %+v\n", config)
}
