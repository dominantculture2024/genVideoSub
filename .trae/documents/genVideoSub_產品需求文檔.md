# genVideo&Sub 產品需求文檔

## 1. Product Overview
genVideo&Sub 是一個輕量化的AI視頻生成服務，採用Streamlit前端和Golang後端架構，專門整合fal.ai API進行視頻生成。
該服務主要解決需要批量處理視頻生成任務的需求，提供簡潔的Web界面和高效的後端處理能力。
目標是提供零依賴、易部署的AI視頻生成解決方案，支援大文件異步處理和批量任務管理。

## 2. Core Features

### 2.1 系統架構
本專案採用輕量化設計，無需複雜的用戶管理，主要提供以下核心功能：
- 任務管理：創建、查詢、管理視頻生成任務
- 文件處理：圖片下載、臨時存儲、文件清理
- AI集成：fal.ai API調用、異步處理、結果回調
- 前端界面：Streamlit測試界面、批量上傳、進度顯示

### 2.2 Feature Module
我們的 genVideo&Sub 服務包含以下主要功能：
1. **Streamlit前端**：多文件選擇、批量上傳、實時狀態顯示
2. **Golang API服務**：RESTful API、任務隊列、並發處理
3. **fal.ai集成**：AI視頻生成、參數配置、結果處理
4. **存儲管理**：內存存儲、JSON持久化、臨時文件管理

### 2.3 Module Details
| Module Name | Component | Feature description |
|-------------|-----------|---------------------|
| Streamlit前端 | 文件上傳器 | 多文件選擇、拖拽上傳、文件預覽 |
| Streamlit前端 | 參數配置器 | 字幕顏色、位置、內容設定 |
| Streamlit前端 | 狀態顯示器 | 實時任務狀態、進度條、結果展示 |
| Golang API | 任務管理器 | 創建任務、狀態查詢、任務列表 |
| Golang API | HTTP路由器 | RESTful API、請求處理、響應格式化 |
| Golang API | 並發處理器 | Goroutine Pool、Channel隊列、工作分配 |
| fal.ai集成 | API客戶端 | HTTP請求、認證管理、錯誤處理 |
| fal.ai集成 | 異步處理器 | 任務提交、狀態輪詢、結果獲取 |
| fal.ai集成 | 重試機制 | 指數退避、失敗重試、超時處理 |
| 存儲管理 | 內存存儲 | 任務狀態、臨時數據、快速訪問 |
| 存儲管理 | JSON持久化 | 數據備份、恢復機制、定期保存 |
| 存儲管理 | 文件清理器 | 臨時文件清理、過期檢查、空間管理 |

## 3. Core Process
**模組調用流程：**
用戶通過Streamlit界面使用genVideo&Sub服務的主要操作流程如下：

**視頻生成流程**：
1. 用戶在Streamlit界面選擇圖片文件
2. 配置字幕內容、顏色、位置等參數
3. 提交任務到Golang後端API
4. 系統創建任務並調用fal.ai API
5. 異步處理並實時顯示任務狀態
6. 完成後提供視頻下載連結

**批量處理流程**：
1. 用戶選擇多個圖片文件進行批量上傳
2. 系統創建多個並發任務
3. Goroutine Pool管理並發處理
4. 實時更新每個任務的處理狀態
5. 完成後批量下載所有生成的視頻

```mermaid
graph TD
  A[Streamlit前端] --> B[文件上傳]
  B --> C[參數配置]
  C --> D[Golang API]
  D --> E[任務創建]
  E --> F[fal.ai API]
  F --> G[異步處理]
  G --> H[狀態更新]
  H --> I[結果下載]
  
  D --> J[內存存儲]
  D --> K[JSON持久化]
  
  F --> L[圖片處理]
  F --> M[視頻生成]
  
  G --> N[並發管理]
  G --> O[錯誤處理]
```

## 4. Design Specifications

### 4.1 Design Style
本專案採用Streamlit原生設計風格，簡潔易用：
- 主色調：Streamlit預設藍色 (#FF6B6B)，輔助色：淺灰色 (#F8F9FA)
- 按鈕樣式：Streamlit原生按鈕，圓角設計
- 字體：系統預設字體，標題使用st.title，內容使用st.text
- 佈局風格：側邊欄配置，主區域顯示，響應式設計
- 組件風格：使用st.file_uploader、st.selectbox、st.progress等原生組件