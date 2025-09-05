# GenVideoSub - AI視頻生成服務

基於 fal.ai Kling Video API 的智能視頻生成平台，提供 Streamlit 前端界面和 Golang 後端服務。

## 🚀 功能特點

- 🎬 **AI視頻生成**: 基於 fal.ai Kling Video API
- 📝 **提示詞驅動**: 支持文本到視頻生成
- 🖼️ **圖片參考**: 可選的參考圖片上傳
- ⚙️ **參數調節**: 支持時長、寬高比、CFG Scale 等參數
- 📊 **任務管理**: 完整的任務創建、查詢、管理功能
- 🔄 **異步處理**: 後台異步任務處理機制
- 💾 **數據持久化**: JSON 文件存儲

## 🏗️ 技術架構

### 後端 (Golang)
- **框架**: Gin Web Framework
- **存儲**: JSON 文件存儲
- **API整合**: fal.ai Kling Video API
- **並發處理**: Goroutine 工作池

### 前端 (Streamlit)
- **界面**: Streamlit Web 應用
- **組件**: 模塊化 UI 組件
- **API客戶端**: HTTP 請求處理

## 📋 系統要求

### 後端要求
- Go 1.21 或更高版本
- fal.ai API Key

### 前端要求
- Python 3.8 或更高版本
- pip 包管理器

## 🛠️ 安裝和設置

### 1. 克隆項目
```bash
git clone <repository-url>
cd genVideoSub
```

### 2. 後端設置

#### 安裝 Go
如果尚未安裝 Go，請從 [官方網站](https://golang.org/dl/) 下載並安裝。

#### 安裝依賴
```bash
cd backend
go mod tidy
```

#### 配置環境變量
創建 `backend/.env` 文件：
```env
FAL_API_KEY=your_fal_api_key_here
SERVER_PORT=8080
STORAGE_PATH=./data
```

#### 啟動後端服務
```bash
go run main.go
```

後端服務將在 `http://localhost:8080` 啟動。

### 3. 前端設置

#### 安裝 Python 依賴
```bash
cd frontend
pip install -r requirements.txt
```

#### 配置環境變量
編輯 `frontend/.env` 文件：
```env
BACKEND_URL=http://localhost:8080
FAL_API_KEY=your_fal_api_key_here
```

#### 啟動前端應用
```bash
streamlit run app.py
```

前端應用將在 `http://localhost:8501` 啟動。

## 🔑 獲取 fal.ai API Key

1. 訪問 [fal.ai](https://fal.ai/)
2. 註冊並登錄賬戶
3. 在控制台中生成 API Key
4. 將 API Key 添加到環境變量中

## 📖 使用說明

### 創建視頻生成任務

1. 打開前端應用 (`http://localhost:8501`)
2. 在「創建任務」頁面填寫：
   - **提示詞**: 描述想要生成的視頻內容
   - **參考圖片**: (可選) 上傳參考圖片
   - **視頻時長**: 1-10秒
   - **寬高比**: 16:9, 9:16, 1:1
   - **負面提示詞**: (可選) 不希望出現的內容
   - **CFG Scale**: 控制生成內容與提示詞的相符程度
3. 點擊「創建任務」

### 查詢任務狀態

1. 在「狀態查詢」頁面輸入任務 ID
2. 點擊「查詢狀態」或「查詢結果」
3. 查看任務進度和生成結果

### 管理任務

1. 在「任務管理」頁面查看所有任務
2. 可以查看任務詳情、下載結果或刪除任務

## 🧪 測試策略

本項目採用分層測試架構，確保系統的可靠性和穩定性。詳細的測試策略請參考 [TESTING_STRATEGY.md](docs/TESTING_STRATEGY.md)。

### 測試類型
- **單元測試**: 測試個別組件和函數
- **整合測試**: 測試組件間的交互
- **端到端測試**: 測試完整的用戶流程
- **性能測試**: 測試系統負載和響應時間

### 運行測試

#### 後端測試
```bash
cd backend

# 運行所有測試
make test

# 運行單元測試
make test-unit

# 運行整合測試
make test-integration

# 生成覆蓋率報告
make test-coverage

# 快速查看覆蓋率
make test-coverage-quick
```

#### 前端測試
```bash
cd frontend

# 安裝測試依賴
pip install -r requirements-test.txt

# 運行所有測試
pytest tests/ -v

# 運行單元測試
pytest tests/unit/ -v

# 運行整合測試
pytest tests/integration/ -v

# 運行端到端測試
pytest tests/e2e/ -v

# 生成覆蓋率報告
pytest tests/ --cov=. --cov-report=html
```

### Mock 測試

項目支持 fal.ai API 的 Mock 測試，無需真實 API 調用：

```bash
# 使用 Mock 模式運行測試
export FAL_MOCK_MODE=true
go test ./...
```

## 🔧 API 接口

### 任務管理
- `POST /api/tasks` - 創建任務
- `GET /api/tasks` - 獲取任務列表
- `GET /api/tasks/{id}` - 獲取任務信息
- `GET /api/tasks/{id}/status` - 獲取任務狀態
- `GET /api/tasks/{id}/result` - 獲取任務結果
- `DELETE /api/tasks/{id}` - 刪除任務

### 文件管理
- `POST /api/files/upload` - 上傳文件
- `GET /api/files/{id}` - 獲取文件
- `GET /api/files/{id}/info` - 獲取文件信息
- `DELETE /api/files/{id}` - 刪除文件

## 📁 項目結構

```
genVideoSub/
├── backend/                 # Golang 後端
│   ├── config/             # 配置管理
│   ├── handlers/           # HTTP 處理器
│   ├── models/             # 數據模型
│   ├── services/           # 業務邏輯
│   ├── storage/            # 存儲層
│   ├── main.go             # 主程序
│   └── go.mod              # Go 模塊
├── frontend/               # Streamlit 前端
│   ├── components/         # UI 組件
│   ├── utils/              # 工具函數
│   ├── app.py              # 主應用
│   └── requirements.txt    # Python 依賴
├── docs/                   # 文檔
└── README.md               # 項目說明
```

## 🐛 故障排除

### 常見問題

1. **後端無法啟動**
   - 檢查 Go 是否正確安裝
   - 確認 fal.ai API Key 是否正確設置
   - 檢查端口 8080 是否被占用

2. **前端無法連接後端**
   - 確認後端服務是否正在運行
   - 檢查 `frontend/.env` 中的 `BACKEND_URL` 設置

3. **任務創建失敗**
   - 檢查 fal.ai API Key 是否有效
   - 確認網絡連接正常
   - 查看後端日誌獲取詳細錯誤信息

### 日誌查看

後端日誌會輸出到控制台，包含詳細的錯誤信息和調試信息。

## 🤝 貢獻

歡迎提交 Issue 和 Pull Request 來改進項目。

## 📄 許可證

本項目採用 MIT 許可證。

## 🔗 相關鏈接

- [fal.ai 官網](https://fal.ai/)
- [Streamlit 文檔](https://docs.streamlit.io/)
- [Gin 框架文檔](https://gin-gonic.com/)