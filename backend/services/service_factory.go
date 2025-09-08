package services

import (
	"fmt"
	"log"

	"genVideoSub/config"
)

// ServiceFactory 服務工廠
type ServiceFactory struct {
	config *config.TestConfig
}

// NewServiceFactory 創建服務工廠
func NewServiceFactory(cfg *config.TestConfig) *ServiceFactory {
	return &ServiceFactory{
		config: cfg,
	}
}

// CreateFalAIService 創建FalAI服務實例
func (f *ServiceFactory) CreateFalAIService() (FalAIInterface, error) {
	if f.config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	
	// 驗證配置
	if err := f.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	// 根據配置決定使用哪種服務
	if f.config.IsMockEnabled() {
		log.Printf("Creating FalAI Mock Service (Environment: %s)", f.config.GetEnvironment())
		return NewFalMockService(), nil
	} else {
		log.Printf("Creating FalAI Real Service (Environment: %s)", f.config.GetEnvironment())
		return NewFalAIService(f.config.FalAPIKey), nil
	}
}

// GetServiceType 獲取服務類型
func (f *ServiceFactory) GetServiceType() string {
	if f.config.IsMockEnabled() {
		return "mock"
	}
	return "real"
}

// GetEnvironment 獲取環境信息
func (f *ServiceFactory) GetEnvironment() string {
	return f.config.GetEnvironment()
}

// IsTestMode 檢查是否為測試模式
func (f *ServiceFactory) IsTestMode() bool {
	return f.config.IsTestMode()
}

// GetConfig 獲取配置
func (f *ServiceFactory) GetConfig() *config.TestConfig {
	return f.config
}

// DefaultServiceFactory 創建默認服務工廠
func DefaultServiceFactory() *ServiceFactory {
	config := config.LoadTestConfig()
	return NewServiceFactory(config)
}

// CreateFalAIServiceWithConfig 使用指定配置創建FalAI服務
func CreateFalAIServiceWithConfig(useMock bool, apiKey string) (FalAIInterface, error) {
	if useMock {
		log.Println("Creating FalAI Mock Service")
		return NewFalMockService(), nil
	}
	
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for real service")
	}
	
	log.Println("Creating FalAI Real Service")
	return NewFalAIService(apiKey), nil
}

// CreateTestService 創建測試專用服務
func CreateTestService() FalAIInterface {
	return NewFalMockService()
}

// CreateProductionService 創建生產環境服務
func CreateProductionService(apiKey string) (FalAIInterface, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for production service")
	}
	return NewFalAIService(apiKey), nil
}