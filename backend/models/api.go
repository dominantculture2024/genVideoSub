package models

import (
	"time"
)

// APIResponse 通用API響應結構
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIResult 通用API結果結構
type APIResult struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// NewAPIResponse 創建新的API響應
func NewAPIResponse(success bool, message string, data interface{}) *APIResponse {
	return &APIResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewAPIError 創建錯誤響應
func NewAPIError(message string, err error) *APIResponse {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	return &APIResponse{
		Success:   false,
		Message:   message,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}
}