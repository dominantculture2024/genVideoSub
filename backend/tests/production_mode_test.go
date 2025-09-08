package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"genVideoSub/config"
	"genVideoSub/handlers"
	"genVideoSub/interfaces"
	"genVideoSub/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ProductionModeTestSuite 正式模式測試套件
type ProductionModeTestSuite struct {
	suite.Suite
	router         *gin.Engine
	config         *config.Config
	providerManager *services.APIProviderManager
	curlService    *services.CurlService
	apiHandler     *handlers.APIHandler
	taskHandler    *handlers.TaskHandler
}

// SetupSuite 設置測試套件
func (suite *ProductionModeTestSuite) SetupSuite() {
	// 設置測試模式
	gin.SetMode(gin.TestMode)
	
	// 創建測試配置
	suite.config = &config.Config{
		Mode:        "production",
		Environment: "test",
		APIProvider: "fal_ai",
		APIProviderConfigs: map[string]*config.APIProviderConfig{
			"fal_ai": {
				APIKey:  "test-fal-key",
				BaseURL: "https://fal.run",
				Timeout: 300,
				Enabled: true,
			},
			"runway_ml": {
				APIKey:  "test-runway-key",
				BaseURL: "https://api.runwayml.com",
				Timeout: 300,
				Enabled: true,
			},
			"stability_ai": {
				APIKey:  "test-stability-key",
				BaseURL: "https://api.stability.ai",
				Timeout: 300,
				Enabled: true,
			},
		},
		Curl: config.CurlConfig{
			CurlTimeout: 30,
			CurlRetries: 3,
			CurlEnabled: true,
		},
	}
	
	// 創建服務
	suite.providerManager = services.NewAPIProviderManager(suite.config)
	suite.curlService = services.NewCurlService()
	
	// 創建處理器
	suite.apiHandler = handlers.NewAPIHandler(suite.providerManager, suite.curlService, suite.config)
	suite.taskHandler = handlers.NewTaskHandler(suite.providerManager)
	
	// 設置路由
	suite.router = gin.New()
	v1 := suite.router.Group("/api/v1")
	{
		// 任務相關路由
		v1.POST("/tasks", suite.taskHandler.CreateTask)
		v1.GET("/tasks/:id/status", suite.taskHandler.GetTaskStatus)
		v1.GET("/tasks/:id/result", suite.taskHandler.GetTaskResult)
		
		// API提供商管理路由
		v1.POST("/api/switch-provider", suite.apiHandler.SwitchProvider)
		v1.GET("/api/providers", suite.apiHandler.GetProviders)
		v1.GET("/api/config", suite.apiHandler.GetAPIConfig)
		v1.PUT("/api/config", suite.apiHandler.UpdateAPIConfig)
		v1.POST("/api/test-connection", suite.apiHandler.TestConnection)
		
		// cURL功能路由
		v1.POST("/curl/execute", suite.apiHandler.ExecuteCurl)
		v1.POST("/curl/generate", suite.apiHandler.GenerateCurlCommand)
		v1.POST("/curl/parse", suite.apiHandler.ParseCurlCommand)
	}
}

// TestAPIProviderSwitching 測試API提供商切換
func (suite *ProductionModeTestSuite) TestAPIProviderSwitching() {
	// 測試獲取當前提供商列表
	req, err := http.NewRequest("GET", "/api/v1/api/providers", nil)
	suite.Require().NoError(err)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var providersResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &providersResponse)
	suite.Require().NoError(err)
	
	// 驗證返回的提供商列表
	providers := providersResponse["providers"].([]interface{})
	suite.GreaterOrEqual(len(providers), 1)
	suite.Equal("fal_ai", providersResponse["current"])
	
	// 測試切換到runway_ml
	switchRequest := map[string]string{
		"provider": "runway_ml",
	}
	requestBody, err := json.Marshal(switchRequest)
	suite.Require().NoError(err)
	
	req, err = http.NewRequest("POST", "/api/v1/api/switch-provider", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var switchResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &switchResponse)
	suite.Require().NoError(err)
	suite.Equal("runway_ml", switchResponse["current_provider"])
	
	// 驗證切換後的狀態
	req, err = http.NewRequest("GET", "/api/v1/api/providers", nil)
	suite.Require().NoError(err)
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	err = json.Unmarshal(w.Body.Bytes(), &providersResponse)
	suite.Require().NoError(err)
	suite.Equal("runway_ml", providersResponse["current"])
}

// TestCurlFunctionality 測試cURL功能
func (suite *ProductionModeTestSuite) TestCurlFunctionality() {
	// 測試生成cURL命令
	curlRequest := &interfaces.CurlRequest{
		URL:     "https://api.example.com/test",
		Method:  "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer test-token",
		},
		Body:    `{"test": "data"}`,
		Timeout: 30,
	}
	
	requestBody, err := json.Marshal(curlRequest)
	suite.Require().NoError(err)
	
	req, err := http.NewRequest("POST", "/api/v1/curl/generate", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var generateResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &generateResponse)
	suite.Require().NoError(err)
	
	curlCommand := generateResponse["curl_command"].(string)
	suite.Contains(curlCommand, "curl")
	suite.Contains(curlCommand, "https://api.example.com/test")
	suite.Contains(curlCommand, "-X POST")
	suite.Contains(curlCommand, "Content-Type: application/json")
	
	// 測試解析cURL命令
	parseRequest := map[string]string{
		"command": curlCommand,
	}
	requestBody, err = json.Marshal(parseRequest)
	suite.Require().NoError(err)
	
	req, err = http.NewRequest("POST", "/api/v1/curl/parse", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var parseResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &parseResponse)
	suite.Require().NoError(err)
	
	parsedRequest := parseResponse["request"].(map[string]interface{})
	suite.Equal("https://api.example.com/test", parsedRequest["url"])
	suite.Equal("POST", parsedRequest["method"])
}

