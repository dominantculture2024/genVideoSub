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
	"github.com/sirupsen/logrus"
)

// FalAIService fal.ai API服務
type FalAIService struct {
	apiKey  string
	baseURL string
	timeout time.Duration
	client  *http.Client
}

// 確保FalAIService實現FalAIInterface接口
var _ interfaces.FalAIInterface = (*FalAIService)(nil)

// NewFalAIService 創建新的fal.ai服務實例
func NewFalAIService(apiKey, baseURL string, timeout time.Duration) interfaces.FalAIInterface {
	return &FalAIService{
		apiKey:  apiKey,
		baseURL: baseURL,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// SubmitTask 提交任務到fal.ai
func (f *FalAIService) SubmitTask(request *models.TaskCreateRequest) (*models.FalAIResponse, error) {
	// 設置默認值
	if request.Duration == "" {
		request.Duration = "5"
	}
	if request.AspectRatio == "" {
		request.AspectRatio = "16:9"
	}
	if request.NegativePrompt == "" {
		request.NegativePrompt = "blur, distort, and low quality"
	}
	if request.CfgScale == 0 {
		request.CfgScale = 0.5
	}

	// 構建fal.ai請求
	falRequest := &models.FalAIRequest{
		Input: models.FalAIInput{
			Prompt:         request.Prompt,
			ImageURL:       request.ImageURL,
			Duration:       request.Duration,
			AspectRatio:    request.AspectRatio,
			NegativePrompt: request.NegativePrompt,
			CfgScale:       request.CfgScale,
		},
	}

	// 序列化請求
	jsonData, err := json.Marshal(falRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 創建HTTP請求
	url := fmt.Sprintf("%s/fal-ai/kling-video/v1.6/pro/image-to-video", f.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 設置請求頭
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", f.apiKey))

	// 發送請求
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 讀取響應
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 檢查HTTP狀態碼
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析響應
	var falResponse models.FalAIResponse
	if err := json.Unmarshal(body, &falResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	logrus.Infof("Task submitted to fal.ai with request_id: %s", falResponse.RequestID)
	return &falResponse, nil
}

// GetTaskStatus 查詢任務狀態
func (f *FalAIService) GetTaskStatus(requestID string) (string, error) {
	url := fmt.Sprintf("%s/fal-ai/queue/requests/%s/status", f.baseURL, requestID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 設置請求頭
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", f.apiKey))

	// 發送請求
	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 讀取響應
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 檢查HTTP狀態碼
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析狀態響應
	var statusResp map[string]interface{}
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	status, ok := statusResp["status"].(string)
	if !ok {
		return "", fmt.Errorf("invalid status response format")
	}

	return status, nil
}

// GetTaskResult 獲取任務結果
func (f *FalAIService) GetTaskResult(requestID string) (*models.FalAIResult, error) {
	url := fmt.Sprintf("%s/fal-ai/queue/requests/%s", f.baseURL, requestID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 設置請求頭
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", f.apiKey))

	// 發送請求
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 讀取響應
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 檢查HTTP狀態碼
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析結果響應
	var result models.FalAIResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// RetryWithBackoff 帶指數退避的重試機制
func (f *FalAIService) RetryWithBackoff(operation func() error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		logrus.Warnf("Operation failed (attempt %d/%d): %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			// 指數退避：2^i 秒
			backoff := time.Duration(1<<uint(i)) * time.Second
			logrus.Infof("Retrying in %v...", backoff)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}