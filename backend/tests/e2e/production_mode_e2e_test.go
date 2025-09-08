package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

// ProductionModeE2ETestSuite 正式模式端到端測試套件
type ProductionModeE2ETestSuite struct {
	suite.Suite
	router           *gin.Engine
	config           *config.Config
	apiProviderMgr   *services.APIProviderManager
	curlService      *services.CurlService
	taskHandler      *handlers.TaskHandler
	apiHandler       *handlers.APIHandler
	testVideoPath    string
}

// SetupSuite 設置測試套件
func (suite *ProductionModeE2ETestSuite) SetupSuite() {
	// 設置Gin為測試模式
	gin.SetMode(gin.TestMode)

	// 創建測試配置
	suite.config = &config.Config{
		Mode:        "production",
		Environment: "test",
		APIProvider: "fal_ai",
		Server: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		APIProviderConfigs: config.APIProviderConfigs{
			FalAI: config.ProviderConfig{
				APIKey:  "test-fal-key",
				BaseURL: "https://fal.run/fal-ai",
				Enabled: true,
			},
			RunwayML: config.ProviderConfig{
				APIKey:  "test-runway-key",
				BaseURL: "https://api.runwayml.com",
				Enabled: true,
			},
			StabilityAI: config.ProviderConfig{
				APIKey:  "test-stability-key",
				BaseURL: "https://api.stability.ai",
				Enabled: true,
			},
		},
		Curl: config.CurlConfig{
			CurlEnabled: true,
			CurlTimeout: 30,
			CurlRetries: 3,
		},
	}

	// 初始化服務
	suite.apiProviderMgr = services.NewAPIProviderManager(suite.config)
	suite.curlService = services.NewCurlService()

	// 初始化處理器
	suite.taskHandler = handlers.NewTaskHandler(suite.apiProviderMgr)
	suite.apiHandler = handlers.NewAPIHandler(suite.apiProviderMgr, suite.curlService, suite.config)

	// 設置路由
	suite.router = gin.New()
	suite.setupRoutes()

	// 設置測試視頻文件路徑
	suite.testVideoPath = "C:\\Paul\\toDo\\videoSimpleSub\\測試檔案\\1.mp4"
}

// setupRoutes 設置路由
func (suite *ProductionModeE2ETestSuite) setupRoutes() {
	v1 := suite.router.Group("/api/v1")
	{
		// 任務相關路由
		v1.POST("/tasks", suite.taskHandler.SubmitTask)
		v1.GET("/tasks/:id/status", suite.taskHandler.GetTaskStatus)
		v1.GET("/tasks/:id/result", suite.taskHandler.GetTaskResult)

		// API管理路由
		v1.GET("/providers", suite.apiHandler.GetProviders)
		v1.POST("/providers/switch", suite.apiHandler.SwitchProvider)
		v1.GET("/config", suite.apiHandler.GetAPIConfig)
		v1.PUT("/config", suite.apiHandler.UpdateAPIConfig)

		// cURL相關路由
		v1.POST("/curl/execute", suite.apiHandler.ExecuteCurl)
		v1.POST("/curl/generate", suite.apiHandler.GenerateCurlCommand)
		v1.POST("/curl/parse", suite.apiHandler.ParseCurlCommand)
	}
}

