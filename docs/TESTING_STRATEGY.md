# GenVideoSub 測試策略文檔

## 目錄
- [測試策略概述](#測試策略概述)
- [測試目標](#測試目標)
- [分層測試架構](#分層測試架構)
- [fal.ai API Mock策略](#falai-api-mock策略)
- [測試工具和框架](#測試工具和框架)
- [測試優先級和實施階段](#測試優先級和實施階段)
- [測試環境配置](#測試環境配置)
- [CI/CD整合策略](#cicd整合策略)
- [測試數據管理](#測試數據管理)
- [錯誤處理和邊界測試](#錯誤處理和邊界測試)
- [實施計畫](#實施計畫)

## 測試策略概述

GenVideoSub 項目採用分層測試策略，確保系統的可靠性、穩定性和用戶體驗。測試策略涵蓋從單元測試到端到端測試的完整測試金字塔，特別關注 fal.ai API 整合和異步任務處理的測試。

### 系統架構
- **後端**: Golang 服務，負責 API 整合和任務管理
- **前端**: Streamlit 應用，提供用戶界面
- **外部服務**: fal.ai API 用於視頻生成
- **存儲**: JSON 文件存儲任務數據

## 測試目標

### 主要目標
1. **功能正確性**: 確保所有功能按預期工作
2. **API 整合穩定性**: 驗證與 fal.ai API 的可靠整合
3. **異步處理可靠性**: 確保任務狀態管理和異步處理正確
4. **用戶體驗**: 驗證前端界面的響應性和易用性
5. **錯誤處理**: 確保系統能優雅處理各種錯誤情況
6. **性能**: 驗證系統在不同負載下的表現

### 質量指標
- 代碼覆蓋率: ≥ 80%
- API 響應時間: < 2秒
- 錯誤率: < 1%
- 系統可用性: ≥ 99%

## 分層測試架構

### 1. 單元測試 (Unit Tests)

#### 後端 Golang 組件

**Models 測試**
```go
// backend/tests/models/task_test.go
package models_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "genVideoSub/models"
)

func TestTaskCreation(t *testing.T) {
    task := models.NewTask("test-prompt", "image.jpg")
    assert.NotEmpty(t, task.ID)
    assert.Equal(t, "pending", task.Status)
    assert.Equal(t, "test-prompt", task.Prompt)
}

func TestTaskStatusTransition(t *testing.T) {
    task := models.NewTask("test", "image.jpg")
    task.UpdateStatus("processing")
    assert.Equal(t, "processing", task.Status)
}
```

**Services 測試**
```go
// backend/tests/services/fal_service_test.go
package services_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "genVideoSub/services"
)

func TestFalServiceInterface(t *testing.T) {
    mockService := &MockFalAIService{}
    result, err := mockService.SubmitTask("test-prompt", "image.jpg")
    assert.NoError(t, err)
    assert.NotEmpty(t, result.TaskID)
}
```

**Storage 測試**
```go
// backend/tests/storage/json_storage_test.go
package storage_test

import (
    "testing"
    "os"
    "genVideoSub/storage"
    "genVideoSub/models"
)

func TestSaveAndLoadTask(t *testing.T) {
    storage := storage.NewJSONStorage("test_tasks.json")
    task := models.NewTask("test", "image.jpg")
    
    err := storage.SaveTask(task)
    assert.NoError(t, err)
    
    loadedTask, err := storage.GetTask(task.ID)
    assert.NoError(t, err)
    assert.Equal(t, task.ID, loadedTask.ID)
    
    // 清理測試文件
    os.Remove("test_tasks.json")
}
```

**Handlers 測試**
```go
// backend/tests/handlers/task_handler_test.go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCreateTaskHandler(t *testing.T) {
    payload := map[string]string{
        "prompt": "test prompt",
        "image_path": "test.jpg",
    }
    jsonPayload, _ := json.Marshal(payload)
    
    req := httptest.NewRequest("POST", "/api/tasks", bytes.NewBuffer(jsonPayload))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(CreateTaskHandler)
    handler.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusCreated, rr.Code)
}
```

#### 前端 Streamlit 組件

**API Client 測試**
```python
# frontend/tests/test_api_client.py
import pytest
import requests_mock
from utils.api_client import APIClient

def test_create_task():
    with requests_mock.Mocker() as m:
        m.post('http://localhost:8080/api/tasks', 
               json={'task_id': 'test-123', 'status': 'pending'})
        
        client = APIClient('http://localhost:8080')
        result = client.create_task('test prompt', 'image.jpg')
        
        assert result['task_id'] == 'test-123'
        assert result['status'] == 'pending'

def test_get_task_status():
    with requests_mock.Mocker() as m:
        m.get('http://localhost:8080/api/tasks/test-123',
              json={'task_id': 'test-123', 'status': 'completed'})
        
        client = APIClient('http://localhost:8080')
        result = client.get_task_status('test-123')
        
        assert result['status'] == 'completed'
```

**UI Components 測試**
```python
# frontend/tests/test_components.py
import pytest
from unittest.mock import patch, MagicMock
from components.task_manager import TaskManager

def test_task_creation_ui():
    with patch('streamlit.text_input') as mock_input, \
         patch('streamlit.file_uploader') as mock_uploader, \
         patch('streamlit.button') as mock_button:
        
        mock_input.return_value = "test prompt"
        mock_uploader.return_value = MagicMock()
        mock_button.return_value = True
        
        task_manager = TaskManager()
        result = task_manager.create_task_form()
        
        assert result is not None
```

### 2. 整合測試 (Integration Tests)

#### API 端點測試
```go
// backend/tests/integration/api_test.go
package integration_test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "genVideoSub/main"
)

func TestAPIEndpoints(t *testing.T) {
    server := httptest.NewServer(main.SetupRoutes())
    defer server.Close()
    
    // 測試健康檢查
    resp, err := http.Get(server.URL + "/health")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // 測試任務創建
    // ...
}
```

#### 前後端通信測試
```python
# tests/integration/test_frontend_backend.py
import pytest
import requests
import subprocess
import time

class TestFrontendBackend:
    @classmethod
    def setup_class(cls):
        # 啟動後端服務
        cls.backend_process = subprocess.Popen(['go', 'run', 'main.go'])
        time.sleep(2)  # 等待服務啟動
    
    @classmethod
    def teardown_class(cls):
        cls.backend_process.terminate()
    
    def test_task_lifecycle(self):
        # 創建任務
        response = requests.post('http://localhost:8080/api/tasks', 
                               json={'prompt': 'test', 'image_path': 'test.jpg'})
        assert response.status_code == 201
        
        task_id = response.json()['task_id']
        
        # 檢查任務狀態
        response = requests.get(f'http://localhost:8080/api/tasks/{task_id}')
        assert response.status_code == 200
        assert response.json()['status'] in ['pending', 'processing']
```

### 3. 端到端測試 (E2E Tests)

```python
# tests/e2e/test_full_workflow.py
import pytest
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

class TestFullWorkflow:
    def setup_method(self):
        self.driver = webdriver.Chrome()
        self.driver.get('http://localhost:8501')
    
    def teardown_method(self):
        self.driver.quit()
    
    def test_complete_video_generation_workflow(self):
        # 上傳圖片
        file_input = self.driver.find_element(By.CSS_SELECTOR, "input[type='file']")
        file_input.send_keys("/path/to/test/image.jpg")
        
        # 輸入提示詞
        prompt_input = self.driver.find_element(By.CSS_SELECTOR, "input[placeholder*='prompt']")
        prompt_input.send_keys("A beautiful sunset over the ocean")
        
        # 提交任務
        submit_button = self.driver.find_element(By.XPATH, "//button[contains(text(), 'Generate')]")
        submit_button.click()
        
        # 等待任務完成
        WebDriverWait(self.driver, 60).until(
            EC.presence_of_element_located((By.XPATH, "//div[contains(text(), 'completed')]"))
        )
        
        # 驗證結果顯示
        result_element = self.driver.find_element(By.CSS_SELECTOR, ".video-result")
        assert result_element.is_displayed()
```

### 4. 性能測試 (Performance Tests)

```go
// backend/tests/performance/load_test.go
package performance_test

import (
    "testing"
    "sync"
    "time"
)

func TestConcurrentTaskCreation(t *testing.T) {
    const numGoroutines = 100
    const numRequestsPerGoroutine = 10
    
    var wg sync.WaitGroup
    start := time.Now()
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < numRequestsPerGoroutine; j++ {
                // 發送任務創建請求
                createTask("test prompt", "image.jpg")
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalRequests := numGoroutines * numRequestsPerGoroutine
    avgResponseTime := duration / time.Duration(totalRequests)
    
    t.Logf("Total requests: %d", totalRequests)
    t.Logf("Total time: %v", duration)
    t.Logf("Average response time: %v", avgResponseTime)
    
    // 斷言平均響應時間小於 100ms
    assert.Less(t, avgResponseTime, 100*time.Millisecond)
}
```

## fal.ai API Mock策略

### 當前狀況分析
目前的 `fal_service.go` 實現直接調用 fal.ai API，沒有抽象層或 Mock 機制。

### Mock 策略實施

#### 1. 接口抽象
```go
// backend/services/interfaces.go
package services

type FalAIInterface interface {
    SubmitTask(prompt, imagePath string) (*TaskResponse, error)
    GetTaskStatus(taskID string) (*StatusResponse, error)
    GetTaskResult(taskID string) (*ResultResponse, error)
}

type TaskResponse struct {
    TaskID string `json:"task_id"`
    Status string `json:"status"`
}

type StatusResponse struct {
    TaskID string `json:"task_id"`
    Status string `json:"status"`
    Progress int  `json:"progress"`
}

type ResultResponse struct {
    TaskID   string `json:"task_id"`
    Status   string `json:"status"`
    VideoURL string `json:"video_url"`
}
```

#### 2. Mock 實現
```go
// backend/services/mock_fal_service.go
package services

import (
    "fmt"
    "time"
    "math/rand"
)

type MockFalAIService struct {
    tasks map[string]*MockTask
}

type MockTask struct {
    ID        string
    Status    string
    Progress  int
    CreatedAt time.Time
    VideoURL  string
}

func NewMockFalAIService() *MockFalAIService {
    return &MockFalAIService{
        tasks: make(map[string]*MockTask),
    }
}

func (m *MockFalAIService) SubmitTask(prompt, imagePath string) (*TaskResponse, error) {
    taskID := fmt.Sprintf("mock-task-%d", rand.Int63())
    
    m.tasks[taskID] = &MockTask{
        ID:        taskID,
        Status:    "pending",
        Progress:  0,
        CreatedAt: time.Now(),
    }
    
    return &TaskResponse{
        TaskID: taskID,
        Status: "pending",
    }, nil
}

func (m *MockFalAIService) GetTaskStatus(taskID string) (*StatusResponse, error) {
    task, exists := m.tasks[taskID]
    if !exists {
        return nil, fmt.Errorf("task not found: %s", taskID)
    }
    
    // 模擬進度更新
    elapsed := time.Since(task.CreatedAt)
    if elapsed > 30*time.Second {
        task.Status = "completed"
        task.Progress = 100
        task.VideoURL = fmt.Sprintf("https://mock-cdn.com/videos/%s.mp4", taskID)
    } else if elapsed > 10*time.Second {
        task.Status = "processing"
        task.Progress = int(elapsed.Seconds() * 3) // 模擬進度增長
    }
    
    return &StatusResponse{
        TaskID:   taskID,
        Status:   task.Status,
        Progress: task.Progress,
    }, nil
}

func (m *MockFalAIService) GetTaskResult(taskID string) (*ResultResponse, error) {
    task, exists := m.tasks[taskID]
    if !exists {
        return nil, fmt.Errorf("task not found: %s", taskID)
    }
    
    if task.Status != "completed" {
        return nil, fmt.Errorf("task not completed: %s", taskID)
    }
    
    return &ResultResponse{
        TaskID:   taskID,
        Status:   task.Status,
        VideoURL: task.VideoURL,
    }, nil
}
```

#### 3. 配置驅動的 Mock
```go
// backend/config/config.go
package config

type Config struct {
    FalAI struct {
        UseMock bool   `json:"use_mock"`
        APIKey  string `json:"api_key"`
        BaseURL string `json:"base_url"`
    } `json:"fal_ai"`
}

// backend/services/factory.go
package services

import "genVideoSub/config"

func NewFalAIService(cfg *config.Config) FalAIInterface {
    if cfg.FalAI.UseMock {
        return NewMockFalAIService()
    }
    return NewRealFalAIService(cfg.FalAI.APIKey, cfg.FalAI.BaseURL)
}
```

#### 4. HTTP Mock Server
```go
// backend/tests/mocks/fal_mock_server.go
package mocks

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
)

func NewMockFalAIServer() *httptest.Server {
    mux := http.NewServeMux()
    
    mux.HandleFunc("/fal-ai/kling-video/v1/standard/text-to-video", func(w http.ResponseWriter, r *http.Request) {
        response := map[string]interface{}{
            "request_id": "mock-request-123",
            "status": "IN_QUEUE",
        }
        json.NewEncoder(w).Encode(response)
    })
    
    mux.HandleFunc("/fal-ai/queue/requests/", func(w http.ResponseWriter, r *http.Request) {
        response := map[string]interface{}{
            "status": "COMPLETED",
            "response_url": "https://mock-cdn.com/video.mp4",
        }
        json.NewEncoder(w).Encode(response)
    })
    
    return httptest.NewServer(mux)
}
```

### Mock 數據策略

#### 任務狀態模擬
```go
// backend/tests/data/mock_responses.go
package data

var MockTaskResponses = map[string]interface{}{
    "submit_success": map[string]string{
        "request_id": "test-task-123",
        "status": "IN_QUEUE",
    },
    "status_processing": map[string]interface{}{
        "status": "IN_PROGRESS",
        "progress": 45,
    },
    "status_completed": map[string]interface{}{
        "status": "COMPLETED",
        "response_url": "https://mock-cdn.com/test-video.mp4",
    },
    "error_invalid_input": map[string]string{
        "error": "Invalid input parameters",
        "code": "INVALID_INPUT",
    },
}
```

## 測試工具和框架

### Golang 測試工具

#### 核心測試框架
```go
// go.mod 依賴
require (
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    github.com/gorilla/mux v1.8.0
)
```

#### 測試工具配置
```makefile
# Makefile
.PHONY: test test-unit test-integration test-coverage

test:
	go test ./...

test-unit:
	go test -short ./...

test-integration:
	go test -run Integration ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-benchmark:
	go test -bench=. ./...

mock-generate:
	mockgen -source=services/interfaces.go -destination=mocks/mock_fal_service.go
```

### API 測試工具

#### Postman Collection
```json
{
  "info": {
    "name": "GenVideoSub API Tests",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Create Task",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"prompt\": \"A beautiful sunset\",\n  \"image_path\": \"test.jpg\"\n}"
        },
        "url": {
          "raw": "{{base_url}}/api/tasks",
          "host": ["{{base_url}}"],
          "path": ["api", "tasks"]
        }
      }
    }
  ]
}
```

### 前端測試工具

#### pytest 配置
```ini
# frontend/pytest.ini
[tool:pytest]
addopts = -v --tb=short --strict-markers
testpaths = tests
markers =
    unit: Unit tests
    integration: Integration tests
    e2e: End-to-end tests
    slow: Slow running tests
```

#### 測試依賴
```txt
# frontend/requirements-test.txt
pytest==7.4.0
pytest-mock==3.11.1
requests-mock==1.11.0
selenium==4.11.2
streamlit-testing==0.1.0
```

### 性能測試工具

#### Artillery 配置
```yaml
# tests/performance/artillery-config.yml
config:
  target: 'http://localhost:8080'
  phases:
    - duration: 60
      arrivalRate: 10
    - duration: 120
      arrivalRate: 20
  payload:
    path: "./test-data.csv"
    fields:
      - "prompt"
      - "image_path"

scenarios:
  - name: "Create and monitor tasks"
    weight: 100
    flow:
      - post:
          url: "/api/tasks"
          json:
            prompt: "{{ prompt }}"
            image_path: "{{ image_path }}"
          capture:
            - json: "$.task_id"
              as: "taskId"
      - get:
          url: "/api/tasks/{{ taskId }}"
```

## 測試優先級和實施階段

### 第一階段：核心功能測試 (高優先級)
1. **fal.ai API 整合測試**
   - Mock 服務實現
   - API 調用測試
   - 錯誤處理測試

2. **任務管理核心測試**
   - 任務創建和狀態更新
   - JSON 存儲操作
   - 任務生命週期管理

3. **基本通信測試**
   - 前後端 API 通信
   - HTTP 端點測試
   - 數據序列化/反序列化

### 第二階段：整合和界面測試 (中優先級)
1. **前端組件測試**
   - Streamlit 組件單元測試
   - 用戶界面交互測試
   - 文件上傳功能測試

2. **端到端工作流測試**
   - 完整視頻生成流程
   - 用戶操作模擬
   - 結果展示驗證

3. **錯誤處理和邊界測試**
   - 異常情況處理
   - 輸入驗證
   - 網絡錯誤恢復

### 第三階段：性能和穩定性測試 (低優先級)
1. **性能測試**
   - 負載測試
   - 壓力測試
   - 並發處理測試

2. **長期穩定性測試**
   - 長時間運行測試
   - 內存洩漏檢測
   - 資源使用監控

3. **邊緣案例測試**
   - 極端輸入測試
   - 資源限制測試
   - 網絡中斷恢復

## 測試環境配置

### 開發環境
```yaml
# config/test-config.yml
development:
  fal_ai:
    use_mock: true
    api_key: "test-key"
    base_url: "http://localhost:8081"
  storage:
    data_dir: "./test_data"
  server:
    port: 8080
    debug: true
```

### CI/CD 環境
```yaml
# config/ci-config.yml
ci:
  fal_ai:
    use_mock: true
    api_key: "${FAL_API_KEY_TEST}"
  storage:
    data_dir: "/tmp/genVideoSub_test"
  server:
    port: 8080
```

### 測試數據庫配置
```go
// backend/tests/setup/test_setup.go
package setup

import (
    "os"
    "path/filepath"
)

func SetupTestEnvironment() {
    // 創建測試數據目錄
    testDir := filepath.Join(os.TempDir(), "genVideoSub_test")
    os.MkdirAll(testDir, 0755)
    
    // 設置環境變量
    os.Setenv("DATA_DIR", testDir)
    os.Setenv("FAL_USE_MOCK", "true")
}

func CleanupTestEnvironment() {
    testDir := filepath.Join(os.TempDir(), "genVideoSub_test")
    os.RemoveAll(testDir)
}
```

## CI/CD整合策略

### GitHub Actions 配置
```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [ main, dev ]
  pull_request:
    branches: [ main ]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Install dependencies
      run: |
        cd backend
        go mod download
    
    - name: Run unit tests
      run: |
        cd backend
        go test -short ./...
    
    - name: Run integration tests
      run: |
        cd backend
        go test -run Integration ./...
    
    - name: Generate coverage report
      run: |
        cd backend
        go test -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage artifacts
      uses: actions/upload-artifact@v3
      with:
        name: backend-coverage
        path: backend/coverage.html

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.9'
    
    - name: Install dependencies
      run: |
        cd frontend
        pip install -r requirements.txt
        pip install -r requirements-test.txt
    
    - name: Run tests
      run: |
        cd frontend
        pytest tests/ -v --cov=. --cov-report=html --cov-report=term
    
    - name: Upload frontend coverage artifacts
      uses: actions/upload-artifact@v3
      with:
        name: frontend-coverage
        path: frontend/htmlcov/

  e2e-tests:
    runs-on: ubuntu-latest
    needs: [backend-tests, frontend-tests]
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go and Python
      # ... setup steps
    
    - name: Start services
      run: |
        cd backend && go run main.go &
        cd frontend && streamlit run app.py &
        sleep 10
    
    - name: Run E2E tests
      run: |
        pytest tests/e2e/ -v
```

### 測試報告整合
```yaml
# .github/workflows/test-report.yml
name: Test Report

on:
  workflow_run:
    workflows: ["Test Suite"]
    types:
      - completed

jobs:
  test-report:
    runs-on: ubuntu-latest
    steps:
    - name: Generate Test Report
      uses: dorny/test-reporter@v1
      with:
        name: Test Results
        path: 'test-results.xml'
        reporter: java-junit
```

## 測試數據管理

### 測試數據結構
```
tests/
├── data/
│   ├── images/
│   │   ├── test-image-1.jpg
│   │   ├── test-image-2.png
│   │   └── invalid-image.txt
│   ├── prompts/
│   │   ├── valid-prompts.json
│   │   └── invalid-prompts.json
│   └── responses/
│       ├── fal-success-response.json
│       └── fal-error-response.json
└── fixtures/
    ├── tasks.json
    └── mock-config.yml
```

### 測試數據生成
```go
// backend/tests/data/generator.go
package data

import (
    "encoding/json"
    "genVideoSub/models"
)

func GenerateTestTasks(count int) []*models.Task {
    tasks := make([]*models.Task, count)
    for i := 0; i < count; i++ {
        tasks[i] = &models.Task{
            ID:        fmt.Sprintf("test-task-%d", i),
            Prompt:    fmt.Sprintf("Test prompt %d", i),
            ImagePath: fmt.Sprintf("test-image-%d.jpg", i),
            Status:    "pending",
            CreatedAt: time.Now(),
        }
    }
    return tasks
}

func LoadTestData(filename string) (interface{}, error) {
    data, err := os.ReadFile(filepath.Join("tests/data", filename))
    if err != nil {
        return nil, err
    }
    
    var result interface{}
    err = json.Unmarshal(data, &result)
    return result, err
}
```

## 錯誤處理和邊界測試

### 錯誤場景測試
```go
// backend/tests/error_handling_test.go
package tests

func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name           string
        input          interface{}
        expectedError  string
        expectedStatus int
    }{
        {
            name:           "Empty prompt",
            input:          map[string]string{"prompt": "", "image_path": "test.jpg"},
            expectedError:  "prompt cannot be empty",
            expectedStatus: 400,
        },
        {
            name:           "Invalid image format",
            input:          map[string]string{"prompt": "test", "image_path": "test.txt"},
            expectedError:  "unsupported image format",
            expectedStatus: 400,
        },
        {
            name:           "Missing image file",
            input:          map[string]string{"prompt": "test", "image_path": "nonexistent.jpg"},
            expectedError:  "image file not found",
            expectedStatus: 404,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 執行測試邏輯
        })
    }
}
```

### 網絡錯誤模擬
```go
// backend/tests/network_error_test.go
package tests

func TestNetworkErrorHandling(t *testing.T) {
    // 模擬網絡超時
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(10 * time.Second) // 模擬超時
    }))
    defer server.Close()
    
    service := NewFalAIService(server.URL, "test-key")
    _, err := service.SubmitTask("test", "test.jpg")
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}

func TestAPIRateLimiting(t *testing.T) {
    // 模擬 API 限流
    callCount := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        callCount++
        if callCount > 3 {
            w.WriteHeader(http.StatusTooManyRequests)
            return
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()
    
    // 測試重試邏輯
}
```

## 實施計畫

### 第1週：基礎設施建設
- [ ] 設置測試目錄結構
- [ ] 配置測試工具和依賴
- [ ] 實現 fal.ai API Mock 接口
- [ ] 創建基本測試配置

### 第2週：核心功能測試
- [ ] 實現後端單元測試
- [ ] 創建 API 整合測試
- [ ] 實現前端組件測試
- [ ] 設置 CI/CD 基礎流程

### 第3週：整合和端到端測試
- [ ] 實現端到端測試套件
- [ ] 創建性能測試基準
- [ ] 完善錯誤處理測試
- [ ] 優化測試覆蓋率

### 第4週：優化和文檔
- [ ] 性能測試和優化
- [ ] 完善測試文檔
- [ ] 建立測試最佳實踐
- [ ] 團隊培訓和知識轉移

### 持續改進
- 定期審查測試覆蓋率
- 更新測試策略和工具
- 收集和分析測試指標
- 優化測試執行效率

---

## 結論

本測試策略提供了 GenVideoSub 項目的全面測試方案，涵蓋從單元測試到端到端測試的完整測試金字塔。通過實施分層測試架構、fal.ai API Mock 策略和完善的 CI/CD 整合，我們能夠確保系統的可靠性、穩定性和用戶體驗。

測試策略的成功實施需要團隊的共同努力和持續改進。建議按照實施計畫逐步推進，並根據項目發展需要調整測試策略和優先級。