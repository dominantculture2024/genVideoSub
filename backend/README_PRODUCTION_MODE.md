# GenVideoSub 正式模式使用指南

## 概述

本項目已成功實現正式模式功能，支援多個API提供商的視頻生成服務，包括：
- **fal.ai** (預設)
- **Runway ML**
- **Stability AI**

## 主要功能

### 1. 正式模式 (Production Mode)
- 預設使用 fal.ai API
- 支援動態切換API提供商
- 完整的配置管理系統
- 生產級錯誤處理和日誌記錄

### 2. API提供商管理
- 統一的API提供商接口
- 動態切換不同的API服務
- 配置驗證和連接測試
- 重試機制和錯誤恢復

### 3. cURL功能
- 生成cURL命令
- 解析cURL命令
- 執行HTTP請求
- API請求格式轉換

## 配置文件

### config.json 結構
```json
{
  "mode": "production",
  "environment": "production",
  "api_provider": "fal_ai",
  "server": {
    "port": 8080,
    "host": "localhost"
  },
  "api_provider_configs": {
    "fal_ai": {
      "api_key": "YOUR_FAL_API_KEY",
      "base_url": "https://fal.run/fal-ai",
      "timeout": 300,
      "enabled": true
    },
    "runway_ml": {
      "api_key": "YOUR_RUNWAY_API_KEY",
      "base_url": "https://api.runwayml.com",
      "timeout": 300,
      "enabled": false
    },
    "stability_ai": {
      "api_key": "YOUR_STABILITY_API_KEY",
      "base_url": "https://api.stability.ai",
      "timeout": 300,
      "enabled": false
    }
  },
  "curl": {
    "curl_enabled": true,
    "curl_timeout": 30,
    "curl_retries": 3
  }
}
```

## API 端點

### 任務管理
- `POST /api/v1/tasks` - 提交視頻生成任務
- `GET /api/v1/tasks/:id/status` - 獲取任務狀態
- `GET /api/v1/tasks/:id/result` - 獲取任務結果

### API提供商管理
- `GET /api/v1/providers` - 獲取所有提供商信息
- `POST /api/v1/providers/switch` - 切換API提供商
- `GET /api/v1/config` - 獲取當前配置
- `PUT /api/v1/config` - 更新配置

### cURL功能
- `POST /api/v1/curl/execute` - 執行cURL請求
- `POST /api/v1/curl/generate` - 生成cURL命令
- `POST /api/v1/curl/parse` - 解析cURL命令

## 使用方法

### 1. 啟動服務器
```bash
cd backend
go run main.go
```

### 2. 切換API提供商
```bash
curl -X POST http://localhost:8080/api/v1/providers/switch \
  -H "Content-Type: application/json" \
  -d '{"provider": "runway_ml"}'
```

### 3. 提交視頻生成任務
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -F "video=@test_video.mp4" \
  -F "prompt=Generate a new video based on this input"
```

### 4. 使用cURL功能
```bash
# 生成cURL命令
curl -X POST http://localhost:8080/api/v1/curl/generate \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://api.example.com/test",
    "method": "POST",
    "headers": {"Content-Type": "application/json"},
    "body": "{\"test\": \"data\"}",
    "timeout": 30
  }'
```

## 測試

### 運行所有測試
```bash
# Windows
run_tests.bat

# 或手動運行
go test ./tests/unit/... -v
go test ./tests/integration_test.go ./tests/test_setup.go -v
go test ./tests/production_mode_test.go ./tests/test_setup.go -v
go test ./tests/e2e/... -v
```

### 測試覆蓋範圍
- **單元測試**: API提供商服務、cURL服務、配置管理
- **集成測試**: API端點、服務集成
- **端到端測試**: 完整工作流程測試
- **生產模式測試**: 正式模式特定功能測試

## 項目結構

```
backend/
├── config/
│   ├── config.go           # 配置管理
│   └── test_config.go      # 測試配置
├── handlers/
│   ├── api_handler.go      # API管理處理器
│   └── task_handler.go     # 任務處理器
├── interfaces/
│   └── interfaces.go       # 接口定義
├── services/
│   ├── api_provider_manager.go  # API提供商管理器
│   ├── curl_service.go          # cURL服務
│   ├── fal_service.go           # fal.ai服務
│   ├── runway_service.go        # Runway ML服務
│   └── stability_service.go     # Stability AI服務
├── tests/
│   ├── unit/                    # 單元測試
│   ├── e2e/                     # 端到端測試
│   ├── integration_test.go      # 集成測試
│   ├── production_mode_test.go  # 生產模式測試
│   └── test_setup.go           # 測試設置
├── config.json             # 配置文件
├── main.go                 # 主程序
└── README_PRODUCTION_MODE.md
```

## 環境要求

- Go 1.21+
- 有效的API密鑰（fal.ai、Runway ML、Stability AI）
- 網絡連接

## 故障排除

### 常見問題

1. **API密鑰無效**
   - 檢查config.json中的API密鑰是否正確
   - 確認API密鑰有足夠的權限

2. **服務無法啟動**
   - 檢查端口8080是否被占用
   - 確認配置文件格式正確

3. **測試失敗**
   - 確認Go環境正確安裝
   - 檢查網絡連接
   - 驗證測試數據文件存在

### 日誌查看
服務器日誌會顯示詳細的錯誤信息和調試信息，幫助診斷問題。

## 開發指南

### 添加新的API提供商
1. 在`services/`目錄創建新的服務文件
2. 實現`VideoGenerationInterface`接口
3. 在`api_provider_manager.go`中註冊新服務
4. 更新配置結構和測試

### 擴展cURL功能
1. 在`curl_service.go`中添加新方法
2. 在`api_handler.go`中添加對應的HTTP端點
3. 編寫相應的測試用例

## 安全注意事項

- 不要在代碼中硬編碼API密鑰
- 使用環境變量或安全的配置管理
- 定期輪換API密鑰
- 監控API使用情況和異常活動

## 性能優化

- 使用連接池管理HTTP連接
- 實現請求緩存機制
- 監控API響應時間
- 設置合理的超時和重試策略

## 貢獻指南

1. Fork項目
2. 創建功能分支
3. 編寫測試
4. 提交Pull Request
5. 確保所有測試通過

## 許可證

本項目使用MIT許可證。詳見LICENSE文件。