package services

import (
	"fmt"
	"time"

	"genVideoSub/interfaces"
	"genVideoSub/models"
	"genVideoSub/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TaskService 任務管理服務
type TaskService struct {
	storage     storage.Storage
	videoService interfaces.VideoGenerationInterface
	workerNum   int
	taskChan    chan string
	stopChan    chan bool
}

// NewTaskService 創建新的任務服務實例
func NewTaskService(storage storage.Storage, videoService interfaces.VideoGenerationInterface, workerNum int) *TaskService {
	return &TaskService{
		storage:     storage,
		videoService: videoService,
		workerNum:   workerNum,
		taskChan:    make(chan string, 100),
		stopChan:    make(chan bool),
	}
}

// CreateTask 創建新任務
func (ts *TaskService) CreateTask(request *models.TaskCreateRequest) (*models.TaskResponse, error) {
	// 生成任務ID
	taskID := uuid.New().String()

	// 創建任務對象
	task := &models.Task{
		ID:             taskID,
		Prompt:         request.Prompt,
		ImageURL:       request.ImageURL,
		Duration:       request.Duration,
		AspectRatio:    request.AspectRatio,
		NegativePrompt: request.NegativePrompt,
		CfgScale:       request.CfgScale,
		Status:         models.TaskStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存任務到存儲
	if err := ts.storage.SaveTask(task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	// 將任務加入處理隊列
	select {
	case ts.taskChan <- taskID:
		logrus.Infof("Task %s added to processing queue", taskID)
	default:
		logrus.Warnf("Task queue is full, task %s will be processed later", taskID)
	}

	// 返回任務響應
	return &models.TaskResponse{
		ID:        taskID,
		Status:    string(task.Status),
		CreatedAt: task.CreatedAt,
	}, nil
}

// GetTask 獲取任務信息
func (ts *TaskService) GetTask(taskID string) (*models.Task, error) {
	task, err := ts.storage.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// GetTaskStatus 獲取任務狀態
func (ts *TaskService) GetTaskStatus(taskID string) (*models.TaskStatusResponse, error) {
	task, err := ts.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	return &models.TaskStatusResponse{
		ID:        task.ID,
		Status:    string(task.Status),
		Progress:  task.Progress,
		Message:   task.Message,
		UpdatedAt: task.UpdatedAt,
	}, nil
}

// GetTaskResult 獲取任務結果
func (ts *TaskService) GetTaskResult(taskID string) (*models.TaskResultResponse, error) {
	task, err := ts.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	if task.Status != models.TaskStatusCompleted {
		return nil, fmt.Errorf("task is not completed yet")
	}

	return &models.TaskResultResponse{
		ID:        task.ID,
		Status:    string(task.Status),
		VideoURL:  task.VideoURL,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}, nil
}

// ListTasks 列出所有任務
func (ts *TaskService) ListTasks(limit, offset int) ([]*models.Task, error) {
	tasks, err := ts.storage.ListTasks(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

// UpdateTaskStatus 更新任務狀態
func (ts *TaskService) UpdateTaskStatus(taskID string, status models.TaskStatus, message string, progress int) error {
	task, err := ts.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Status = status
	task.Message = message
	task.Progress = progress
	task.UpdatedAt = time.Now()

	if err := ts.storage.SaveTask(task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	logrus.Infof("Task %s status updated to %s", taskID, status)
	return nil
}

// StartWorkers 啟動工作協程池
func (ts *TaskService) StartWorkers() {
	logrus.Infof("Starting %d task workers", ts.workerNum)

	for i := 0; i < ts.workerNum; i++ {
		go ts.worker(i)
	}
}

// StopWorkers 停止工作協程池
func (ts *TaskService) StopWorkers() {
	logrus.Info("Stopping task workers")
	for i := 0; i < ts.workerNum; i++ {
		ts.stopChan <- true
	}
}

// worker 工作協程
func (ts *TaskService) worker(workerID int) {
	logrus.Infof("Worker %d started", workerID)

	for {
		select {
		case taskID := <-ts.taskChan:
			logrus.Infof("Worker %d processing task %s", workerID, taskID)
			ts.processTask(taskID)
		case <-ts.stopChan:
			logrus.Infof("Worker %d stopped", workerID)
			return
		}
	}
}

// processTask 處理單個任務
func (ts *TaskService) processTask(taskID string) {
	// 更新任務狀態為處理中
	if err := ts.UpdateTaskStatus(taskID, models.TaskStatusProcessing, "Submitting to fal.ai", 10); err != nil {
		logrus.Errorf("Failed to update task status: %v", err)
		return
	}

	// 獲取任務詳情
	task, err := ts.GetTask(taskID)
	if err != nil {
		logrus.Errorf("Failed to get task %s: %v", taskID, err)
		ts.UpdateTaskStatus(taskID, models.TaskStatusFailed, fmt.Sprintf("Failed to get task: %v", err), 0)
		return
	}

	// 構建fal.ai請求
	request := &models.TaskCreateRequest{
		Prompt:         task.Prompt,
		ImageURL:       task.ImageURL,
		Duration:       task.Duration,
		AspectRatio:    task.AspectRatio,
		NegativePrompt: task.NegativePrompt,
		CfgScale:       task.CfgScale,
	}

	// 提交任務到視頻生成服務
	var falResponse *models.APIResponse
	err = ts.videoService.RetryWithBackoff(func() error {
		var submitErr error
		falResponse, submitErr = ts.videoService.SubmitTask(request)
		return submitErr
	}, 3)

	if err != nil {
		logrus.Errorf("Failed to submit task %s to fal.ai: %v", taskID, err)
		ts.UpdateTaskStatus(taskID, models.TaskStatusFailed, fmt.Sprintf("Failed to submit: %v", err), 0)
		return
	}

	// 保存請求ID
	task.RequestID = falResponse.RequestID
	if err := ts.storage.SaveTask(task); err != nil {
		logrus.Errorf("Failed to save request ID for task %s: %v", taskID, err)
	}

	// 更新狀態為等待處理
	ts.UpdateTaskStatus(taskID, models.TaskStatusProcessing, "Waiting for fal.ai processing", 30)

	// 輪詢任務狀態
	ts.pollTaskStatus(taskID, falResponse.RequestID)
}

// pollTaskStatus 輪詢任務狀態
func (ts *TaskService) pollTaskStatus(taskID, requestID string) {
	maxPolls := 60 // 最多輪詢60次（10分鐘）
	pollInterval := 10 * time.Second

	for i := 0; i < maxPolls; i++ {
		time.Sleep(pollInterval)

		// 查詢視頻生成服務狀態
		status, err := ts.videoService.GetTaskStatus(requestID)
		if err != nil {
			logrus.Errorf("Failed to get status for task %s: %v", taskID, err)
			continue
		}

		logrus.Infof("Task %s status: %s (poll %d/%d)", taskID, status, i+1, maxPolls)

		switch status {
		case "COMPLETED":
			// 獲取結果
			result, err := ts.videoService.GetTaskResult(requestID)
			if err != nil {
				logrus.Errorf("Failed to get result for task %s: %v", taskID, err)
				ts.UpdateTaskStatus(taskID, models.TaskStatusFailed, fmt.Sprintf("Failed to get result: %v", err), 90)
				return
			}

			// 更新任務結果
			task, _ := ts.GetTask(taskID)
			if videoData, ok := result.Result.(map[string]interface{}); ok {
				if videoURL, exists := videoData["video_url"]; exists {
					task.VideoURL = videoURL.(string)
				}
			}
			if err := ts.storage.SaveTask(task); err != nil {
				logrus.Errorf("Failed to save task result for %s: %v", taskID, err)
			}

			// 標記為完成
			ts.UpdateTaskStatus(taskID, models.TaskStatusCompleted, "Video generation completed", 100)
			return

		case "FAILED":
			ts.UpdateTaskStatus(taskID, models.TaskStatusFailed, "fal.ai processing failed", 0)
			return

		case "IN_PROGRESS", "IN_QUEUE":
			// 更新進度
			progress := 30 + (i * 60 / maxPolls)
			ts.UpdateTaskStatus(taskID, models.TaskStatusProcessing, fmt.Sprintf("Processing... (%s)", status), progress)
			continue

		default:
			logrus.Warnf("Unknown status for task %s: %s", taskID, status)
			continue
		}
	}

	// 超時處理
	logrus.Errorf("Task %s polling timeout after %d attempts", taskID, maxPolls)
	ts.UpdateTaskStatus(taskID, models.TaskStatusFailed, "Processing timeout", 0)
}