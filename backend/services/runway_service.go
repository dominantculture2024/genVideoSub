package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"genVideoSub/interfaces"
	"genVideoSub/models"
)

// RunwayMLService 實現 VideoGenerationInterface 接口
type RunwayMLService struct {
	apiKey  string
	baseURL string
	timeout time.Duration
	client  *http.Client
}

// NewRunwayMLService 創建新的 Runway ML 服務實例
func NewRunwayMLService(apiKey, baseURL string, timeout time.Duration) *RunwayMLService {
	return &RunwayMLService{
		apiKey:  apiKey,
		baseURL: baseURL,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// SubmitTask 提交視頻生成任務到 Runway ML
func (r *RunwayMLService) SubmitTask(request *models.TaskCreateRequest) (*models.APIResponse, error) {
	// 轉換請求格式為 Runway ML 格式
	runwayRequest := map[string]interface{}{
		"text_prompt": request.Prompt,
		"image_prompt": request.ImageURL,
		"model": "gen3a_turbo", // Runway ML 的模型
		"aspect_ratio": "16:9",
		"duration": 5,
	}

	// 序列化請求
	requestBody, err := json.Marshal(runwayRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 創建 HTTP 請求
	url := fmt.Sprintf("%s/v1/image_to_video", r.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 設置請求頭
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
	req.Header.Set("X-Runway-Version", "2024-09-13")

	// 發送請求
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 讀取響應
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 檢查 HTTP 狀態碼
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析響應
	var runwayResponse map[string]interface{}
	if err := json.Unmarshal(body, &runwayResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 轉換為標準格式
	requestID := fmt.Sprintf("%v", runwayResponse["id"])
	response := &models.APIResponse{
		Success:   true,
		Message:   "Task submitted successfully",
		Data:      map[string]interface{}{"request_id": requestID},
		RequestID: requestID,
		Timestamp: time.Now(),
	}

	return response, nil
}

// GetTaskStatus 獲取任務狀態
func (r *RunwayMLService) GetTaskStatus(requestID string) (string, error) {
	url := fmt.Sprintf("%s/v1/tasks/%s", r.baseURL, requestID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
	req.Header.Set("X-Runway-Version", "2024-09-13")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var runwayResponse map[string]interface{}
	if err := json.Unmarshal(body, &runwayResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 轉換狀態
	status := "IN_QUEUE"
	if runwayStatus, ok := runwayResponse["status"].(string); ok {
		switch runwayStatus {
		case "PENDING":
			status = "IN_QUEUE"
		case "RUNNING":
			status = "IN_PROGRESS"
		case "SUCCEEDED":
			status = "COMPLETED"
		case "FAILED":
			status = "FAILED"
		}
	}

	return status, nil
}

// GetTaskResult 獲取任務結果
func (r *RunwayMLService) GetTaskResult(requestID string) (*models.APIResult, error) {
	status, err := r.GetTaskStatus(requestID)
	if err != nil {
		return nil, err
	}

	if status != "COMPLETED" {
		return nil, fmt.Errorf("task not completed, current status: %s", status)
	}

	// 獲取完整的任務信息
	url := fmt.Sprintf("%s/v1/tasks/%s", r.baseURL, requestID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
	req.Header.Set("X-Runway-Version", "2024-09-13")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var runwayResponse map[string]interface{}
	if err := json.Unmarshal(body, &runwayResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 提取視頻URL
	videoURL := ""
	if output, ok := runwayResponse["output"].(map[string]interface{}); ok {
		if video, ok := output["video"].(string); ok {
			videoURL = video
		}
	}

	return &models.APIResult{
		ID:     requestID,
		Status: "completed",
		Result: map[string]interface{}{
			"video_url": videoURL,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// GetProviderName 返回提供商名稱
func (r *RunwayMLService) GetProviderName() interfaces.APIProvider {
	return interfaces.APIProviderRunwayML
}

// ValidateConfig 驗證配置
func (r *RunwayMLService) ValidateConfig() error {
	if r.apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	if r.baseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	return nil
}

// RetryWithBackoff 實現重試機制
func (r *RunwayMLService) RetryWithBackoff(operation func() error, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}