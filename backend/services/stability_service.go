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

// StabilityAIService 實現 VideoGenerationInterface 接口
type StabilityAIService struct {
	apiKey  string
	baseURL string
	timeout time.Duration
	client  *http.Client
}

// NewStabilityAIService 創建新的 Stability AI 服務實例
func NewStabilityAIService(apiKey, baseURL string, timeout time.Duration) *StabilityAIService {
	return &StabilityAIService{
		apiKey:  apiKey,
		baseURL: baseURL,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// SubmitTask 提交視頻生成任務到 Stability AI
func (s *StabilityAIService) SubmitTask(request *models.TaskCreateRequest) (*models.APIResponse, error) {
	// 轉換請求格式為 Stability AI 格式
	stabilityRequest := map[string]interface{}{
		"image": request.ImageURL,
		"cfg_scale": 1.8,
		"motion_bucket_id": 127,
		"seed": 0,
	}

	// 如果有提示詞，添加到請求中
	if request.Prompt != "" {
		stabilityRequest["prompt"] = request.Prompt
	}

	// 序列化請求
	requestBody, err := json.Marshal(stabilityRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 創建 HTTP 請求
	url := fmt.Sprintf("%s/v2beta/image-to-video", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 設置請求頭
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Accept", "application/json")

	// 發送請求
	resp, err := s.client.Do(req)
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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析響應
	var stabilityResponse map[string]interface{}
	if err := json.Unmarshal(body, &stabilityResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 轉換為標準格式
	requestID := fmt.Sprintf("%v", stabilityResponse["id"])
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
func (s *StabilityAIService) GetTaskStatus(requestID string) (string, error) {
	url := fmt.Sprintf("%s/v2beta/image-to-video/result/%s", s.baseURL, requestID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 處理不同的狀態碼
	var status string
	switch resp.StatusCode {
	case http.StatusOK:
		status = "COMPLETED"
	case http.StatusAccepted:
		status = "IN_PROGRESS"
	case http.StatusNotFound:
		return "", fmt.Errorf("task not found: %s", requestID)
	default:
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return status, nil
}

// GetTaskResult 獲取任務結果
func (s *StabilityAIService) GetTaskResult(requestID string) (*models.APIResult, error) {
	status, err := s.GetTaskStatus(requestID)
	if err != nil {
		return nil, err
	}

	if status != "COMPLETED" {
		return nil, fmt.Errorf("task not completed, current status: %s", status)
	}

	// 獲取任務結果
	url := fmt.Sprintf("%s/v2beta/image-to-video/result/%s", s.baseURL, requestID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
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

	// 解析響應
	var stabilityResponse map[string]interface{}
	if err := json.Unmarshal(body, &stabilityResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 提取視頻URL
	var videoURL string
	if video, ok := stabilityResponse["video"].(string); ok {
		videoURL = video
	} else if artifacts, ok := stabilityResponse["artifacts"].([]interface{}); ok && len(artifacts) > 0 {
		if artifact, ok := artifacts[0].(map[string]interface{}); ok {
			if url, ok := artifact["url"].(string); ok {
				videoURL = url
			}
		}
	}

	if videoURL == "" {
		return nil, fmt.Errorf("no video URL found in response")
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
func (s *StabilityAIService) GetProviderName() interfaces.APIProvider {
	return interfaces.APIProviderStability
}

// ValidateConfig 驗證配置
func (s *StabilityAIService) ValidateConfig() error {
	if s.apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	if s.baseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	return nil
}

// RetryWithBackoff 實現重試機制
func (s *StabilityAIService) RetryWithBackoff(operation func() error, maxRetries int) error {
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