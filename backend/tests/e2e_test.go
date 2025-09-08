package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"genVideoSub/handlers"
	"genVideoSub/services"
	"genVideoSub/testdata"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite 端到端測試套件
type E2ETestSuite struct {
	suite.Suite
	server     *httptest.Server
	client     *http.Client
	mockService *services.FalMockService
	tempDir    string
}

// SetupSuite 設置測試套件
func (suite *E2ETestSuite) SetupSuite() {
	// 創建臨時目錄
	tempDir, err := os.MkdirTemp("", "genVideoSub_e2e_test")
	suite.Require().NoError(err)
	suite.tempDir = tempDir
	
	// 設置測試模式
	gin.SetMode(gin.TestMode)
	
	// 創建mock服務
	suite.mockService = services.NewFalMockService().(*services.FalMockService)
	
	// 創建處理器
	taskHandler := handlers.NewTaskHandler(suite.mockService)
	
	// 設置路由
	router := gin.New()
	router.Use(gin.Recovery())
	
	// 添加CORS中間件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	v1 := router.Group("/api/v1")
	{
		v1.POST("/tasks", taskHandler.CreateTask)
		v1.GET("/tasks/:id/status", taskHandler.GetTaskStatus)
		v1.GET("/tasks/:id/result", taskHandler.GetTaskResult)
	}
	
	// 創建測試服務器
	suite.server = httptest.NewServer(router)
	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}
}

// TearDownSuite 清理測試套件
func (suite *E2ETestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestCompleteVideoGenerationE2E 測試完整的視頻生成端到端流程
func (suite *E2ETestSuite) TestCompleteVideoGenerationE2E() {
	// 步驟1: 提交視頻生成任務
	taskRequest := testdata.MockTaskRequests["basic_request"]
	requestBody, err := json.Marshal(taskRequest)
	suite.Require().NoError(err)
	
	resp, err := suite.client.Post(
		suite.server.URL+"/api/v1/tasks",
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Equal(http.StatusOK, resp.StatusCode)
	
	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	suite.Require().NoError(err)
	
	requestID := createResponse["request_id"].(string)
	suite.NotEmpty(requestID)
	suite.NotEmpty(createResponse["status"])
	
	// 步驟2: 輪詢任務狀態直到完成
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	
	finalStatus := suite.pollTaskStatus(ctx, requestID)
	suite.Equal("COMPLETED", finalStatus)
	
	// 步驟3: 獲取任務結果
	resp, err = suite.client.Get(suite.server.URL + "/api/v1/tasks/" + requestID + "/result")
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Equal(http.StatusOK, resp.StatusCode)
	
	var resultResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&resultResponse)
	suite.Require().NoError(err)
	
	// 驗證結果結構
	suite.Contains(resultResponse, "video")
	suite.Contains(resultResponse, "seed")
	suite.Contains(resultResponse, "timings")
	
	videoData := resultResponse["video"].(map[string]interface{})
	suite.Contains(videoData, "url")
	suite.Contains(videoData, "width")
	suite.Contains(videoData, "height")
	suite.Contains(videoData, "content_type")
}

// TestMultipleTasksE2E 測試多任務端到端流程
func (suite *E2ETestSuite) TestMultipleTasksE2E() {
	taskTypes := []string{"basic_request", "complex_request", "minimal_request"}
	requestIDs := make([]string, len(taskTypes))
	
	// 並發提交多個任務
	for i, taskType := range taskTypes {
		taskRequest := testdata.MockTaskRequests[taskType]
		requestBody, err := json.Marshal(taskRequest)
		suite.Require().NoError(err)
		
		resp, err := suite.client.Post(
			suite.server.URL+"/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(requestBody),
		)
		suite.Require().NoError(err)
		defer resp.Body.Close()
		
		suite.Equal(http.StatusOK, resp.StatusCode)
		
		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		suite.Require().NoError(err)
		
		requestIDs[i] = createResponse["request_id"].(string)
		suite.NotEmpty(requestIDs[i])
	}
	
	// 等待所有任務完成
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	
	for i, requestID := range requestIDs {
		finalStatus := suite.pollTaskStatus(ctx, requestID)
		suite.Equal("COMPLETED", finalStatus, "Task %d (%s) should complete", i, taskTypes[i])
	}
	
	// 驗證所有任務都有結果
	for i, requestID := range requestIDs {
		resp, err := suite.client.Get(suite.server.URL + "/api/v1/tasks/" + requestID + "/result")
		suite.Require().NoError(err)
		defer resp.Body.Close()
		
		suite.Equal(http.StatusOK, resp.StatusCode, "Task %d (%s) should have result", i, taskTypes[i])
	}
}

// TestErrorScenariosE2E 測試錯誤場景端到端流程
func (suite *E2ETestSuite) TestErrorScenariosE2E() {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		{
			name:           "invalid_json",
			method:         "POST",
			path:           "/api/v1/tasks",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid JSON should return 400",
		},
		{
			name:           "empty_request",
			method:         "POST",
			path:           "/api/v1/tasks",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			description:    "Empty request should return 400",
		},
		{
			name:           "nonexistent_task_status",
			method:         "GET",
			path:           "/api/v1/tasks/nonexistent/status",
			body:           nil,
			expectedStatus: http.StatusNotFound,
			description:    "Nonexistent task status should return 404",
		},
		{
			name:           "nonexistent_task_result",
			method:         "GET",
			path:           "/api/v1/tasks/nonexistent/result",
			body:           nil,
			expectedStatus: http.StatusNotFound,
			description:    "Nonexistent task result should return 404",
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var req *http.Request
			var err error
			
			if tt.body != nil {
				var requestBody []byte
				if str, ok := tt.body.(string); ok {
					requestBody = []byte(str)
				} else {
					requestBody, err = json.Marshal(tt.body)
					suite.Require().NoError(err)
				}
				req, err = http.NewRequest(tt.method, suite.server.URL+tt.path, bytes.NewBuffer(requestBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, suite.server.URL+tt.path, nil)
			}
			
			suite.Require().NoError(err)
			
			resp, err := suite.client.Do(req)
			suite.Require().NoError(err)
			defer resp.Body.Close()
			
			suite.Equal(tt.expectedStatus, resp.StatusCode, tt.description)
		})
	}
}

