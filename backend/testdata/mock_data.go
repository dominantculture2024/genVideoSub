package testdata

import (
	"fmt"
	"math/rand"
	"time"

	"genVideoSub/models"
)

// MockTaskRequests 預定義的mock任務請求
var MockTaskRequests = map[string]*models.TaskCreateRequest{
	"basic_request": {
		Prompt:         "A beautiful sunset over the ocean",
		ImageURL:       "https://example.com/sunset.jpg",
		Duration:       "5",
		AspectRatio:    "16:9",
		NegativePrompt: "blur, distort, and low quality",
		CfgScale:       0.5,
	},
	"complex_request": {
		Prompt:         "A futuristic city with flying cars and neon lights",
		ImageURL:       "https://example.com/city.jpg",
		Duration:       "10",
		AspectRatio:    "9:16",
		NegativePrompt: "low quality, blurry, distorted",
		CfgScale:       0.8,
	},
	"minimal_request": {
		Prompt:   "Simple animation",
		ImageURL: "https://example.com/simple.jpg",
	},
}

// MockFalAIResponses 預定義的fal.ai響應
var MockFalAIResponses = map[string]*models.FalAIResponse{
	"success_response": {
		RequestID: "mock_12345_6789",
	},
	"another_response": {
		RequestID: "mock_98765_4321",
	},
}

// MockTaskStatuses 預定義的任務狀態
var MockTaskStatuses = []string{
	"IN_QUEUE",
	"IN_PROGRESS",
	"COMPLETED",
	"FAILED",
}

// MockFalAIResults 預定義的任務結果
var MockFalAIResults = map[string]*models.FalAIResult{
	"completed_result": {
		Video: models.VideoResult{
			URL:      "https://mock-storage.example.com/videos/mock_12345_6789.mp4",
			Width:    1920,
			Height:   1080,
			Duration: 5.0,
		},
		Seed:           123456789,
		Timings:        map[string]interface{}{"inference": 25.5, "total": 30.2},
		HasNsfwConcepts: []bool{false},
		Prompt:         "A beautiful sunset over the ocean",
	},
	"hd_result": {
		Video: models.VideoResult{
			URL:      "https://mock-storage.example.com/videos/mock_98765_4321.mp4",
			Width:    3840,
			Height:   2160,
			Duration: 10.0,
		},
		Seed:           987654321,
		Timings:        map[string]interface{}{"inference": 45.8, "total": 52.1},
		HasNsfwConcepts: []bool{false},
		Prompt:         "A futuristic city with flying cars and neon lights",
	},
}

// MockErrorResponses 預定義的錯誤響應
var MockErrorResponses = map[string]error{
	"task_not_found":    fmt.Errorf("task not found: mock_invalid_id"),
	"api_rate_limit":    fmt.Errorf("API rate limit exceeded"),
	"invalid_request":   fmt.Errorf("invalid request parameters"),
	"service_unavailable": fmt.Errorf("service temporarily unavailable"),
}

// TestScenarios 測試場景定義
type TestScenario struct {
	Name        string
	Description string
	Request     *models.TaskCreateRequest
	ExpectedResponse *models.FalAIResponse
	ExpectedError    error
	StatusSequence   []string
	FinalResult     *models.FalAIResult
}

// MockTestScenarios 預定義的測試場景
var MockTestScenarios = []TestScenario{
	{
		Name:        "successful_video_generation",
		Description: "成功的視頻生成流程",
		Request:     MockTaskRequests["basic_request"],
		ExpectedResponse: MockFalAIResponses["success_response"],
		ExpectedError:    nil,
		StatusSequence:   []string{"IN_QUEUE", "IN_PROGRESS", "COMPLETED"},
		FinalResult:     MockFalAIResults["completed_result"],
	},
	{
		Name:        "complex_video_generation",
		Description: "複雜視頻生成流程",
		Request:     MockTaskRequests["complex_request"],
		ExpectedResponse: MockFalAIResponses["another_response"],
		ExpectedError:    nil,
		StatusSequence:   []string{"IN_QUEUE", "IN_PROGRESS", "COMPLETED"},
		FinalResult:     MockFalAIResults["hd_result"],
	},
	{
		Name:        "task_not_found_error",
		Description: "任務未找到錯誤",
		Request:     MockTaskRequests["minimal_request"],
		ExpectedResponse: nil,
		ExpectedError:    MockErrorResponses["task_not_found"],
		StatusSequence:   []string{},
		FinalResult:     nil,
	},
}

// GetMockTaskByID 根據ID獲取mock任務
func GetMockTaskByID(requestID string) *TestScenario {
	for _, scenario := range MockTestScenarios {
		if scenario.ExpectedResponse != nil && scenario.ExpectedResponse.RequestID == requestID {
			return &scenario
		}
	}
	return nil
}

// GenerateRandomRequestID 生成隨機請求ID
func GenerateRandomRequestID() string {
	return fmt.Sprintf("mock_%d_%d", time.Now().Unix(), rand.Intn(10000))
}

// CreateTestRequest 創建測試請求
func CreateTestRequest(prompt, imageURL string) *models.TaskCreateRequest {
	return &models.TaskCreateRequest{
		Prompt:         prompt,
		ImageURL:       imageURL,
		Duration:       "5",
		AspectRatio:    "16:9",