// TestCompleteProductionWorkflow 測試完整的正式模式工作流程
func (suite *ProductionModeE2ETestSuite) TestCompleteProductionWorkflow() {
	// 1. 檢查初始API提供商
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/providers", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var providersResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &providersResp)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "fal_ai", providersResp["current_provider"])

	// 2. 切換到RunwayML
	w = httptest.NewRecorder()
	switchReq := map[string]string{"provider": "runway_ml"}
	reqBody, _ := json.Marshal(switchReq)
	req, _ = http.NewRequest("POST", "/api/v1/providers/switch", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 3. 驗證切換成功
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/providers", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &providersResp)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "runway_ml", providersResp["current_provider"])

	// 4. 測試cURL功能
	curlReq := interfaces.CurlRequest{
		URL:    "https://api.example.com/test",
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:    `{"test": "data"}`,
		Timeout: 30,
	}

	w = httptest.NewRecorder()
	reqBody, _ = json.Marshal(curlReq)
	req, _ = http.NewRequest("POST", "/api/v1/curl/generate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var curlResp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &curlResp)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), curlResp["curl_command"], "curl")
	assert.Contains(suite.T(), curlResp["curl_command"], "https://api.example.com/test")

	// 5. 測試任務提交（會失敗，因為使用測試API密鑰）
	if _, err := os.Stat(suite.testVideoPath); err == nil {
		// 創建multipart表單
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加文件
		file, err := os.Open(suite.testVideoPath)
		if err == nil {
			defer file.Close()
			part, _ := writer.CreateFormFile("video", "1.mp4")
			io.Copy(part, file)
		}

		// 添加其他字段
		writer.WriteField("prompt", "Test video generation")
		writer.Close()

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/v1/tasks", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		suite.router.ServeHTTP(w, req)

		// 由於使用測試API密鑰，預期會失敗
		assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	}
}

// TestAPIProviderSwitching 測試API提供商切換
func (suite *ProductionModeE2ETestSuite) TestAPIProviderSwitching() {
	providers := []string{"fal_ai", "runway_ml", "stability_ai"}

	for _, provider := range providers {
		// 切換提供商
		w := httptest.NewRecorder()
		switchReq := map[string]string{"provider": provider}
		reqBody, _ := json.Marshal(switchReq)
		req, _ := http.NewRequest("POST", "/api/v1/providers/switch", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		// 驗證切換成功
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/providers", nil)
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var providersResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &providersResp)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), provider, providersResp["current_provider"])
	}
}

// TestCurlFunctionality 測試cURL功能
func (suite *ProductionModeE2ETestSuite) TestCurlFunctionality() {
	// 測試生成cURL命令
	curlReq := interfaces.CurlRequest{
		URL:    "https://api.example.com/test",
		Method: "GET",
		Timeout: 10,
	}

	w := httptest.NewRecorder()
	reqBody, _ := json.Marshal(curlReq)
	req, _ := http.NewRequest("POST", "/api/v1/curl/generate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var generateResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &generateResp)
	require.NoError(suite.T(), err)
	curlCommand := generateResp["curl_command"]
	assert.Contains(suite.T(), curlCommand, "curl")
	assert.Contains(suite.T(), curlCommand, "https://api.example.com/test")

	// 測試解析cURL命令
	parseReq := map[string]string{"curl_command": curlCommand}
	w = httptest.NewRecorder()
	reqBody, _ = json.Marshal(parseReq)
	req, _ = http.NewRequest("POST", "/api/v1/curl/parse", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var parseResp interfaces.CurlRequest
	err = json.Unmarshal(w.Body.Bytes(), &parseResp)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "https://api.example.com/test", parseResp.URL)
	assert.Equal(suite.T(), "GET", parseResp.Method)
}

// TestConfigurationManagement 測試配置管理
func (suite *ProductionModeE2ETestSuite) TestConfigurationManagement() {
	// 獲取當前配置
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/config", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var configResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &configResp)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "production", configResp["mode"])
	assert.Equal(suite.T(), "fal_ai", configResp["api_provider"])

	// 更新配置
	updateReq := map[string]interface{}{
		"api_provider": "runway_ml",
		"api_provider_configs": map[string]interface{}{
			"runway_ml": map[string]interface{}{
				"api_key": "new-runway-key",
				"enabled": true,
			},
		},
	}

	w = httptest.NewRecorder()
	reqBody, _ := json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", "/api/v1/config", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 驗證配置更新
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/config", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &configResp)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "runway_ml", configResp["api_provider"])
}

// TestErrorHandling 測試錯誤處理
func (suite *ProductionModeE2ETestSuite) TestErrorHandling() {
	// 測試無效的提供商切換
	w := httptest.NewRecorder()
	switchReq := map[string]string{"provider": "invalid_provider"}
	reqBody, _ := json.Marshal(switchReq)
	req, _ := http.NewRequest("POST", "/api/v1/providers/switch", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// 測試無效的cURL命令解析
	parseReq := map[string]string{"curl_command": "invalid curl command"}
	w = httptest.NewRecorder()
	reqBody, _ = json.Marshal(parseReq)
	req, _ = http.NewRequest("POST", "/api/v1/curl/parse", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// 測試無效的任務ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/tasks/invalid-id/status", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// TestProductionModeE2E 運行端到端測試套件
func TestProductionModeE2E(t *testing.T) {
	suite.Run(t, new(ProductionModeE2ETestSuite))
}