// TestConcurrentRequestsE2E 測試並發請求端到端流程
func (suite *E2ETestSuite) TestConcurrentRequestsE2E() {
	concurrency := 10
	results := make(chan string, concurrency)
	errors := make(chan error, concurrency)
	
	// 並發提交任務
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			taskRequest := testdata.MockTaskRequests["basic_request"]
			requestBody, err := json.Marshal(taskRequest)
			if err != nil {
				errors <- err
				return
			}
			
			resp, err := suite.client.Post(
				suite.server.URL+"/api/v1/tasks",
				"application/json",
				bytes.NewBuffer(requestBody),
			)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				return
			}
			
			var createResponse map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&createResponse)
			if err != nil {
				errors <- err
				return
			}
			
			requestID := createResponse["request_id"].(string)
			results <- requestID
		}(i)
	}
	
	// 收集結果
	successCount := 0
	errorCount := 0
	requestIDs := make([]string, 0, concurrency)
	
	for i := 0; i < concurrency; i++ {
		select {
		case requestID := <-results:
			successCount++
			requestIDs = append(requestIDs, requestID)
		case err := <-errors:
			errorCount++
			suite.T().Logf("Concurrent request error: %v", err)
		case <-time.After(30 * time.Second):
			suite.Fail("Timeout waiting for concurrent requests")
			return
		}
	}
	
	// 驗證結果
	suite.Equal(concurrency, successCount, "All concurrent requests should succeed")
	suite.Equal(0, errorCount, "No errors should occur")
	suite.Len(requestIDs, concurrency)
	
	// 驗證所有request_id都是唯一的
	uniqueIDs := make(map[string]bool)
	for _, id := range requestIDs {
		suite.False(uniqueIDs[id], "Duplicate request ID found: %s", id)
		uniqueIDs[id] = true
	}
}

// TestSystemResourcesE2E 測試系統資源管理
func (suite *E2ETestSuite) TestSystemResourcesE2E() {
	// 提交大量任務測試系統穩定性
	taskCount := 50
	requestIDs := make([]string, 0, taskCount)
	
	for i := 0; i < taskCount; i++ {
		taskRequest := testdata.MockTaskRequests["minimal_request"]
		requestBody, err := json.Marshal(taskRequest)
		suite.Require().NoError(err)
		
		resp, err := suite.client.Post(
			suite.server.URL+"/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(requestBody),
		)
		suite.Require().NoError(err)
		defer resp.Body.Close()
		
		suite.Equal(http.StatusOK, resp.StatusCode)
		
		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		suite.Require().NoError(err)
		
		requestID := createResponse["request_id"].(string)
		requestIDs = append(requestIDs, requestID)
		
		// 短暫延遲避免過度負載
		time.Sleep(10 * time.Millisecond)
	}
	
	// 驗證所有任務都能正常查詢狀態
	for i, requestID := range requestIDs {
		resp, err := suite.client.Get(suite.server.URL + "/api/v1/tasks/" + requestID + "/status")
		suite.Require().NoError(err, "Task %d status check failed", i)
		defer resp.Body.Close()
		
		suite.Equal(http.StatusOK, resp.StatusCode, "Task %d status should be accessible", i)
	}
}

// pollTaskStatus 輪詢任務狀態直到完成或超時
func (suite *E2ETestSuite) pollTaskStatus(ctx context.Context, requestID string) string {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			suite.Fail("Timeout waiting for task completion")
			return ""
		case <-ticker.C:
			resp, err := suite.client.Get(suite.server.URL + "/api/v1/tasks/" + requestID + "/status")
			suite.Require().NoError(err)
			defer resp.Body.Close()
			
			suite.Equal(http.StatusOK, resp.StatusCode)
			
			var statusResponse map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&statusResponse)
			suite.Require().NoError(err)
			
			status := statusResponse["status"].(string)
			suite.T().Logf("Task %s status: %s", requestID, status)
			
			if status == "COMPLETED" || status == "FAILED" {
				return status
			}
		}
	}
}

// TestE2ESuite 運行端到端測試套件
func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}