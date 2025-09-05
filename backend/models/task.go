package models

import (
	"time"
)

// TaskStatus 任務狀態枚舉
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// Task 任務結構 - 根據fal.ai API規格調整
type Task struct {
	ID             string     `json:"id"`
	Prompt         string     `json:"prompt"`
	ImageURL       string     `json:"image_url"`
	Duration       string     `json:"duration"`        // "5" or "10"
	AspectRatio    string     `json:"aspect_ratio"`    // "16:9", "9:16", "1:1"
	NegativePrompt string     `json:"negative_prompt"` // 負面提示詞
	CfgScale       float64    `json:"cfg_scale"`       // CFG引導強度
	Status         TaskStatus `json:"status"`
	RequestID      string     `json:"request_id"`      // fal.ai請求ID
	VideoURL       string     `json:"video_url"`       // 生成的視頻URL
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

// TempFile 臨時文件結構
type TempFile struct {
	ID        string    `json:"id"`
	TaskID    int       `json:"task_id"`
	FilePath  string    `json:"file_path"`
	FileType  string    `json:"file_type"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// TaskCreateRequest 創建任務請求結構 - 根據fal.ai API規格
type TaskCreateRequest struct {
	Prompt         string  `json:"prompt" binding:"required"`
	ImageURL       string  `json:"image_url" binding:"required"`
	Duration       string  `json:"duration"`        // 可選，默認"5"
	AspectRatio    string  `json:"aspect_ratio"`    // 可選，默認"16:9"
	NegativePrompt string  `json:"negative_prompt"` // 可選
	CfgScale       float64 `json:"cfg_scale"`       // 可選，默認0.5
}

// TaskResponse 任務響應結構
type TaskResponse struct {
	Status    TaskStatus `json:"status"`
	RequestID string     `json:"request_id"`
	TaskID    string     `json:"task_id"`
}

// TaskStatusResponse 任務狀態響應結構
type TaskStatusResponse struct {
	ID          string     `json:"id"`
	Status      TaskStatus `json:"status"`
	RequestID   string     `json:"request_id"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// TaskResultResponse 任務結果響應結構
type TaskResultResponse struct {
	Success   bool   `json:"success"`
	VideoURL  string `json:"video_url"`
	RequestID string `json:"request_id"`
	Status    TaskStatus `json:"status"`
}

// FalAIRequest fal.ai API請求結構
type FalAIRequest struct {
	Input FalAIInput `json:"input"`
}

// FalAIInput fal.ai API輸入參數
type FalAIInput struct {
	Prompt         string  `json:"prompt"`
	ImageURL       string  `json:"image_url"`
	Duration       string  `json:"duration"`
	AspectRatio    string  `json:"aspect_ratio"`
	NegativePrompt string  `json:"negative_prompt"`
	CfgScale       float64 `json:"cfg_scale"`
}

// FalAIResponse fal.ai API響應結構
type FalAIResponse struct {
	RequestID string `json:"request_id"`
}

// FalAIResult fal.ai API結果結構
type FalAIResult struct {
	Video struct {
		URL string `json:"url"`
	} `json:"video"`
}