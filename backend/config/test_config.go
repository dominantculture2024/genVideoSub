package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TestConfig 測試配置結構
type TestConfig struct {
	// 服務配置
	UseMockService bool   `json:"use_mock_service"`
	FalAPIKey      string `json:"fal_api_key"`
	FalAPIURL      string `json:"fal_api_url"`
	
	// 測試配置
	TestTimeout    int    `json:"test_timeout"`
	TestDataPath   string `json:"test_data_path"`
	LogLevel       string `json:"log_level"`
	
	// Mock配置
	MockDelay      int    `json:"mock_delay"`
	MockFailRate   float64 `json:"mock_fail_rate"`
	MockDataSize   int    `json:"mock_data_size"`
	
	// 並發配置
	MaxConcurrent  int    `json:"max_concurrent"`
	RateLimit      int    `json:"rate_limit"`
}

// DefaultTestConfig 默認測試配置
var DefaultTestConfig = TestConfig{
	UseMockService: true,
	FalAPIKey:      "",
	FalAPIURL:      "https://fal.run/fal-ai/fast-svd",
	TestTimeout:    300, // 5分鐘
	TestDataPath:   "./testdata",
	LogLevel:       "info",
	MockDelay:      1000, // 1秒
	MockFailRate:   0.0,  // 0% 失敗率
	MockDataSize:   5,    // 5個mock數據項
	MaxConcurrent:  10,
	RateLimit:      100,
}

// LoadTestConfig 從環境變量加載測試配置
func LoadTestConfig() *TestConfig {
	config := DefaultTestConfig
	
	// 從環境變量覆蓋配置
	if val := os.Getenv("USE_MOCK_SERVICE"); val != "" {
		config.UseMockService = strings.ToLower(val) == "true"
	}
	
	if val := os.Getenv("FAL_API_KEY"); val != "" {
		config.FalAPIKey = val
	}
	
	if val := os.Getenv("FAL_API_URL"); val != "" {
		config.FalAPIURL = val
	}
	
	if val := os.Getenv("TEST_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil {
			config.TestTimeout = timeout
		}
	}
	
	if val := os.Getenv("TEST_DATA_PATH"); val != "" {
		config.TestDataPath = val
	}
	
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.LogLevel = val
	}
	
	if val := os.Getenv("MOCK_DELAY"); val != "" {
		if delay, err := strconv.Atoi(val); err == nil {
			config.MockDelay = delay
		}
	}
	
	if val := os.Getenv("MOCK_FAIL_RATE"); val != "" {
		if rate, err := strconv.ParseFloat(val, 64); err == nil {
			config.MockFailRate = rate
		}
	}
	
	if val := os.Getenv("MOCK_DATA_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.MockDataSize = size
		}
	}
	
	if val := os.Getenv("MAX_CONCURRENT"); val != "" {
		if max, err := strconv.Atoi(val); err == nil {
			config.MaxConcurrent = max
		}
	}
	
	if val := os.Getenv("RATE_LIMIT"); val != "" {
		if limit, err := strconv.Atoi(val); err == nil {
			config.RateLimit = limit
		}
	}
	
	return &config
}

// IsTestMode 檢查是否為測試模式
func (c *TestConfig) IsTestMode() bool {
	return c.UseMockService || c.FalAPIKey == "" || strings.Contains(c.FalAPIKey, "test")
}

// IsMockEnabled 檢查是否啟用mock
func (c *TestConfig) IsMockEnabled() bool {
	return c.UseMockService
}

// IsProductionMode 檢查是否為生產模式
func (c *TestConfig) IsProductionMode() bool {
	return !c.IsTestMode() && c.FalAPIKey != "" && !strings.Contains(c.FalAPIKey, "test")
}

// GetEnvironment 獲取當前環境
func (c *TestConfig) GetEnvironment() string {
	if c.IsProductionMode() {
		return "production"
	} else if c.IsMockEnabled() {
		return "mock"
	} else {
		return "test"
	}
}

// Validate 驗證配置
func (c *TestConfig) Validate() error {
	if !c.UseMockService && c.FalAPIKey == "" {
		return fmt.Errorf("FAL_API_KEY is required when not using mock service")
	}
	
	if c.TestTimeout <= 0 {
		return fmt.Errorf("test timeout must be positive")
	}
	
	if c.MockFailRate < 0 || c.MockFailRate > 1 {
		return fmt.Errorf("mock fail rate must be between 0 and 1")
	}
	
	if c.MaxConcurrent <= 0 {
		return fmt.Errorf("max concurrent must be positive")
	}
	
	return nil
}