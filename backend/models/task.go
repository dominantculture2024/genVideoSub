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

// Task 任務結構
type Task struct {
	ID               int        `json:"id"`
	ImagePath        string     `json:"image_path"`
	SubtitleColor    string     `json:"subtitle_color"`
	SubtitlePosition string     `json:"subtitle_position"`
	Subtitle         string     `json:"subtitle"`
	Prompt           string     `json:"prompt"`
	Status           TaskStatus `json:"status"`
	ExternalID       string     `json:"external_id"`
	ResultFilePath   string     `json:"result_file_path"`
	CreatedAt        time.Time  `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at"`
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

// TaskCreateRequest 創建任務請求結構
type TaskCreateRequest struct {
	ID               int    `json:"id" binding:"required"`
	ImagePath        string `json:"image_path" binding:"required"`
	SubtitleColor    string `json:"subtitle_color" binding:"required"`
	SubtitlePosition string `json:"subtitle_position" binding:"required"`
	Subtitle         string `json:"subtitle" binding:"required"`
	Prompt           string `json:"prompt" binding:"required"`
}

// TaskResponse 任務響應結構
type TaskResponse struct {
	Status     TaskStatus `json:"status"`
	ExternalID string     `json:"external_id"`
}

// TaskStatusResponse 任務狀態響應結構
type TaskStatusResponse struct {
	ID          int        `json:"id"`
	Status      TaskStatus `json:"status"`
	ExternalID  string     `json:"external_id"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// TaskCompletedRequest 任務完成請求結構
type TaskCompletedRequest struct {
	ExternalTaskID string `form:"external_task_id"`
}

// TaskCompletedResponse 任務完成響應結構
type TaskCompletedResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		FilePath string `json:"file_path"`
		TaskID   int    `json:"task_id"`
	} `json:"data"`
}