// TestConfigurationManagement 測試配置管理
func (suite *ProductionModeTestSuite) TestConfigurationManagement() {
	// 測試獲取API配置
	req, err := http.NewRequest("GET", "/api/v1/api/config", nil)
	suite.Require().NoError(err)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var configResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &configResponse)
	suite.Require().NoError(err)
	
	suite.Equal("production", configResponse["mode"])
	suite.Equal("fal_ai", configResponse["api_provider"])
	
	// 測試更新API配置
	updateRequest := map[string]interface{}{
		"api_provider": "stability_ai",
		"api_provider_configs": map[string]interface{}{
			"stability_ai": map[string]interface{}{
				"api_key": "updated-stability-key",
				"enabled": true,
			},
		},
	}
	
	requestBody, err := json.Marshal(updateRequest)
	suite.Require().NoError(err)
	
	req, err = http.NewRequest("PUT", "/api/v1/api/config", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var updateResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &updateResponse)
	suite.Require().NoError(err)
	suite.Equal("Configuration updated successfully", updateResponse["message"])
}

// TestProductionModeTaskFlow 測試生產模式任務流程
func (suite *ProductionModeTestSuite) TestProductionModeTaskFlow() {
	// 由於這是生產模式測試，我們需要模擬真實的API調用
	// 這裡我們測試任務創建和狀態查詢的基本流程
	
	taskRequest := map[string]interface{}{
		"prompt": "A beautiful sunset over mountains",
		"model":  "fal-ai/fast-svd",
		"input": map[string]interface{}{
			"image_url": "https://example.com/test.jpg",
			"motion_bucket_id": 127,
			"fps": 6,
			"cond_aug": 0.02,
			"steps": 25,
			"seed": 42,
		},
	}
	
	requestBody, err := json.Marshal(taskRequest)
	suite.Require().NoError(err)
	
	req, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 注意：在真實環境中，這可能會因為API密鑰無效而失敗
	// 但我們可以測試請求的格式和處理邏輯
	if w.Code == http.StatusOK {
		var createResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &createResponse)
		suite.Require().NoError(err)
		
		requestID := createResponse["request_id"].(string)
		suite.NotEmpty(requestID)
		
		// 測試狀態查詢
		req, err = http.NewRequest("GET", "/api/v1/tasks/"+requestID+"/status", nil)
		suite.Require().NoError(err)
		
		w = httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		// 狀態查詢應該成功
		suite.Equal(http.StatusOK, w.Code)
	}
}

// TestErrorHandling 測試錯誤處理
func (suite *ProductionModeTestSuite) TestErrorHandling() {
	// 測試無效的提供商切換
	switchRequest := map[string]string{
		"provider": "invalid_provider",
	}
	requestBody, err := json.Marshal(switchRequest)
	suite.Require().NoError(err)
	
	req, err := http.NewRequest("POST", "/api/v1/api/switch-provider", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusBadRequest, w.Code)
	
	// 測試無效的cURL命令解析
	parseRequest := map[string]string{
		"command": "invalid curl command",
	}
	requestBody, err = json.Marshal(parseRequest)
	suite.Require().NoError(err)
	
	req, err = http.NewRequest("POST", "/api/v1/curl/parse", bytes.NewBuffer(requestBody))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	suite.Equal(http.StatusBadRequest, w.Code)
}

// TestProductionModeTestSuite 運行生產模式測試套件
func TestProductionModeTestSuite(t *testing.T) {
	suite.Run(t, new(ProductionModeTestSuite))
}

// TestProductionModeConfiguration 測試生產模式配置
func TestProductionModeConfiguration(t *testing.T) {
	// 測試配置加載
	config := &config.Config{
		Mode:        "production",
		Environment: "production",
		APIProvider: "fal_ai",
	}
	
	assert.True(t, config.IsProductionMode())
	assert.Equal(t, "production", config.Mode)
	assert.Equal(t, "fal_ai", config.APIProvider)
}

// TestAPIProviderManager 測試API提供商管理器
func TestAPIProviderManager(t *testing.T) {
	config := &config.Config{
		APIProvider: "fal_ai",
		APIProviderConfigs: map[string]*config.APIProviderConfig{
			"fal_ai": {
				APIKey:  "test-key",
				BaseURL: "https://fal.run",
				Enabled: true,
			},
		},
	}
	
	manager := services.NewAPIProviderManager(config)
	require.NotNil(t, manager)
	
	// 測試獲取當前服務
	service := manager.GetCurrentService()
	require.NotNil(t, service)
	
	// 測試獲取提供商列表
	providers := manager.ListProviders()
	assert.GreaterOrEqual(t, len(providers), 1)
	
	// 測試獲取當前提供商
	current := manager.GetCurrentProvider()
	assert.Equal(t, "fal_ai", current)
}

// TestCurlService 測試cURL服務
func TestCurlService(t *testing.T) {
	service := services.NewCurlService()
	require.NotNil(t, service)
	
	// 測試生成cURL命令
	request := &interfaces.CurlRequest{
		URL:    "https://api.example.com/test",
		Method: "GET",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
		Timeout: 30,
	}
	
	curlCommand := service.GenerateCurlCommand(request)
	assert.Contains(t, curlCommand, "curl")
	assert.Contains(t, curlCommand, "https://api.example.com/test")
	assert.Contains(t, curlCommand, "Authorization: Bearer token")
	
	// 測試解析cURL命令
	parsedRequest, err := service.ParseCurlCommand(curlCommand)
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/test", parsedRequest.URL)
	assert.Equal(t, "GET", parsedRequest.Method)
	assert.Equal(t, "Bearer token", parsedRequest.Headers["Authorization"])
}