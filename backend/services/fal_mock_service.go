package services

import (
	"fmt"
	"math/rand"
	"time"

	"genVideoSub/interfaces"
	"genVideoSub/models"
	"github.com/sirupsen/logrus"
)

// FalMockService fal.ai API的mock實現
type FalMockService struct {
	tasks map[string]*MockTask // 存儲mock任務
}

// MockTask mock任務結構
type MockTask struct {
	RequestID   string
	Status      string
	CreatedAt   time.Time
	CompletedAt *time.Time
	Request     *models.TaskCreateRequest
	Result      *models.FalAIResult
}

// 確保FalMockService實現FalAIInterface接口
var _ interfaces.FalAIInterface = (*FalMockService)(nil)

// NewFalMockService 創建新的fal.ai mock服務實例
func NewFalMockService() interfaces.FalAIInterface {
	return &FalMockService{
		tasks: make(map[string]*MockTask),
	}
}

// SubmitTask 模擬提交任務到fal.ai
func (m *FalMockService) SubmitTask(request *models.TaskCreateRequest) (*models.FalAIResponse, error) {
	// 生成mock request ID
	requestID := fmt.Sprintf("mock_%d_%d", time.Now().Unix(), rand.Intn(10000))
	
	// 創建mock任務
	task := &MockTask{
		RequestID: requestID,
		Status:    "IN_QUEUE",
		CreatedAt: time.Now(),
		Request:   request,
	}
	
	// 存儲任務
	m.tasks[requestID] = task
	
	logrus.Infof("Mock task submitted with request_id: %s", requestID)
	
	return &models.FalAIResponse{
		RequestID: requestID,
	}, nil
}

// GetTaskStatus 模擬查詢任務狀態
func (m *FalMockService) GetTaskStatus(requestID string) (string, error) {
	task, exists := m.tasks[requestID]
	if !exists {
		return "", fmt.Errorf("task not found: %s", requestID)
	}
	
	// 模擬狀態轉換邏輯
	m.simulateStatusTransition(task)
	
	return task.Status, nil
}

// GetTaskResult 模擬獲取任務結果
func (m *FalMockService) GetTaskResult(requestID string) (*models.FalAIResult, error) {
	task, exists := m.tasks[requestID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", requestID)
	}
	
	// 模擬狀態轉換
	m.simulateStatusTransition(task)
	
	// 如果任務未完成，返回錯誤
	if task.Status != "COMPLETED" {
		return nil, fmt.Errorf("task not completed yet, current status: %s", task.Status)
	}
	
	// 如果結果還未生成，創建mock結果
	if task.Result == nil {
		task.Result = m.generateMockResult(task)
	}
	
	return task.Result, nil
}

// RetryWithBackoff 模擬重試機制
func (m *FalMockService) RetryWithBackoff(operation func() error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}
		
		logrus.Warnf("Mock operation failed (attempt %d/%d): %v", i+1, maxRetries, err)
		
		if i < maxRetries-1 {
			// 模擬指數退避
			backoff := time.Duration(1<<uint(i)) * time.Second
			logrus.Infof("Mock retrying in %v...", backoff)
			time.Sleep(backoff)
		}
	}
	
	return fmt.Errorf("mock operation failed after %d retries: %w", maxRetries, err)
}

// simulateStatusTransition 模擬任務狀態轉換
func (m *FalMockService) simulateStatusTransition(task *MockTask) {
	now := time.Now()
	elapsed := now.Sub(task.CreatedAt)
	
	switch task.Status {
	case "IN_QUEUE":
		// 5秒後轉為IN_PROGRESS
		if elapsed > 5*time.Second {
			task.Status = "IN_PROGRESS"
			logrus.Infof("Mock task %s status changed to IN_PROGRESS", task.RequestID)
		}
	case "IN_PROGRESS":
		// 30秒後轉為COMPLETED
		if elapsed > 30*time.Second {
			task.Status = "COMPLETED"
			completedAt := now
			task.CompletedAt = &completedAt
			logrus.Infof("Mock task %s status changed to COMPLETED", task.RequestID)
		}
	}
}

// generateMockResult 生成mock結果
func (m *FalMockService) generateMockResult(task *MockTask) *models.FalAIResult {
	// 生成mock視頻URL
	mockVideoURL := fmt.Sprintf("https://mock-storage.example.com/videos/%s.mp4", task.RequestID)
	
	return &models.FalAIResult{
		Video: models.VideoResult{
			URL:      mockVideoURL,
			Width:    1920,
			Height:   1080,
			Duration: 5.0,
		},
		Seed:           rand.Int63(),
		Timings:        map[string]interface{}{"inference": 25.5, "total": 30.2},
		HasNsfwConcepts: []bool{false},
		Prompt:         task.Request.Prompt,
	}
}