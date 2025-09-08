# GenVideoSub v2 - AI 影片生成服務

一個基於 Golang 的 AI 影片生成服務，專門處理從外部 API 接收圖片和字幕資訊，透過 AI 生成影片並加上字幕後回傳的完整工作流程。

## 🚀 項目概述

本項目是 GenVideoSub 的 v2 版本，採用全新的架構設計，專注於提供高效能、可擴展的 AI 影片生成服務。

### 核心功能

- **影片生成任務接收**: 接收包含圖片、提示詞和字幕的 JSON 任務
- **AI 影片生成**: 整合外部 AI 服務將圖片轉換為影片
- **字幕處理**: 自動為生成的影片添加字幕
- **任務狀態管理**: 完整的任務生命週期追蹤
- **異步處理**: 支援大量並發任務處理

## 🏗️ 技術架構

### 技術棧
- **後端**: Go 1.21+ + Gin Web Framework
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

## 📁 項目結構

```
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