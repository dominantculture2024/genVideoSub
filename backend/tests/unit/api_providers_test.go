package unit

import (
	"testing"
	"time"

	"genVideoSub/config"
	"genVideoSub/interfaces"
	"genVideoSub/models"
	"genVideoSub/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFalAIService 測試FalAI服務
func TestFalAIService(t *testing.T) {
	config := &config.APIProviderConfig{
		APIKey:  "test-fal-key",
		BaseURL: "https://fal.run",
		Timeout: 300,
		Enabled: true,
	}
	
	service := services.NewFalAIService(config.APIKey, config.BaseURL, time.Duration(config.Timeout)*time.Second)
	require.NotNil(t, service)
	
	// 測試提供商名稱
	assert.Equal(t, "fal_ai", service.GetProviderName())
	
	// 測試配置驗證
	err := service.ValidateConfig()
	assert.NoError(t, err) // 基本配置驗證應該通過
	
	// 測試任務提交（這會失敗，因為使用測試API密鑰）
	task := &models.Task{
		Prompt: "Test prompt",
		Model:  "fal-ai/fast-svd",
		FalAIInput: &models.FalAIInput{
			ImageURL:       "https://example.com/test.jpg",
			MotionBucketID: 127,
			FPS:            6,
			CondAug:        0.02,
			Steps:          25,
			Seed:           42,
		},
	}
	
	// 注意：這個測試會失敗，因為我們使用的是測試API密鑰
	// 但我們可以測試方法是否正確調用
	_, err = service.SubmitTask(task)
	assert.Error(t, err) // 預期會有錯誤，因為API密鑰無效
}

// TestRunwayMLService 測試RunwayML服務
func TestRunwayMLService(t *testing.T) {
	config := &config.APIProviderConfig{
		APIKey:  "test-runway-key",
		BaseURL: "https://api.runwayml.com",
		Timeout: 300,
		Enabled: true,
	}
	
	service := services.NewRunwayMLService(config.APIKey, config.BaseURL, time.Duration(config.Timeout)*time.Second)
	require.NotNil(t, service)
	
	// 測試提供商名稱
	assert.Equal(t, "runway_ml", service.GetProviderName())
	
	// 測試配置驗證
	err := service.ValidateConfig()
	assert.NoError(t, err)
	
	// 測試任務提交
	task := &models.Task{
		Prompt: "Test runway prompt",
		Model:  "gen2",
		FalAIInput: &models.FalAIInput{
			ImageURL: "https://example.com/test.jpg",
			FPS:      24,
			Steps:    25,
		},
	}
	
	_, err = service.SubmitTask(task)
	assert.Error(t, err) // 預期會有錯誤，因為API密鑰無效
}

// TestStabilityAIService 測試StabilityAI服務
func TestStabilityAIService(t *testing.T) {
	config := &config.APIProviderConfig{
		APIKey:  "test-stability-key",
		BaseURL: "https://api.stability.ai",
		Timeout: 300,
		Enabled: true,
	}
	
	service := services.NewStabilityAIService(config.APIKey, config.BaseURL, time.Duration(config.Timeout)*time.Second)
	require.NotNil(t, service)
	
	// 測試提供商名稱
	assert.Equal(t, "stability_ai", service.GetProviderName())
	
	// 測試配置驗證
	err := service.ValidateConfig()
	assert.NoError(t, err)
	
	// 測試任務提交
	task := &models.Task{
		Prompt: "Test stability prompt",
		Model:  "stable-video-diffusion-img2vid-xt",
		FalAIInput: &models.FalAIInput{
			ImageURL:       "https://example.com/test.jpg",
			MotionBucketID: 127,
			Seed:           42,
		},
	}
	
	_, err = service.SubmitTask(task)
	assert.Error(t, err) // 預期會有錯誤，因為API密鑰無效
}

// TestAPIProviderManager 測試API提供商管理器
func TestAPIProviderManager(t *testing.T) {
	config := &config.Config{
		APIProvider: "fal_ai",
		APIProviderConfigs: map[string]*config.APIProviderConfig{
			"fal_ai": {
				APIKey:  "test-fal-key",
				BaseURL: "https://fal.run",
				Timeout: 300,
				Enabled: true,
			},
			"runway_ml": {
				APIKey:  "test-runway-key",
				BaseURL: "https://api.runwayml.com",
				Timeout: 300,
				Enabled: true,
			},
			"stability_ai": {
				APIKey:  "test-stability-key",
				BaseURL: "https://api.stability.ai",
				Timeout: 300,
				Enabled: false, // 測試禁用狀態
			},
		},
	}
	
	manager := services.NewAPIProviderManager(config)
	require.NotNil(t, manager)
	
	// 測試獲取當前提供商
	currentProvider := manager.GetCurrentProvider()
	assert.Equal(t, "fal_ai", currentProvider)
	
	// 測試獲取當前服務
	currentService := manager.GetCurrentService()
	require.NotNil(t, currentService)
	assert.Equal(t, "fal_ai", currentService.GetProviderName())
	
	// 測試列出提供商
	providers := manager.ListProviders()
	assert.Contains(t, providers, "fal_ai")
	assert.Contains(t, providers, "runway_ml")
	// stability_ai 被禁用，所以不應該在列表中
	assert.NotContains(t, providers, "stability_ai")
	
	// 測試切換提供商
	err := manager.SwitchProvider("runway_ml")
	assert.NoError(t, err)
	assert.Equal(t, "runway_ml", manager.GetCurrentProvider())
	
	// 測試切換到無效提供商
	err = manager.SwitchProvider("invalid_provider")
	assert.Error(t, err)
	
	// 測試切換到禁用的提供商
	err = manager.SwitchProvider("stability_ai")
	assert.Error(t, err)
}

// TestAPIProviderManagerTaskOperations 測試API提供商管理器的任務操作
func TestAPIProviderManagerTaskOperations(t *testing.T) {
	config := &config.Config{
		APIProvider: "fal_ai",
		APIProviderConfigs: map[string]*config.APIProviderConfig{
			"fal_ai": {
				APIKey:  "test-fal-key",
				BaseURL: "https://fal.run",
				Timeout: 300,
				Enabled: true,
			},
		},
	}
	
	manager := services.NewAPIProviderManager(config)
	require.NotNil(t, manager)
	
	task := &models.Task{
		Prompt: "Test task",
		Model:  "fal-ai/fast-svd",
		FalAIInput: &models.FalAIInput{
			ImageURL: "https://example.com/test.jpg",
			FPS:      6,
		},
	}
	
	// 測試提交任務
	_, err := manager.SubmitTask(task)
	assert.Error(t, err) // 預期會有錯誤，因為API密鑰無效
	
	// 測試獲取任務狀態
	_, err = manager.GetTaskStatus("test-request-id")
	assert.Error(t, err) // 預期會有錯誤
	
	// 測試獲取任務結果
	_, err = manager.GetTaskResult("test-request-id")
	assert.Error(t, err) // 預期會有錯誤
}

// TestProviderServiceAdapters 測試提供商服務適配器
func TestProviderServiceAdapters(t *testing.T) {
	// 測試FalAI適配器
	falService := services.NewFalAIService("test-key", "https://fal.run", 300*time.Second)
	falAdapter := &services.FalAIProviderService{Service: falService}
	
	assert.Equal(t, "fal_ai", falAdapter.GetProviderName())
	assert.NoError(t, falAdapter.ValidateConfig())
	
	// 測試RunwayML適配器
	runwayService := services.NewRunwayMLService("test-key", "https://api.runwayml.com", 300*time.Second)
	runwayAdapter := &services.RunwayMLProviderService{Service: runwayService}
	
	assert.Equal(t, "runway_ml", runwayAdapter.GetProviderName())
	assert.NoError(t, runwayAdapter.ValidateConfig())
	
	// 測試StabilityAI適配器
	stabilityService := services.NewStabilityAIService("test-key", "https://api.stability.ai", 300*time.Second)
	stabilityAdapter := &services.StabilityAIProviderService{Service: stabilityService}
	
	assert.Equal(t, "stability_ai", stabilityAdapter.GetProviderName())
	assert.NoError(t, stabilityAdapter.ValidateConfig())
}

// TestProviderConfigValidation 測試提供商配置驗證
func TestProviderConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.APIProviderConfig
		expectErr bool
	}{
		{
			name: "valid_config",
			config: &config.APIProviderConfig{
				APIKey:  "valid-key",
				BaseURL: "https://api.example.com",
				Timeout: 300,
				Enabled: true,
			},
			expectErr: false,
		},
		{
			name: "empty_api_key",
			config: &config.APIProviderConfig{
				APIKey:  "",
				BaseURL: "https://api.example.com",
				Timeout: 300,
				Enabled: true,
			},
			expectErr: true,
		},
		{
			name: "empty_base_url",
			config: &config.APIProviderConfig{
				APIKey:  "valid-key",
				BaseURL: "",
				Timeout: 300,
				Enabled: true,
			},
			expectErr: true,
		},
		{
			name: "zero_timeout",
			config: &config.APIProviderConfig{
				APIKey:  "valid-key",
				BaseURL: "https://api.example.com",
				Timeout: 0,
				Enabled: true,
			},
			expectErr: false, // 零超時應該使用默認值
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 測試FalAI服務配置驗證
			falService := services.NewFalAIService(
				tt.config.APIKey,
				tt.config.BaseURL,
				time.Duration(tt.config.Timeout)*time.Second,
			)
			
			err := falService.ValidateConfig()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRetryWithBackoff 測試重試機制
func TestRetryWithBackoff(t *testing.T) {
	service := services.NewFalAIService("test-key", "https://fal.run", 300*time.Second)
	
	// 測試重試函數（這會失敗，但我們可以測試重試邏輯）
	retryCount := 0
	maxRetries := 3
	
	err := service.RetryWithBackoff(func() error {
		retryCount++
		if retryCount < maxRetries {
			return assert.AnError // 模擬錯誤
		}
		return nil // 最後一次成功
	}, maxRetries, time.Millisecond)
	
	assert.NoError(t, err)
	assert.Equal(t, maxRetries, retryCount)
	
	// 測試所有重試都失敗的情況
	retryCount = 0
	err = service.RetryWithBackoff(func() error {
		retryCount++
		return assert.AnError // 總是失敗
	}, maxRetries, time.Millisecond)
	
	assert.Error(t, err)
	assert.Equal(t, maxRetries, retryCount)
}