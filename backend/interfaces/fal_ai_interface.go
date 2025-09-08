package interfaces

import (
	"genVideoSub/models"
)

// FalAIInterface 定義fal.ai服務的接口
type FalAIInterface interface {
	// SubmitTask 提交任務到fal.ai
	SubmitTask(request *models.TaskCreateRequest) (*models.FalAIResponse, error)
	
	// GetTaskStatus 查詢任務狀態
	GetTaskStatus(requestID string) (string, error)
	
	// GetTaskResult 獲取任務結果
	GetTaskResult(requestID string) (*models.FalAIResult, error)
	
	// RetryWithBackoff 帶指數退避的重試機制
	RetryWithBackoff(operation func() error, maxRetries int) error
}