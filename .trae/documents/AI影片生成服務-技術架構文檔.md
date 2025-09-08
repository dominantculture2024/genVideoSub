# AI 影片生成服務 - 技術架構文檔

## 1. Architecture design

```mermaid
graph TD
    A[用戶瀏覽器] --> B[React 前端應用]
    B --> C[Golang HTTP Server]
    C --> D[任務管理模組]
    D --> E[Redis 任務佇列]
    C --> F[影片處理模組]
    F --> G[外部 AI 服務 API]
    F --> H[模擬影片服務]
    F --> I[FFmpeg 字幕處理]
    C --> J[檔案存儲模組]
    J --> K[本地檔案系統]
    C --> L[回調通知模組]
    L --> M[外部回調端點]
    C --> N[配置管理模組]

    subgraph "前端層"
        B
    end

    subgraph "應用層"
        C
        D
        F
        J
        L
        N
    end

    subgraph "存儲層"
        E
        K
    end

    subgraph "外部服務"
        G
        M
    end

    subgraph "模擬服務"
        H
    end
```

## 2. Technology Description

* Frontend: React\@18 + TypeScript + Vite + Tailwind CSS

* Backend: Go\@1.21 + Gin\@1.9 + Redis\@7.0

* 影片處理: FFmpeg

* HTTP Client: Go 標準庫 net/http

* 任務佇列: Redis

* 日誌: logrus

## 3. Route definitions

### 3.1 前端路由

| Route     | Purpose      |
| --------- | ------------ |
| /         | 主頁面 - 任務提交和管理 |
| /tasks    | 任務列表頁面       |
| /settings | 設定頁面 - 模式切換   |

### 3.2 後端 API 路由

| Route                         | Purpose    |
| ----------------------------- | ---------- |
| POST /api/v1/generate-video   | 接收影片生成任務請求 |
| GET /api/v1/tasks/{id}/status | 查詢任務執行狀態   |
| GET /api/v1/tasks             | 獲取任務列表     |
| POST /api/v1/config/mode      | 切換運行模式     |
| GET /api/v1/config            | 獲取當前配置     |
| GET /health                   | 服務健康檢查端點   |
| GET /metrics                  | 服務監控指標端點   |

## 4. API definitions

### 4.1 Core API

#### 配置管理

```
GET /api/v1/config
```

Response:

| Param Name | Param Type | Description |
|------------|------------|--------------|
| mode       | string     | 當前運行模式（"mock" 或 "production"） |
| ai_service_url | string | AI 服務端點 URL |
| mock_video_path | string | 模擬影片檔案路徑 |

```
POST /api/v1/config/mode
```

Request:

| Param Name | Param Type | isRequired | Description |
|------------|------------|------------|-------------|
| mode       | string     | true       | 運行模式（"mock" 或 "production"） |

Response:

| Param Name | Param Type | Description |
|------------|------------|--------------|
| success    | boolean    | 操作是否成功 |
| message    | string     | 狀態訊息 |
| mode       | string     | 更新後的模式 |

#### 任務列表

```
GET /api/v1/tasks
```

Response:

| Param Name | Param Type | Description |
|------------|------------|--------------|
| tasks      | array      | 任務列表 |
| total      | integer    | 總任務數量 |
| page       | integer    | 當前頁碼 |
| limit      | integer    | 每頁數量 |

#### 影片生成任務

```
POST /api/v1/generate-video
```

Request:

| Param Name         | Param Type | isRequired | Description                       |
| ------------------ | ---------- | ---------- | --------------------------------- |
| id                 | integer    | true       | 任務唯一識別碼                           |
| image\_path        | string     | true       | 圖片檔案的 URL 路徑                      |
| subtitle\_color    | string     | true       | 字幕顏色（如 "white", "black"）          |
| subtitle\_position | string     | true       | 字幕位置（如 "bottom", "top", "center"） |
| subtitle           | string     | true       | 字幕內容文字                            |
| prompt             | string     | true       | AI 影片生成的提示詞                       |

Response:

| Param Name   | Param Type | Description                                          |
| ------------ | ---------- | ---------------------------------------------------- |
| status       | string     | 任務狀態（"pending", "processing", "completed", "failed"） |
| external\_id | string     | 內部生成的任務 ID                                           |
| message      | string     | 狀態描述訊息                                               |

Example Request:

```json
{
  "id": 1,
  "image_path": "http://your-domain.com/storage/video-task-images/image1.jpg",
  "subtitle_color": "white",
  "subtitle_position": "bottom",
  "subtitle": "歡迎來到我們的平台",
  "prompt": "創建一個專業的介紹影片，包含動畫效果"
}
```

Example Response:

```json
{
  "status": "pending",
  "external_id": "task_1704067200_abc123",
  "message": "任務已接收，正在處理中"
}
```

