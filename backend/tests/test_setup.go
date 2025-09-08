package tests

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"genVideoSub/config"
	"genVideoSub/services"
	"github.com/gin-gonic/gin"
)

// TestEnvironment 測試環境結構
type TestEnvironment struct {
	Config      *config.TestConfig
	Factory     *services.ServiceFactory
	Service     services.FalAIInterface
	TempDir     string
	CleanupFunc func()
}

// SetupTestEnvironment 設置測試環境
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	// 設置測試模式
	gin.SetMode(gin.TestMode)
	
	// 創建臨時目錄
	tempDir, err := os.MkdirTemp("", "genVideoSub_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// 設置測試環境變量
	setTestEnvVars(tempDir)
	
	// 加載測試配置
	config := config.LoadTestConfig()
	
	// 創建服務工廠
	factory := services.NewServiceFactory(config)
	
	// 創建服務實例
	service, err := factory.CreateFalAIService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	
	// 創建清理函數
	cleanupFunc := func() {
		cleanupTestEnvironment(tempDir)
	}
	
	return &TestEnvironment{
		Config:      config,
		Factory:     factory,
		Service:     service,
		TempDir:     tempDir,
		CleanupFunc: cleanupFunc,
	}
}

// SetupMockEnvironment 設置Mock測試環境
func SetupMockEnvironment(t *testing.T) *TestEnvironment {
	// 強制使用Mock服務
	os.Setenv("USE_MOCK_SERVICE", "true")
	os.Setenv("FAL_API_KEY", "")
	
	return SetupTestEnvironment(t)
}

// SetupRealServiceEnvironment 設置真實服務測試環境
func SetupRealServiceEnvironment(t *testing.T, apiKey string) *TestEnvironment {
	if apiKey == "" {
		t.Skip("API key not provided, skipping real service test")
	}
	
	// 設置真實服務環境變量
	os.Setenv("USE_MOCK_SERVICE", "false")
	os.Setenv("FAL_API_KEY", apiKey)
	
	return SetupTestEnvironment(t)
}

// setTestEnvVars 設置測試環境變量
func setTestEnvVars(tempDir string) {
	// 設置默認測試環境變量
	envVars := map[string]string{
		"USE_MOCK_SERVICE": "true",
		"TEST_TIMEOUT":     "300",
		"TEST_DATA_PATH":   filepath.Join(tempDir, "testdata"),
		"LOG_LEVEL":        "debug",
		"MOCK_DELAY":       "100", // 減少延遲以加快測試
		"MOCK_FAIL_RATE":   "0.0",
		"MOCK_DATA_SIZE":   "5",
		"MAX_CONCURRENT":   "10",
		"RATE_LIMIT":       "100",
	}
	
	for key, value := range envVars {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// cleanupTestEnvironment 清理測試環境
func cleanupTestEnvironment(tempDir string) {
	// 清理臨時目錄
	if tempDir != "" {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Failed to cleanup temp dir %s: %v", tempDir, err)
		}
	}
	
	// 清理環境變量
	envVarsToClean := []string{
		"USE_MOCK_SERVICE",
		"FAL_API_KEY",
		"TEST_TIMEOUT",
		"TEST_DATA_PATH",
		"LOG_LEVEL",
		"MOCK_DELAY",
		"MOCK_FAIL_RATE",
		"MOCK_DATA_SIZE",
		"MAX_CONCURRENT",
		"RATE_LIMIT",
	}
	
	for _, envVar := range envVarsToClean {
		os.Unsetenv(envVar)
	}
}

// CreateTestDataDir 創建測試數據目錄
func CreateTestDataDir(baseDir string) error {
	testDataDir := filepath.Join(baseDir, "testdata")
	return os.MkdirAll(testDataDir, 0755)
}

// GetTestAPIKey 獲取測試API密鑰
func GetTestAPIKey() string {
	return os.Getenv("FAL_API_KEY_TEST")
}

// SkipIfNoAPIKey 如果沒有API密鑰則跳過測試
func SkipIfNoAPIKey(t *testing.T) {
	if GetTestAPIKey() == "" {
		t.Skip("FAL_API_KEY_TEST not set, skipping integration test with real API")
	}
}

// TestMain 測試主函數，用於全局測試設置
func TestMain(m *testing.M) {
	// 全局測試設置
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// 設置測試模式
	gin.SetMode(gin.TestMode)
	
	// 運行測試
	code := m.Run()
	
	// 全局清理
	// 這裡可以添加全局清理邏輯
	
	os.Exit(code)
}

// AssertServiceType 斷言服務類型
func AssertServiceType(t *testing.T, service services.FalAIInterface, expectedType string) {
	switch expectedType {
	case "mock":
		if _, ok := service.(*services.FalMockService); !ok {
			t.Errorf("Expected mock service, got %T", service)
		}
	case "real":
		if _, ok := service.(*services.FalAIService); !ok {
			t.Errorf("Expected real service, got %T", service)
		}
	default:
		t.Errorf