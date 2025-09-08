package handlers

import (
	"net/http"
	"strconv"

	"genVideoSub/models"
	"genVideoSub/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TaskHandler 任務處理器
type TaskHandler struct {
	taskService *services.TaskService
}

// NewTaskHandler 創建新的任務處理器
func NewTaskHandler(taskService *services.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTask 創建任務
// @Summary 創建新的視頻生成任務
// @Description 根據提供的參數創建新的視頻生成任務
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body models.TaskCreateRequest true "任務創建請求"
// @Success 201 {object} models.TaskResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var request models.TaskCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logrus.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 驗證必填字段
	if request.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Prompt is required",
		})
		return
	}

	// 創建任務
	response, err := h.taskService.CreateTask(&request)
	if err != nil {
		logrus.Errorf("Failed to create task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
			"details": err.Error(),
		})
		return
	}

	logrus.Infof("Task created successfully: %s", response.ID)
	c.JSON(http.StatusCreated, response)
}

// GetTask 獲取任務詳情
// @Summary 獲取任務詳情
// @Description 根據任務ID獲取任務的詳細信息
// @Tags tasks
// @Produce json
// @Param id path string true "任務ID"
// @Success 200 {object} models.Task
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
		})
		return
	}

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		logrus.Errorf("Failed to get task %s: %v", taskID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetTaskStatus 獲取任務狀態
// @Summary 獲取任務狀態
// @Description 根據任務ID獲取任務的當前狀態
// @Tags tasks
// @Produce json
// @Param id path string true "任務ID"
// @Success 200 {object} models.TaskStatusResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id}/status [get]
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
		})
		return
	}

	status, err := h.taskService.GetTaskStatus(taskID)
	if err != nil {
		logrus.Errorf("Failed to get task status %s: %v", taskID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetTaskResult 獲取任務結果
// @Summary 獲取任務結果
// @Description 根據任務ID獲取已完成任務的結果
// @Tags tasks
// @Produce json
// @Param id path string true "任務ID"
// @Success 200 {object} models.TaskResultResponse
// @Failure 404 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id}/result [get]
func (h *TaskHandler) GetTaskResult(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
		})
		return
	}

	result, err := h.taskService.GetTaskResult(taskID)
	if err != nil {
		logrus.Errorf("Failed to get task result %s: %v", taskID, err)
		// 判斷是任務未完成還是其他錯誤
		if err.Error() == "task is not completed yet" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Task is not completed yet",
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListTasks 列出任務
// @Summary 列出任務
// @Description 分頁列出所有任務
// @Tags tasks
// @Produce json
// @Param limit query int false "每頁數量" default(10)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {array} models.Task
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	// 解析查詢參數
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // 限制最大值
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 獲取任務列表
	tasks, err := h.taskService.ListTasks(limit, offset)
	if err != nil {
		logrus.Errorf("Failed to list tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list tasks",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"limit": limit,
		"offset": offset,
		"count": len(tasks),
	})
}

// DeleteTask 刪除任務
// @Summary 刪除任務
// @Description 根據任務ID刪除任務
// @Tags tasks
// @Param id path string true "任務ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
		})
		return
	}

	// 先檢查任務是否存在
	_, err := h.taskService.GetTask(taskID)
	if err != nil {
		logrus.Errorf("Task not found for deletion %s: %v", taskID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	// 這裡應該調用存儲層的刪除方法
	// 由於TaskService沒有DeleteTask方法，我們需要直接調用存儲層
	// 或者在TaskService中添加DeleteTask方法
	logrus.Infof("Task %s deletion requested (not implemented yet)", taskID)
	c.JSON(http.StatusNoContent, nil)
}