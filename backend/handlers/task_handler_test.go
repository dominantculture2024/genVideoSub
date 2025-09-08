package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"genVideoSub/services"
	"genVideoSub/testdata"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	// 設置測試模式
	gin.SetMode(gin.TestMode)
	
	// 創建mock服務
	mockService := services.NewFalMockService()
	
	// 創建處理器
	handler := NewTaskHandler(mockService)
	
	// 設置路由
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.POST("/tasks", handler.CreateTask)
		v1.GET("/tasks/:id/status", handler.GetTaskStatus)
		v1.GET("/tasks/:id/result", handler.GetTaskResult)
	}
	
	return router
}

func TestTaskHandler_CreateTask(t *testing.T) {
	router := setupTestRouter()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid_basic_request",
			requestBody:    testdata.MockTaskRequests["basic_request"],
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "valid_complex_request",
			requestBody:    testdata.MockTaskRequests["complex_request"],
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "valid_minimal_request",
			requestBody:    testdata.MockTaskRequests["minimal_request"],
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid_json",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "empty_request",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 準備請求體
			var requestBody []byte
			var err error
			
			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}
			
			// 創建請求
			req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			
			// 執行請求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// 驗證響應
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if !tt.expectError {
				// 解析響應
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				// 驗證響應包含request_id
				assert.Contains(t, response, "request_id")
				assert.NotEmpty(t, response["request_id"])
			}
		})
	}
}

func TestTaskHandler_GetTaskStatus(t *testing.T) {
	router := setupTestRouter()
	
	// 首先創建一個任務
	requestBody, err := json.Marshal(testdata.MockTaskRequests["basic_request"])
	require.NoError(t, err)
	
	req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	
	// 解析響應獲取request_id
	var createResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResponse)
	require.NoError(t, err)
	
	requestID := createResponse["request_id"].(string)
	require.NotEmpty(t, requestID)
	
	tests := []struct {
		name           string
		requestID      string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid_request_id",
			requestID:      requestID,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid_request_id",
			requestID:      "invalid_id",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 創建請求
			req, err := http.NewRequest("GET", "/api/v1/tasks/"+tt.requestID+"/status", nil)
			require.NoError(t, err)
			
			// 執行請求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// 驗證響應
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if !tt.expectError {
				// 解析響應
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				// 驗證響應包含status
				assert.Contains(t, response, "status")
				assert.NotEmpty(t, response["status"])
			}
		})
	}
}

func TestTaskHandler_GetTaskResult(t *testing.T) {
	router := setupTestRouter()
	
	// 首先創建一個任務
	requestBody, err := json.Marshal(testdata.MockTaskRequests["basic_request"])
	require.NoError(t, err)
	
	req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	
	// 解析響應獲取request_id
	var createResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResponse)
	require.NoError(t, err)
	
	requestID := createResponse["request_id"].(string)
	require.NotEmpty(t, requestID)
	
	tests := []struct {
		name           string
		requestID      string
		expectedStatus int
		expectError    bool
		waitTime       time.Duration
	}{
		{
			name:           "task_not_completed",
			requestID:      requestID,
			expectedStatus: http.StatusAccepted, // 任務未完成
			expectError:    true,
			waitTime:       0,
		},
		{
			name:           "invalid_request_id",
			requestID:      "invalid_id",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			waitTime:       0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.waitTime > 0 {
				time.Sleep(tt.waitTime)
			}
			
			// 創建請求
			req, err := http.NewRequest("GET", "/api/v1/tasks/"+tt.requestID+"/result", nil)
			require.NoError(t, err)
			
			// 執行請求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// 驗證響應
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if !tt.expectError {
				// 解析響應
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				// 驗證響應包含video信息
				assert.Contains(t, response, "video")
			}
		})
	}
}

func TestTaskHandler_ConcurrentRequests(t *testing.T) {
	router := setupTestRouter()
	
	// 測試並發請求
	concurrency := 10
	done := make(chan bool, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			// 準備請求
			requestBody, err := json.Marshal(testdata.MockTaskRequests["basic_request"])
			assert.NoError(t, err)
			
			req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			
			// 執行請求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// 驗證響應
			assert.Equal(t, http.StatusOK, w.Code)
			
			done <- true
		}(i)
	}
	
	// 等待所有請求完成
	for i := 0; i < concurrency; i++ {
		<-done
	}
}