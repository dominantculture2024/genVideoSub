# GenVideoSub - 視頻字幕生成系統

一個基於 Streamlit 前端和 Golang 後端的輕量級視頻字幕生成系統，整合 fal.ai API 提供 AI 驅動的字幕生成功能。

## 🚀 功能特色

- **輕量級架構**: 無需外部數據庫，使用內存存儲 + JSON 文件持久化
- **現代化界面**: 基於 Streamlit 的直觀 Web 界面
- **高性能後端**: Golang + Gin 框架，支持並發處理
- **AI 字幕生成**: 整合 fal.ai API，提供高質量字幕生成
- **任務管理**: 完整的任務狀態追蹤和管理系統
- **容器化部署**: 支持 Docker 和 Docker Compose 部署

## 🏗️ 系統架構

```
Streamlit Frontend (Port 8501)
        ↓
    HTTP API
        ↓
Golang Backend (Port 8080)
        ↓
   fal.ai API
```

### 技術棧

**前端**:
- Streamlit 1.28+
- Python 3.9+
- Requests, Pandas, Pillow

**後端**:
- Golang 1.21+
- Gin Web Framework
- 內存存儲 + JSON 持久化
- Goroutine 並發處理

**AI 服務**:
- fal.ai API 整合
- 支持多種視頻字幕生成模型

## 📁 專案結構

```
genVideoSub/
├── backend/                 # Golang 後端
│   ├── config/             # 配置管理
│   ├── handlers/           # HTTP 處理器
│   ├── services/           # 業務邏輯服務
│   ├── models/             # 數據模型
│   ├── storage/            # 存儲層
│   ├── utils/              # 工具函數
│   ├── data/               # JSON 數據文件
│   ├── temp/               # 臨時文件
│   ├── go.mod              # Go 模組定義
│   └── main.go             # 程序入口
├── frontend/               # Streamlit 前端
│   ├── components/         # UI 組件
│   ├── utils/              # 工具模組
│   ├── app.py              # 主應用
│   └── requirements.txt    # Python 依賴
├── docs/                   # 文檔
├── scripts/                # 部署腳本
├── docker-compose.yml      # Docker Compose 配置
├── .gitignore              # Git 忽略文件
└── README.md               # 專案說明
```

## 🚀 快速開始

### 前置需求

- Go 1.21+
- Python 3.9+
- fal.ai API Key

### 本地開發

1. **克隆專案**
```bash
git clone https://github.com/dominantculture2024/genVideoSub.git
cd genVideoSub
```

2. **設置後端**
```bash
cd backend
go mod tidy
go run main.go
```

3. **設置前端**
```bash
cd frontend
pip install -r requirements.txt
streamlit run app.py
```

4. **訪問應用**
- 前端界面: http://localhost:8501
- 後端API: http://localhost:8080

### Docker 部署

```bash
# 構建並啟動所有服務
docker-compose up -d

# 查看服務狀態
docker-compose ps

# 查看日誌
docker-compose logs -f
```

## 📖 API 文檔

### 核心端點

- `GET /api/health` - 健康檢查
- `POST /api/tasks` - 創建新任務
- `GET /api/tasks/:id/status` - 查詢任務狀態
- `GET /api/tasks` - 獲取任務列表
- `POST /api/tasks/:id/completed` - 上傳完成的視頻
- `DELETE /api/tasks/:id` - 刪除任務
- `GET /api/metrics` - 系統指標

### 任務狀態

- `pending` - 等待處理
- `processing` - 處理中
- `completed` - 已完成
- `failed` - 處理失敗

## 🔧 配置

### 後端配置 (backend/config/app.json)

```json
{
  "server": {
    "port": 8080,
    "host": "0.0.0.0"
  },
  "fal_ai": {
    "api_key": "your-fal-ai-api-key",
    "base_url": "https://fal.run/fal-ai"
  },
  "storage": {
    "data_path": "./data",
    "temp_path": "./temp"
  },
  "worker": {
    "pool_size": 5,
    "queue_size": 100
  }
}
```

### 前端配置

在 Streamlit 界面中可以動態配置 API 服務器地址和連接設置。

## 🧪 測試

```bash
# 後端測試
cd backend
go test ./...

# 前端測試
cd frontend
python -m pytest tests/
```

## 📊 監控

系統提供以下監控指標：

- 任務處理統計
- 系統資源使用情況
- API 響應時間
- 錯誤率統計

## 🤝 貢獻

1. Fork 專案
2. 創建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 開啟 Pull Request

## 📝 開發分支

- `main` - 生產環境分支
- `dev` - 開發分支
- `feature/*` - 功能開發分支

## 📄 許可證

本專案採用 MIT 許可證 - 詳見 [LICENSE](LICENSE) 文件

## 🆘 支持

如有問題或建議，請：

1. 查看 [Issues](https://github.com/dominantculture2024/genVideoSub/issues)
2. 創建新的 Issue
3. 聯繫維護團隊

## 🔄 版本歷史

- v1.0.0 - 初始版本
  - 基礎架構實現
  - Streamlit 前端界面
  - Golang 後端 API
  - fal.ai 整合
  - Docker 支持