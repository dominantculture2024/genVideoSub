package interfaces

import (
	"genVideoSub/models"
)

// APIProvider API提供商類型
type APIProvider string

const (
	APIProviderFalAI     APIProvider = "fal_ai"
	APIProviderRunwayML  APIProvider = "runway_ml"
	APIProviderStability APIProvider = "stability_ai"
	APIProviderCustom    APIProvider = "custom"
)

// VideoGenerationInterface 視頻生成服務的通用接口
type VideoGenerationInterface interface {
	// SubmitTask 提交任務
	SubmitTask(request *models.TaskCreateRequest) (*models.APIResponse, error)
	
	// GetTaskStatus 查詢任務狀態
	GetTaskStatus(requestID string) (string, error)
	
	// GetTaskResult 獲取任務結果
	GetTaskResult(requestID string) (*models.APIResult, error)
	
	// GetProviderName 獲取提供商名稱
	GetProviderName() APIProvider
	
	// ValidateConfig 驗證配置
	ValidateConfig() error
	
	// RetryWithBackoff 帶指數退避的重試機制
	RetryWithBackoff(operation func() error, maxRetries int) error
}

// APIProviderConfig API提供商配置
type APIProviderConfig struct {
	Provider APIProvider          `json:"provider"`
	APIKey   string               `json:"api_key"`
	BaseURL  string               `json:"base_url"`
	Timeout  int                  `json:"timeout"`
	Headers  map[string]string    `json:"headers,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

// CurlRequest Curl請求結構
type CurlRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Timeout int               `json:"timeout"`
}

// CurlResponse Curl響應結構
type CurlResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Error      string            `json:"error,omitempty"`
}

// CurlInterface Curl功能接口
type CurlInterface interface {
	// ExecuteCurl 執行curl請求
	ExecuteCurl(request *CurlRequest) (*CurlResponse, error)
	
	// GenerateCurlCommand 生成curl命令
	GenerateCurlCommand(request *CurlRequest) string
	
	// ParseCurlCommand 解析curl命令
	ParseCurlCommand(command string) (*CurlRequest, error)
}