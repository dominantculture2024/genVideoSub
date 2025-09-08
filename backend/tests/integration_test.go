package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// IntegrationTestSuite 集成測試套件
type IntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
	mockService *services.FalMockService
}

// SetupSuite 設置測試套件
func (suite *IntegrationTestSuite) SetupSuite() {
	// 設置測試模式
	gin.SetMode(gin.TestMode)
	
	// 創建mock服務
	suite.mockService = services.NewFalMockService().(*services.FalMockService)
	
	// 創建處理器
	taskHandler := handlers.NewTaskHandler(suite.mockService)
	
	// 設置路由
	suite.router = gin.New()
	v1 := suite.router.Group("/api/v1")
	{
		v1.POST("/tasks", taskHandler.CreateTask)
		v1.GET("/tasks/:id/status", taskHandler.GetTaskStatus)
		v1.GET("/tasks/:id/result", taskHandler.GetTaskResult)
	}
}

// TestCompleteVideoGenerationWorkflow 測試完整的視頻生成工作流程
func (suite *IntegrationTestSuite) TestCompleteVideoGenerationWorkflow() {
	// 步驟1: 提交任務
	requestBody, err := json.Marshal(testdata.MockTaskRequests["basic_request"])
	suite.Require().NoError(err)
	
	req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 驗證任務創建成功
	suite.Equal(http.StatusOK, w.Code)
	
	var createResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResponse)
	suite.Require().NoError(err)
	
	requestID := createResponse["request_id"].(string)
	suite.NotEmpty(requestID)
	
	// 步驟2: 檢查初始狀態
	req, err = http.NewRequest("GET", "/api/v1/tasks/"+requestID+"/status", nil)
	suite.Require().NoError(err)
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var statusResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &statusResponse)
	suite.Require().NoError(err)
	
	suite.Equal("IN_QUEUE", statusResponse["status"])
	
	// 步驟3: 等待狀態轉換為IN_PROGRESS
	time.Sleep(6 * time.Second)
	
	req, err = http.NewRequest("GET", "/api/v1/tasks/"+requestID+"/status", nil)
	suite.Require().NoError(err)
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	err = json.Unmarshal(w.Body.Bytes(), &statusResponse)
	suite.Require().NoError(err)
	
	suite.Equal("IN_PROGRESS", statusResponse["status"])
	
	// 步驟4: 嘗試獲取結果（應該失敗，因為任務未完成）
	req, err = http.NewRequest("GET", "/api/v1/tasks/"+requestID+"/result", nil)
	suite.Require().NoError(err)
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 任務未完成，應該返回錯誤
	suite.NotEqual(http.StatusOK, w.Code)
}

// TestMultipleTasksWorkflow 測試多任務工作流程
func (suite *IntegrationTestSuite) TestMultipleTasksWorkflow() {
	taskRequests := []string{"basic_request", "complex_request", "minimal_request"}
	requestIDs := make([]string, len(taskRequests))
	
	// 提交多個任務
	for i, taskKey := range taskRequests {
		requestBody, err := json.Marshal(testdata.MockTaskRequests[taskKey])
		suite.Require().NoError(err)
		
		req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(requestBody))
		suite.Require().NoError(err)
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		suite.Equal(http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		
		requestIDs[i] = response["request_id"].(string)
		suite.NotEmpty(requestIDs[i])
	}
	
	// 檢查所有任務的狀態
	for _, requestID := range requestIDs {
		req, err := http.NewRequest("GET", "/api/v1/tasks/"+requestID+"/status", nil)
		suite.Require().NoError(err)
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		suite.Equal(http.StatusOK, w.Code)
		
		var statusResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &statusResponse)
		suite.Require().NoError(err)
		
		suite.Equal("IN_QUEUE", statusResponse["status"])
	}
}

// TestErrorHandling 測試錯誤處理
func (suite *IntegrationTestSuite) TestErrorHandling() {
	tests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "invalid_json_body",
			method:         "POST",
			url:            "/api/v1/tasks",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty_request_body",
			method:         "POST",
			url:            "/api/v1/tasks",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_task_id_status",
			method:         "GET",
			url:            "/api/v1/tasks/invalid_id/status",
			body:           nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid_task_id_result",
			method:         "GET",
			url:            "/api/v1/tasks/invalid_id/result",
			body:           nil,
			expectedStatus: http.StatusNotFound,
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var requestBody []byte
			var err error
			
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					requestBody = []byte(str)
				} else {
					requestBody, err = json.Marshal(tt.body)
					suite.Require().NoError(err)
				}
			}
			
			req, err := http.NewRequest(tt.method, tt.url, bytes.NewBuffer(requestBody))
			suite.Require().NoError(err)
			
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			
			suite.Equal(tt.expectedStatus, w.Code)
		})
	}
}

// TestConcurrentRequests 測試並發請求
func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	concurrency := 20
	results := make(chan string, concurrency)
	errors := make(chan error, concurrency)
	
	// 並發提交任務
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			requestBody, err := json.Marshal(testdata.MockTaskRequests["basic_request"])
			if err != nil {
				errors <- err
				return
			}
			
			req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(requestBody))
			if err != nil {
				errors <- err
				return
			}
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				errors <- assert.AnError
				return
			}
			
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				errors <- err
				return
			}
			
			requestID := response["request_id"].(string)
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
		case <-errors:
			errorCount++
		case <-time.After(10 * time.Second):
			suite.Fail("Timeout waiting for concurrent requests")
		}
	}
	
	// 驗證結果
	suite.Equal(concurrency, successCount)
	suite.Equal(0, errorCount)
	suite.Len(requestIDs, concurrency)
	
	// 驗證所有request_id都是唯一的
	uniqueIDs := make(map[string]bool)
	for _, id := range requestIDs {
		suite.False(uniqueIDs[id], "Duplicate request ID found: %s", id)
		uniqueIDs[id] = true
	}
}

// TestIntegrationSuite 運行集成測試套件
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}