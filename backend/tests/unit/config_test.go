package unit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"genVideoSub/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfig 測試配置加載
func TestLoadConfig(t *testing.T) {
	// 創建臨時配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	testConfig := map[string]interface{}{
		"mode":        "production",
		"environment": "production",
		"api_provider": "fal_ai",
		"server": map[string]interface{}{
			"port": 8080,
			"host": "localhost",
		},
		"api_provider_configs": map[string]interface{}{
			"fal_ai": map[string]interface{}{
				"api_key":  "test-key",
				"base_url": "https://fal.run/fal-ai",
				"enabled":  true,
			},
		},
	}

	configData, err := json.Marshal(testConfig)
	require.NoError(t, err)

	err = os.WriteFile(configPath, configData, 0644)
	require.NoError(t, err)

	// 測試加載配置
	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, "production", cfg.Mode)
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "fal_ai", cfg.APIProvider)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.True(t, cfg.APIProviderConfigs.FalAI.Enabled)
	assert.Equal(t, "test-key", cfg.APIProviderConfigs.FalAI.APIKey)
}

// TestIsProductionMode 測試生產模式檢查
func TestIsProductionMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"production_mode", "production", true},
		{"development_mode", "development", false},
		{"test_mode", "test", false},
		{"empty_mode", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Mode: tt.mode}
			assert.Equal(t, tt.expected, cfg.IsProductionMode())
		})
	}
}

// TestGetCurrentProviderConfig 測試獲取當前提供商配置
func TestGetCurrentProviderConfig(t *testing.T) {
	cfg := &config.Config{
		APIProvider: "fal_ai",
		APIProviderConfigs: config.APIProviderConfigs{
			FalAI: config.ProviderConfig{
				APIKey:  "fal-key",
				BaseURL: "https://fal.run/fal-ai",
				Enabled: true,
			},
			RunwayML: config.ProviderConfig{
				APIKey:  "runway-key",
				BaseURL: "https://api.runwayml.com",
				Enabled: false,
			},
		},
	}

	// 測試獲取fal_ai配置
	providerCfg := cfg.GetCurrentProviderConfig()
	assert.Equal(t, "fal-key", providerCfg.APIKey)
	assert.Equal(t, "https://fal.run/fal-ai", providerCfg.BaseURL)
	assert.True(t, providerCfg.Enabled)

	// 測試切換到runway_ml
	cfg.APIProvider = "runway_ml"
	providerCfg = cfg.GetCurrentProviderConfig()
	assert.Equal(t, "runway-key", providerCfg.APIKey)
	assert.Equal(t, "https://api.runwayml.com", providerCfg.BaseURL)
	assert.False(t, providerCfg.Enabled)
}

// TestSetAPIProvider 測試設置API提供商
func TestSetAPIProvider(t *testing.T) {
	cfg := &config.Config{
		APIProvider: "fal_ai",
	}

	cfg.SetAPIProvider("runway_ml")
	assert.Equal(t, "runway_ml", cfg.APIProvider)

	cfg.SetAPIProvider("stability_ai")
	assert.Equal(t, "stability_ai", cfg.APIProvider)
}

// TestEnableDisableProvider 測試啟用/禁用提供商
func TestEnableDisableProvider(t *testing.T) {
	cfg := &config.Config{
		APIProviderConfigs: config.APIProviderConfigs{
			FalAI: config.ProviderConfig{
				Enabled: true,
			},
			RunwayML: config.ProviderConfig{
				Enabled: false,
			},
		},
	}

	// 測試禁用fal_ai
	cfg.DisableProvider("fal_ai")
	assert.False(t, cfg.APIProviderConfigs.FalAI.Enabled)

	// 測試啟用runway_ml
	cfg.EnableProvider("runway_ml")
	assert.True(t, cfg.APIProviderConfigs.RunwayML.Enabled)

	// 測試無效提供商
	cfg.EnableProvider("invalid_provider")
	// 應該不會崩潰，但也不會改變任何狀態
}

// TestDefaultConfig 測試默認配置
func TestDefaultConfig(t *testing.T) {
	cfg := config.NewDefaultConfig()

	assert.Equal(t, "production", cfg.Mode)
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "fal_ai", cfg.APIProvider)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.True(t, cfg.APIProviderConfigs.FalAI.Enabled)
	assert.False(t, cfg.APIProviderConfigs.RunwayML.Enabled)
	assert.False(t, cfg.APIProviderConfigs.StabilityAI.Enabled)
	assert.True(t, cfg.Curl.CurlEnabled)
	assert.Equal(t, 30, cfg.Curl.CurlTimeout)
	assert.Equal(t, 3, cfg.Curl.CurlRetries)
}

// TestConfigValidation 測試配置驗證
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		valid   bool
	}{
		{
			name: "valid_config",
			config: &config.Config{
				Mode:        "production",
				Environment: "production",
				APIProvider: "fal_ai",
				Server: config.ServerConfig{
					Port: 8080,
					Host: "localhost",
				},
				APIProviderConfigs: config.APIProviderConfigs{
					FalAI: config.ProviderConfig{
						APIKey:  "test-key",
						Enabled: true,
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid_port",
			config: &config.Config{
				Mode:        "production",
				Environment: "production",
				APIProvider: "fal_ai",
				Server: config.ServerConfig{
					Port: -1, // 無效端口
					Host: "localhost",
				},
			},
			valid: false,
		},
		{
			name: "empty_api_provider",
			config: &config.Config{
				Mode:        "production",
				Environment: "production",
				APIProvider: "", // 空的API提供商
				Server: config.ServerConfig{
					Port: 8080,
					Host: "localhost",
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}