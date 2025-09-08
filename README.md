# GenVideoSub v2 - AI 影片生成服務

一個基於 Golang 的 AI 影片生成服務，專門處理從外部 API 接收圖片和字幕資訊，透過 AI 生成影片並加上字幕後回傳的完整工作流程。

本項目是 GenVideoSub 的 v2 版本，採用全新的架構設計，專注於提供高效能、可擴展的 AI 影片生成服務，並整合了基於 fal.ai Kling Video API 的智能視頻生成平台，提供 Streamlit 前端界面和 Golang 後端服務。

## 🚀 項目概述

### 核心功能

- **影片生成任務接收**: 接收包含圖片、提示詞和字幕的 JSON 任務
- **AI 影片生成**: 整合外部 AI 服務將圖片轉換為影片
- **字幕處理**: 自動為生成的影片添加字幕
- **任務狀態管理**: 完整的任務生命週期追蹤
- **異步處理**: 支援大量並發任務處理
- **測試前端**: 支援模擬模式和正式模式的測試界面

## 🏗️ 技術架構

### 技術棧
- **後端**: Go 1.21+ + Gin Web Framework
- **前端**: React 18 + TypeScript + Vite + Tailwind CSS
- **數據存儲**: Redis (任務佇列 + 狀態管理)
- **影片處理**: FFmpeg
- **部署**: Docker + Docker Compose

### 系統架構
```
外部 API
    ↓
Gin HTTP Server (Port 8080)
    ↓
Redis 任務佇列
    ↓
後台處理器 (Goroutines)
    ↓
AI 影片生成服務
    ↓
FFmpeg 字幕處理
    ↓
回傳完成影片
```

## 📋 API 設計

### 核心端點

#### 1. 接收影片生成任務
```http
POST /api/v1/generate-video
Content-Type: application/json

{
  "id": "task-123",
  "image_path": "https://example.com/image.jpg",
  "subtitle_color": "white",
  "subtitle_position": "bottom",
  "subtitle": "這是字幕內容",
  "prompt": "生成一個美麗的風景影片"
}
```

#### 2. 查詢任務狀態
```http
GET /api/v1/tasks/{id}/status

Response:
{
  "id": "task-123",
  "status": "processing",
  "progress": 50,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

#### 3. 影片完成回調
```http
POST /video-tasks/{id}/completed
Content-Type: multipart/form-data

video_file: [影片檔案]
external_task_id: "optional-external-id"
```

### 任務狀態
- `pending`: 等待處理
- `processing`: 處理中
- `completed`: 已完成
- `failed`: 處理失敗

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
>>>>>>> d7c80a8135ea0d73ceeb0ea9c68c9626fda4b8a8

## 📁 項目結構

```
<<<<<<< HEAD
genVideoSub-v2/
├── .trae/
│   └── documents/              # 項目文檔
│       ├── AI影片生成服務-產品需求文檔.md
│       └── AI影片生成服務-技術架構文檔.md
├── cmd/                        # 應用程式入口
├── internal/                   # 內部套件
│   ├── handlers/              # HTTP 處理器
│   ├── services/              # 業務邏輯服務
│   ├── models/                # 數據模型
│   ├── queue/                 # 任務佇列
│   └── config/                # 配置管理
├── pkg/                       # 公共套件
├── scripts/                   # 部署腳本
├── docker-compose.yml         # Docker Compose 配置
├── Dockerfile                 # Docker 構建文件
├── go.mod                     # Go 模組定義
├── go.sum                     # Go 依賴鎖定
├── .gitignore                 # Git 忽略文件
└── README.md                  # 項目說明
```

## 🚀 快速開始

### 前置需求
- Go 1.21+
- Redis 6.0+
- FFmpeg
- Docker (可選)

### 本地開發

1. **克隆項目**
```bash
git clone https://github.com/dominantculture2024/genVideoSub.git
cd genVideoSub-v2
```

2. **安裝依賴**
```bash
go mod tidy
```

3. **啟動 Redis**
```bash
# 使用 Docker
docker run -d -p 6379:6379 redis:alpine

# 或使用本地安裝的 Redis
redis-server
```

4. **運行服務**
```bash
go run cmd/server/main.go
```

### Docker 部署

```bash
# 構建並啟動所有服務
docker-compose up -d

# 查看服務狀態
docker-compose ps

# 查看日誌
docker-compose logs -f
```

## 📖 開發狀態

### 當前版本: v2.0.0-alpha

**已完成**:
- ✅ 產品需求文檔
- ✅ 技術架構文檔
- ✅ 項目結構規劃

**開發中**:
- 🔄 核心 API 實現
- 🔄 Redis 任務佇列
- 🔄 影片處理服務
- 🔄 Docker 配置

**計劃中**:
- ⏳ 單元測試
- ⏳ 整合測試
- ⏳ 性能優化
- ⏳ 監控和日誌

## 🤝 貢獻

歡迎提交 Issue 和 Pull Request 來改進這個項目。

### 開發分支
- `main` - 生產環境分支
- `develop` - 開發分支
- `feature/*` - 功能開發分支

## 📄 許可證

本項目採用 MIT 許可證 - 詳見 [LICENSE](LICENSE) 文件

## 🔗 相關連結

- [v1 版本](https://github.com/dominantculture2024/genVideoSub) - 基於 Streamlit + Golang 的原始版本
- [產品需求文檔](.trae/documents/AI影片生成服務-產品需求文檔.md)
- [技術架構文檔](.trae/documents/AI影片生成服務-技術架構文檔.md)

---

**注意**: 這是 v2 版本的重新設計，專注於企業級的可擴展性和性能優化。
=======
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
>>>>>>> d7c80a8135ea0d73ceeb0ea9c68c9626fda4b8a8
