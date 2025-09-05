package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"genVideoSub/models"
	"genVideoSub/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// FileHandler 文件處理器
type FileHandler struct {
	storage   storage.Storage
	uploadDir string
	maxSize   int64
}

// NewFileHandler 創建新的文件處理器
func NewFileHandler(storage storage.Storage, uploadDir string, maxSize int64) *FileHandler {
	// 創建上傳目錄
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logrus.Errorf("Failed to create upload directory: %v", err)
	}

	return &FileHandler{
		storage:   storage,
		uploadDir: uploadDir,
		maxSize:   maxSize,
	}
}

// UploadFile 上傳文件
// @Summary 上傳文件
// @Description 上傳圖片文件用於視頻生成
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要上傳的文件"
// @Success 200 {object} models.TempFile
// @Failure 400 {object} map[string]string
// @Failure 413 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/files/upload [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	// 獲取上傳的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logrus.Errorf("Failed to get uploaded file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	// 檢查文件大小
	if header.Size > h.maxSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": fmt.Sprintf("File too large. Maximum size is %d bytes", h.maxSize),
		})
		return
	}

	// 檢查文件類型
	if !h.isValidImageType(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file type. Only JPG, JPEG, PNG, GIF, BMP, WEBP are allowed",
		})
		return
	}

	// 生成文件ID和路徑
	fileID := uuid.New().String()
	fileExt := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", fileID, fileExt)
	filePath := filepath.Join(h.uploadDir, filename)

	// 創建目標文件
	dst, err := os.Create(filePath)
	if err != nil {
		logrus.Errorf("Failed to create file %s: %v", filePath, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
			"details": err.Error(),
		})
		return
	}
	defer dst.Close()

	// 複製文件內容
	if _, err := io.Copy(dst, file); err != nil {
		logrus.Errorf("Failed to copy file content: %v", err)
		os.Remove(filePath) // 清理失敗的文件
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
			"details": err.Error(),
		})
		return
	}

	// 創建臨時文件記錄
	tempFile := &models.TempFile{
		ID:           fileID,
		OriginalName: header.Filename,
		Filename:     filename,
		FilePath:     filePath,
		Size:         header.Size,
		MimeType:     h.getMimeType(fileExt),
		CreatedAt:    time.Now(),
	}

	// 保存文件記錄到存儲
	if err := h.storage.SaveTempFile(tempFile); err != nil {
		logrus.Errorf("Failed to save temp file record: %v", err)
		os.Remove(filePath) // 清理文件
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file record",
			"details": err.Error(),
		})
		return
	}

	logrus.Infof("File uploaded successfully: %s (%s)", fileID, header.Filename)
	c.JSON(http.StatusOK, tempFile)
}

// GetFile 獲取文件
// @Summary 獲取文件
// @Description 根據文件ID獲取文件內容
// @Tags files
// @Param id path string true "文件ID"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/files/{id} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File ID is required",
		})
		return
	}

	// 獲取文件記錄
	tempFile, err := h.storage.GetTempFile(fileID)
	if err != nil {
		logrus.Errorf("Failed to get temp file %s: %v", fileID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	// 檢查文件是否存在
	if _, err := os.Stat(tempFile.FilePath); os.IsNotExist(err) {
		logrus.Errorf("File does not exist: %s", tempFile.FilePath)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found on disk",
		})
		return
	}

	// 設置響應頭
	c.Header("Content-Type", tempFile.MimeType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", tempFile.OriginalName))
	c.Header("Cache-Control", "public, max-age=3600")

	// 返回文件內容
	c.File(tempFile.FilePath)
}

// GetFileInfo 獲取文件信息
// @Summary 獲取文件信息
// @Description 根據文件ID獲取文件的詳細信息
// @Tags files
// @Produce json
// @Param id path string true "文件ID"
// @Success 200 {object} models.TempFile
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/files/{id}/info [get]
func (h *FileHandler) GetFileInfo(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File ID is required",
		})
		return
	}

	// 獲取文件記錄
	tempFile, err := h.storage.GetTempFile(fileID)
	if err != nil {
		logrus.Errorf("Failed to get temp file %s: %v", fileID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	c.JSON(http.StatusOK, tempFile)
}

// DeleteFile 刪除文件
// @Summary 刪除文件
// @Description 根據文件ID刪除文件
// @Tags files
// @Param id path string true "文件ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File ID is required",
		})
		return
	}

	// 獲取文件記錄
	tempFile, err := h.storage.GetTempFile(fileID)
	if err != nil {
		logrus.Errorf("Failed to get temp file %s: %v", fileID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	// 刪除物理文件
	if err := os.Remove(tempFile.FilePath); err != nil && !os.IsNotExist(err) {
		logrus.Warnf("Failed to delete physical file %s: %v", tempFile.FilePath, err)
	}

	// 刪除文件記錄
	if err := h.storage.DeleteTempFile(fileID); err != nil {
		logrus.Errorf("Failed to delete temp file record %s: %v", fileID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file record",
		})
		return
	}

	logrus.Infof("File deleted successfully: %s", fileID)
	c.Status(http.StatusNoContent)
}

// isValidImageType 檢查是否為有效的圖片類型
func (h *FileHandler) isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// getMimeType 根據文件擴展名獲取MIME類型
func (h *FileHandler) getMimeType(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}