#### 任務狀態查詢

```
GET /api/v1/tasks/{id}/status
```

Response:

| Param Name  | Param Type | Description |
| ----------- | ---------- | ----------- |
| task\_id    | string     | 任務 ID       |
| status      | string     | 當前狀態        |
| progress    | integer    | 處理進度百分比     |
| created\_at | string     | 任務創建時間      |
| updated\_at | string     | 最後更新時間      |

#### 影片上傳回調（調用外部 API）

```
POST /video-tasks/{id}/completed
```

Request (Form Data):

| Param Name         | Param Type | isRequired | Description                        |
| ------------------ | ---------- | ---------- | ---------------------------------- |
| video\_file        | file       | true       | 完成的影片檔案（mp4/mpeg/quicktime，≤100MB） |
| external\_task\_id | string     | false      | 外部任務 ID                            |

## 5. Server architecture diagram

```mermaid
graph TD
    A[HTTP Handler Layer] --> B[Service Layer]
    B --> C[Repository Layer]
    C --> D[(Redis)]
    C --> E[(File System)]
    
    B --> F[External API Client]
    F --> G[AI Video Service]
    
    B --> H[Video Processing]
    H --> I[FFmpeg]
    
    B --> J[Callback Client]
    J --> K[External Callback API]
    
    B --> L[Config Service]
    L --> M[Mock Video Service]
    
    B --> N[Frontend Static Files]

    subgraph Server
        A
        B
        C
        F
        H
        J
        L
        N
    end
    
    subgraph Frontend
        O[React App]
        P[Task Management]
        Q[Mode Switcher]
        R[Video Preview]
    end
    
    O --> A
    P --> A
    Q --> A
    R --> A
```

## 6. Data model

### 6.1 Data model definition

```mermaid
erDiagram
    TASK {
        string id PK
        integer external_id
        string image_path
        string subtitle_color
        string subtitle_position
        string subtitle
        string prompt
        string status
        string video_path
        datetime created_at
        datetime updated_at
    }
    
    TASK_LOG {
        string id PK
        string task_id FK
        string level
        string message
        datetime created_at
    }
    
    TASK ||--o{ TASK_LOG : logs
    
    CONFIG {
        string key PK
        string value
        string type
        datetime updated_at
    }
```

### 6.2 Data Definition Language

由於使用 Redis 作為主要存儲，數據結構以 JSON 格式存儲：

#### 任務資料結構 (Redis Hash)

```go
// Redis Key: task:{task_id}
type Task struct {
    ID              string    `json:"id"`
    ExternalID      int       `json:"external_id"`
    ImagePath       string    `json:"image_path"`
    SubtitleColor   string    `json:"subtitle_color"`
    SubtitlePosition string   `json:"subtitle_position"`
    Subtitle        string    `json:"subtitle"`
    Prompt          string    `json:"prompt"`
    Status          string    `json:"status"` // pending, processing, completed, failed
    VideoPath       string    `json:"video_path,omitempty"`
    ErrorMessage    string    `json:"error_message,omitempty"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

#### 任務佇列結構 (Redis List)

```go
// Redis Key: task_queue
// 存儲待處理的任務 ID 列表
// LPUSH task_queue {task_id}
// RPOP task_queue -> {task_id}
```

#### 任務狀態索引 (Redis Set)

```go
// Redis Key: tasks_by_status:{status}
// 按狀態分組的任務 ID 集合
// SADD tasks_by_status:pending {task_id}
// SMEMBERS tasks_by_status:completed
```

#### 初始化配置

```go
// 服務啟動時的 Redis 配置
const (
    TaskKeyPrefix = "task:"
    TaskQueueKey = "task_queue"
    TaskStatusPrefix = "tasks_by_status:"
    TaskTTL = 24 * time.Hour // 任務資料保留 24 小時
)

// 狀態常數
const (
    StatusPending    = "pending"
    StatusProcessing = "processing"
    StatusCompleted  = "completed"
    StatusFailed     = "failed"
)
```

#### Redis 初始化腳本

```
-- 任務狀態索引 (Redis Set)
SADD tasks_by_status:pending {task_id}
SADD tasks_by_status:processing {task_id}
SADD tasks_by_status:completed {task_id}
SADD tasks_by_status:failed {task_id}

-- 系統配置 (Redis Hash)
HSET config:system mode "production"
HSET config:system ai_service_url "https://api.example.com/generate-video"
HSET config:system mock_video_path "/assets/mock-video.mp4"
HSET config:system callback_url "https://external-api.com/video-tasks"

-- Redis 配置常數
SET config:max_queue_size 1000
SET config:task_timeout 3600
SET config:retry_attempts 3

-- 前端配置
HSET config:frontend theme "light"
HSET config:frontend auto_refresh "true"
HSET config:frontend refresh_interval "5000"
```

