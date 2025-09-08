package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 應用程序配置
type Config struct {
	// 服務器配置
	ServerPort string `json:"server_port"`
	ServerHost string `json:"server_host"`
	
	// 運行模式配置
	Mode        string `json:"mode"`         // "mock", "production"
	Environment string `json:"environment"` // "development", "testing", "production"
	
	// FalAI API配置
	FalAPIKey    string `json:"fal_api_key"`
	FalAPIURL    string `json:"fal_api_url"`
	FalTimeout   int    `json:"fal_timeout"`
	
	// API提供商配置
	APIProvider        string                    `json:"api_provider"`         // "fal_ai", "runway_ml", "stability_ai", "custom"
	APIProviderConfigs map[string]*ProviderConfig `json:"api_provider_configs"` // 各提供商的配置
	
	// 存儲配置
	StoragePath     string `json:"storage_path"`
	MaxFileSize     int64  `json:"max_file_size"`
	AllowedFormats  []string `json:"allowed_formats"`
	
	// 工作池配置
	WorkerPoolSize int `json:"worker_pool_size"`
	QueueSize      int `json:"queue_size"`
	
	// Curl配置
	CurlTimeout     int  `json:"curl_timeout"`
	CurlRetries     int  `json:"curl_retries"`
	CurlEnabled     bool `json:"curl_enabled"`
	
	// 測試配置
	TestConfig *TestConfig `json:"test_config,omitempty"`
}

// ProviderConfig API提供商配置
type ProviderConfig struct {
	APIKey   string            `json:"api_key"`
	BaseURL  string            `json:"base_url"`
	Timeout  int               `json:"timeout"`
	Headers  map[string]string `json:"headers,omitempty"`
	Enabled  bool              `json:"enabled"`
}

// TestConfig 測試配置
type TestConfig struct {
	UseMock     bool   `json:"use_mock"`
	Environment string `json:"environment"`
	FalAPIKey   string `json:"fal_api_key,omitempty"`
}

// Validate 驗證測試配置
func (tc *TestConfig) Validate() error {
	if tc == nil {
		return fmt.Errorf("test config is nil")
	}
	
	// 如果不是 mock 模式，需要檢查 API Key
	if !tc.UseMock && tc.FalAPIKey == "" {
		return fmt.Errorf("FalAI API key is required when not using mock")
	}
	
	return nil
}

// IsMockEnabled 檢查是否啟用Mock模式
func (tc *TestConfig) IsMockEnabled() bool {
	return tc.UseMock
}

// GetEnvironment 獲取環境名稱
func (tc *TestConfig) GetEnvironment() string {
	if tc.Environment == "" {
		return "testing"
	}
	return tc.Environment
}

// IsTestMode 檢查是否為測試模式
func (tc *TestConfig) IsTestMode() bool {
	return tc.Environment == "testing" || tc.Environment == "development"
}

// 默認配置值
func getDefaultConfig() *Config {
	return &Config{
		ServerPort: "8080",
		ServerHost: "localhost",
		
		Mode:        "mock",
		Environment: "development",
		
		FalAPIKey:  "",
		FalAPIURL:  "https://fal.run",
		FalTimeout: 300,
		
		APIProvider: "fal_ai",
		APIProviderConfigs: map[string]*ProviderConfig{
			"fal_ai": {
				APIKey:  "",
				BaseURL: "https://fal.run",
				Timeout: 300,
				Enabled: true,
			},
			"runway_ml": {
				APIKey:  "",
				BaseURL: "https://api.runwayml.com",
				Timeout: 300,
				Enabled: false,
			},
			"stability_ai": {
				APIKey:  "",
				BaseURL: "https://api.stability.ai",
				Timeout: 300,
				Enabled: false,
			},
		},
		
		StoragePath:    "./uploads",
		MaxFileSize:    100 * 1024 * 1024, // 100MB
		AllowedFormats: []string{".mp4", ".avi", ".mov", ".mkv"},
		
		WorkerPoolSize: 5,
		QueueSize:      100,
		
		CurlTimeout: 30,
		CurlRetries: 3,
		CurlEnabled: true,
		
		TestConfig: &TestConfig{
			UseMock:     true,
			Environment: "testing",
		},
	}
}

// LoadConfig 載入配置文件
func LoadConfig(path string) (*Config, error) {
	config := getDefaultConfig()

	// 如果配置文件存在，則載入
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// IsMockEnabled 檢查是否啟用Mock模式
func (c *Config) IsMockEnabled() bool {
	if c.Mode == "mock" {
		return true
	}
	if c.TestConfig != nil {
		return c.TestConfig.UseMock
	}
	return false
}

// IsProductionMode 檢查是否為正式模式
func (c *Config) IsProductionMode() bool {
	return c.Mode == "production"
}

// GetEnvironment 獲取環境名稱
func (c *Config) GetEnvironment() string {
	if c.Environment != "" {
		return c.Environment
	}
	if c.TestConfig != nil {
		return c.TestConfig.Environment
	}
	return "production"
}

// GetCurrentProviderConfig 獲取當前API提供商配置
func (c *Config) GetCurrentProviderConfig() *ProviderConfig {
	if c.APIProviderConfigs == nil {
		return nil
	}
	return c.APIProviderConfigs[c.APIProvider]
}

// SetAPIProvider 設置API提供商
func (c *Config) SetAPIProvider(provider string) error {
	if c.APIProviderConfigs == nil {
		return fmt.Errorf("no provider configs available")
	}
	
	config, exists := c.APIProviderConfigs[provider]
	if !exists {
		return fmt.Errorf("provider %s not configured", provider)
	}
	
	if !config.Enabled {
		return fmt.Errorf("provider %s is disabled", provider)
	}
	
	c.APIProvider = provider
	return nil
}

// EnableProvider 啟用API提供商
func (c *Config) EnableProvider(provider string) error {
	if c.APIProviderConfigs == nil {
		return fmt.Errorf("no provider configs available")
	}
	
	config, exists := c.APIProviderConfigs[provider]
	if !exists {
		return fmt.Errorf("provider %s not configured", provider)
	}
	
	config.Enabled = true
	return nil
}

// DisableProvider 禁用API提供商
func (c *Config) DisableProvider(provider string) error {
	if c.APIProviderConfigs == nil {
		return fmt.Errorf("no provider configs available")
	}
	
	config, exists := c.APIProviderConfigs[provider]
	if !exists {
		return fmt.Errorf("provider %s not configured", provider)
	}
	
	config.Enabled = false
	return nil
}

// LoadTestConfig 載入測試配置
func LoadTestConfig() *TestConfig {
	// 嘗試從環境變量或配置文件載入
	// 這裡先返回默認配置
	return &TestConfig{
		UseMock:     true,
		Environment: "testing",
		FalAPIKey:   os.Getenv("FAL_API_KEY"),
	}
}