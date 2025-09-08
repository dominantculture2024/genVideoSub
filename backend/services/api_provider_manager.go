package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"genVideoSub/config"
	"genVideoSub/interfaces"
	"genVideoSub/models"
)

// APIProviderManager API提供商管理器
type APIProviderManager struct {
	providers        map[interfaces.APIProvider]interfaces.VideoGenerationInterface
	current          interfaces.APIProvider
	defaultProvider  interfaces.APIProvider
	config           *config.Config
	mu               sync.RWMutex
}

// NewAPIProviderManager 創建新的API提供商管理器
func NewAPIProviderManager(cfg *config.Config) *APIProviderManager {
	defaultProv := interfaces.APIProvider(cfg.APIProvider)
	manager := &APIProviderManager{
		providers:       make(map[interfaces.APIProvider]interfaces.VideoGenerationInterface),
		current:         defaultProv,
		defaultProvider: defaultProv,
		config:          cfg,
	}

	// 註冊所有啟用的提供商
	manager.registerProviders()

	return manager
}

// registerProviders 註冊所有可用的API提供商
func (m *APIProviderManager) registerProviders() {
	// 註冊 fal.ai 提供商
	if falConfig, exists := m.config.APIProviderConfigs["fal_ai"]; exists && falConfig.Enabled {
		// 如果是 mock 模式或開發環境，使用 mock 服務
		if m.config.IsMockEnabled() || m.config.Environment == "development" {
			falService := NewFalMockService()
			m.RegisterProvider(interfaces.APIProviderFalAI, falService)
			log.Printf("Registered FalAI Mock service for %s mode", m.config.Mode)
		} else {
			falService := NewFalAIService(falConfig.APIKey, falConfig.BaseURL, time.Duration(falConfig.Timeout)*time.Second)
			m.RegisterProvider(interfaces.APIProviderFalAI, falService)
			log.Printf("Registered FalAI service for production mode")
		}
	}

	// TODO: 註冊其他提供商
	// 註冊 Runway ML 提供商
	// if runwayConfig, exists := m.config.APIProviderConfigs["runway_ml"]; exists && runwayConfig.Enabled {
	//     runwayService := NewRunwayMLService(runwayConfig.APIKey, runwayConfig.BaseURL, time.Duration(runwayConfig.Timeout)*time.Second)
	//     m.RegisterProvider(interfaces.APIProviderRunwayML, runwayService)
	// }

	// 註冊 Stability AI 提供商
	// if stabilityConfig, exists := m.config.APIProviderConfigs["stability_ai"]; exists && stabilityConfig.Enabled {
	//     stabilityService := NewStabilityAIService(stabilityConfig.APIKey, stabilityConfig.BaseURL, time.Duration(stabilityConfig.Timeout)*time.Second)
	//     m.RegisterProvider(interfaces.APIProviderStability, stabilityService)
	// }
}

// RegisterProvider 註冊API提供商
func (m *APIProviderManager) RegisterProvider(provider interfaces.APIProvider, service interfaces.VideoGenerationInterface) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}
	
	if err := service.ValidateConfig(); err != nil {
		return fmt.Errorf("invalid provider config: %w", err)
	}
	
	m.providers[provider] = service
	log.Printf("Registered API provider: %s", provider)
	return nil
}

// SwitchProvider 切換API提供商
func (m *APIProviderManager) SwitchProvider(provider interfaces.APIProvider) error {
	if _, exists := m.providers[provider]; !exists {
		return fmt.Errorf("provider %s is not registered", provider)
	}
	
	m.current = provider
	log.Printf("Switched to API provider: %s", provider)
	return nil
}

// GetCurrentProvider 獲取當前API提供商
func (m *APIProviderManager) GetCurrentProvider() interfaces.APIProvider {
	return m.current
}

// GetCurrentService 獲取當前服務
func (m *APIProviderManager) GetCurrentService() (interfaces.VideoGenerationInterface, error) {
	service, exists := m.providers[m.current]
	if !exists {
		return nil, fmt.Errorf("current provider %s is not available", m.current)
	}
	return service, nil
}

// ListProviders 列出所有已註冊的提供商
func (m *APIProviderManager) ListProviders() []interfaces.APIProvider {
	providers := make([]interfaces.APIProvider, 0, len(m.providers))
	for provider := range m.providers {
		providers = append(providers, provider)
	}
	return providers
}

// SubmitTask 使用當前提供商提交任務
func (m *APIProviderManager) SubmitTask(request *models.TaskCreateRequest) (*models.APIResponse, error) {
	service, err := m.GetCurrentService()
	if err != nil {
		return nil, err
	}
	
	log.Printf("Submitting task using provider: %s", m.current)
	return service.SubmitTask(request)
}

// GetTaskStatus 使用當前提供商查詢任務狀態
func (m *APIProviderManager) GetTaskStatus(requestID string) (string, error) {
	service, err := m.GetCurrentService()
	if err != nil {
		return "", err
	}
	
	return service.GetTaskStatus(requestID)
}

// GetTaskResult 使用當前提供商獲取任務結果
func (m *APIProviderManager) GetTaskResult(requestID string) (*models.APIResult, error) {
	service, err := m.GetCurrentService()
	if err != nil {
		return nil, err
	}
	
	return service.GetTaskResult(requestID)
}

// RetryWithBackoff 使用當前提供商執行重試
func (m *APIProviderManager) RetryWithBackoff(operation func() error, maxRetries int) error {
	service, err := m.GetCurrentService()
	if err != nil {
		return err
	}
	
	return service.RetryWithBackoff(operation, maxRetries)
}

// ResetToDefault 重置為默認提供商
func (m *APIProviderManager) ResetToDefault() error {
	return m.SwitchProvider(m.defaultProvider)
}

// SetDefault 設置默認提供商
func (m *APIProviderManager) SetDefault(provider interfaces.APIProvider) error {
	if _, exists := m.providers[provider]; !exists {
		return fmt.Errorf("provider %s is not registered", provider)
	}
	
	m.defaultProvider = provider
	log.Printf("Set default API provider: %s", provider)
	return nil
}

// GetProviderStatus 獲取提供商狀態
func (m *APIProviderManager) GetProviderStatus() map[interfaces.APIProvider]bool {
	status := make(map[interfaces.APIProvider]bool)
	
	for provider, service := range m.providers {
		// 簡單的健康檢查
		err := service.ValidateConfig()
		status[provider] = err == nil
	}
	
	return status
}