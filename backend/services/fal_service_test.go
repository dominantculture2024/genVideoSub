package services

import (
	"testing"
	"time"

	"genVideoSub/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFalMockService_SubmitTask(t *testing.T) {
	tests := []struct {
		name        string
		request     string // key in MockTaskRequests
		expectError bool
	}{
		{
			name:        "basic_request_success",
			request:     "basic_request",
			expectError: false,
		},
		{
			name:        "complex_request_success",
			request:     "complex_request",
			expectError: false,
		},
		{
			name:        "minimal_request_success",
			request:     "minimal_request",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 創建mock服務
			mockService := NewFalMockService()
			
			// 獲取測試請求
			request := testdata.MockTaskRequests[tt.request]
			require.NotNil(t, request, "Test request should not be nil")
			
			// 提交任務
			response, err := mockService.SubmitTask(request)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.RequestID)
				assert.Contains(t, response.RequestID, "mock_")
			}
		})
	}
}

func TestFalMockService_GetTaskStatus(t *testing.T) {
	mockService := NewFalMockService().(*FalMockService)
	
	// 首先提交一個任務
	request := testdata.MockTaskRequests["basic_request"]
	response, err := mockService.SubmitTask(request)
	require.NoError(t, err)
	require.NotNil(t, response)
	
	requestID := response.RequestID
	
	tests := []struct {
		name           string
		requestID      string
		expectError    bool
		expectedStatus string
	}{
		{
			name:           "valid_request_id",
			requestID:      requestID,
			expectError:    false,
			expectedStatus: "IN_QUEUE", // 剛提交的任務應該是IN_QUEUE狀態
		},
		{
			name:        "invalid_request_id",
			requestID:   "invalid_id",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := mockService.GetTaskStatus(tt.requestID)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}

func TestFalMockService_StatusTransition(t *testing.T) {
	mockService := NewFalMockService().(*FalMockService)
	
	// 提交任務
	request := testdata.MockTaskRequests["basic_request"]
	response, err := mockService.SubmitTask(request)
	require.NoError(t, err)
	
	requestID := response.RequestID
	
	// 測試初始狀態
	status, err := mockService.GetTaskStatus(requestID)
	assert.NoError(t, err)
	assert.Equal(t, "IN_QUEUE", status)
	
	// 等待6秒，狀態應該變為IN_PROGRESS
	time.Sleep(6 * time.Second)
	status, err = mockService.GetTaskStatus(requestID)
	assert.NoError(t, err)
	assert.Equal(t, "IN_PROGRESS", status)
	
	// 注意：完整的狀態轉換測試需要等待30秒，在實際測試中可能需要調整
}

func TestFalMockService_GetTaskResult(t *testing.T) {
	mockService := NewFalMockService().(*FalMockService)
	
	// 提交任務
	request := testdata.MockTaskRequests["basic_request"]
	response, err := mockService.SubmitTask(request)
	require.NoError(t, err)
	
	requestID := response.RequestID
	
	tests := []struct {
		name        string
		requestID   string
		expectError bool
		waitTime    time.Duration
	}{
		{
			name:        "task_not_completed",
			requestID:   requestID,
			expectError: true,
			waitTime:    0,
		},
		{
			name:        "invalid_request_id",
			requestID:   "invalid_id",
			expectError: true,
			waitTime:    0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.waitTime > 0 {
				time.Sleep(tt.waitTime)
			}
			
			result, err := mockService.GetTaskResult(tt.requestID)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Video.URL)
				assert.Greater(t, result.Video.Width, 0)
				assert.Greater(t, result.Video.Height, 0)
				assert.Greater(t, result.Video.Duration, float64(0))
			}
		})
	}
}

func TestFalMockService_RetryWithBackoff(t *testing.T) {
	mockService := NewFalMockService()
	
	tests := []struct {
		name        string
		operation   func() error
		maxRetries  int
		expectError bool
	}{
		{
			name: "successful_operation",
			operation: func() error {
				return nil
			},
			maxRetries:  3,
			expectError: false,
		},
		{
			name: "always_failing_operation",
			operation: func() error {
				return assert.AnError
			},
			maxRetries:  3,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockService.RetryWithBackoff(tt.operation, tt.maxRetries)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFalMockService_InterfaceCompliance(t *testing.T) {
	// 測試mock服務是否正確實現了接口
	mockService := NewFalMockService()
	assert.NotNil(t, mockService)
	
	// 測試所有接口方法都可以調用
	request := testdata.MockTaskRequests["basic_request"]
	
	// SubmitTask
	response, err := mockService.SubmitTask(request)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	
	// GetTaskStatus
	status, err := mockService.GetTaskStatus(response.RequestID)
	assert.NoError(t, err)
	assert.NotEmpty(t, status)
	
	// RetryWithBackoff
	err = mockService.RetryWithBackoff(func() error { return nil }, 1)
	assert.NoError(t, err)
}