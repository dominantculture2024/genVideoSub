package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"genVideoSub/models"
	"github.com/sirupsen/logrus"
)

// Storage 存儲接口
type Storage interface {
	SaveTask(task *models.Task) error
	GetTask(taskID string) (*models.Task, error)
	ListTasks(limit, offset int) ([]*models.Task, error)
	DeleteTask(taskID string) error
	SaveTempFile(file *models.TempFile) error
	GetTempFile(fileID string) (*models.TempFile, error)
	DeleteTempFile(fileID string) error
}

// JSONStorage JSON文件存儲實現
type JSONStorage struct {
	dataDir   string
	tasksDir  string
	filesDir  string
	mutex     sync.RWMutex
	taskCache map[string]*models.Task
	fileCache map[string]*models.TempFile
}

// NewJSONStorage 創建新的JSON存儲實例
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	tasksDir := filepath.Join(dataDir, "tasks")
	filesDir := filepath.Join(dataDir, "files")

	// 創建目錄
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tasks directory: %w", err)
	}
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create files directory: %w", err)
	}

	storage := &JSONStorage{
		dataDir:   dataDir,
		tasksDir:  tasksDir,
		filesDir:  filesDir,
		taskCache: make(map[string]*models.Task),
		fileCache: make(map[string]*models.TempFile),
	}

	// 加載現有數據到緩存
	if err := storage.loadCache(); err != nil {
		logrus.Warnf("Failed to load cache: %v", err)
	}

	return storage, nil
}

// SaveTask 保存任務
func (js *JSONStorage) SaveTask(task *models.Task) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	// 更新緩存
	js.taskCache[task.ID] = task

	// 保存到文件
	filePath := filepath.Join(js.tasksDir, fmt.Sprintf("%s.json", task.ID))
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	logrus.Debugf("Task %s saved to %s", task.ID, filePath)
	return nil
}

// GetTask 獲取任務
func (js *JSONStorage) GetTask(taskID string) (*models.Task, error) {
	js.mutex.RLock()
	defer js.mutex.RUnlock()

	// 先從緩存查找
	if task, exists := js.taskCache[taskID]; exists {
		return task, nil
	}

	// 從文件加載
	filePath := filepath.Join(js.tasksDir, fmt.Sprintf("%s.json", taskID))
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// 更新緩存
	js.taskCache[taskID] = &task
	return &task, nil
}

// ListTasks 列出任務
func (js *JSONStorage) ListTasks(limit, offset int) ([]*models.Task, error) {
	js.mutex.RLock()
	defer js.mutex.RUnlock()

	// 從緩存獲取所有任務
	tasks := make([]*models.Task, 0, len(js.taskCache))
	for _, task := range js.taskCache {
		tasks = append(tasks, task)
	}

	// 按創建時間排序（最新的在前）
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	// 應用分頁
	if offset >= len(tasks) {
		return []*models.Task{}, nil
	}

	end := offset + limit
	if end > len(tasks) {
		end = len(tasks)
	}

	return tasks[offset:end], nil
}

// DeleteTask 刪除任務
func (js *JSONStorage) DeleteTask(taskID string) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	// 從緩存刪除
	delete(js.taskCache, taskID)

	// 刪除文件
	filePath := filepath.Join(js.tasksDir, fmt.Sprintf("%s.json", taskID))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete task file: %w", err)
	}

	logrus.Debugf("Task %s deleted", taskID)
	return nil
}

// SaveTempFile 保存臨時文件信息
func (js *JSONStorage) SaveTempFile(file *models.TempFile) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	// 更新緩存
	js.fileCache[file.ID] = file

	// 保存到文件
	filePath := filepath.Join(js.filesDir, fmt.Sprintf("%s.json", file.ID))
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal temp file: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	logrus.Debugf("Temp file %s saved to %s", file.ID, filePath)
	return nil
}

// GetTempFile 獲取臨時文件信息
func (js *JSONStorage) GetTempFile(fileID string) (*models.TempFile, error) {
	js.mutex.RLock()
	defer js.mutex.RUnlock()

	// 先從緩存查找
	if file, exists := js.fileCache[fileID]; exists {
		return file, nil
	}

	// 從文件加載
	filePath := filepath.Join(js.filesDir, fmt.Sprintf("%s.json", fileID))
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("temp file not found: %s", fileID)
		}
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}

	var file models.TempFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to unmarshal temp file: %w", err)
	}

	// 更新緩存
	js.fileCache[fileID] = &file
	return &file, nil
}

// DeleteTempFile 刪除臨時文件信息
func (js *JSONStorage) DeleteTempFile(fileID string) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	// 從緩存刪除
	delete(js.fileCache, fileID)

	// 刪除文件
	filePath := filepath.Join(js.filesDir, fmt.Sprintf("%s.json", fileID))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete temp file: %w", err)
	}

	logrus.Debugf("Temp file %s deleted", fileID)
	return nil
}

// loadCache 加載現有數據到緩存
func (js *JSONStorage) loadCache() error {
	// 加載任務
	taskFiles, err := filepath.Glob(filepath.Join(js.tasksDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to glob task files: %w", err)
	}

	for _, filePath := range taskFiles {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			logrus.Warnf("Failed to read task file %s: %v", filePath, err)
			continue
		}

		var task models.Task
		if err := json.Unmarshal(data, &task); err != nil {
			logrus.Warnf("Failed to unmarshal task file %s: %v", filePath, err)
			continue
		}

		js.taskCache[task.ID] = &task
	}

	// 加載臨時文件
	fileFiles, err := filepath.Glob(filepath.Join(js.filesDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to glob temp files: %w", err)
	}

	for _, filePath := range fileFiles {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			logrus.Warnf("Failed to read temp file %s: %v", filePath, err)
			continue
		}

		var file models.TempFile
		if err := json.Unmarshal(data, &file); err != nil {
			logrus.Warnf("Failed to unmarshal temp file %s: %v", filePath, err)
			continue
		}

		js.fileCache[file.ID] = &file
	}

	logrus.Infof("Loaded %d tasks and %d temp files to cache", len(js.taskCache), len(js.fileCache))
	return nil
}

// CleanupExpiredFiles 清理過期的臨時文件
func (js *JSONStorage) CleanupExpiredFiles() error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	now := time.Now()
	expiredFiles := make([]string, 0)

	for fileID, file := range js.fileCache {
		if now.Sub(file.CreatedAt) > 24*time.Hour { // 24小時過期
			expiredFiles = append(expiredFiles, fileID)
		}
	}

	for _, fileID := range expiredFiles {
		if err := js.DeleteTempFile(fileID); err != nil {
			logrus.Warnf("Failed to delete expired file %s: %v", fileID, err)
		}
	}

	if len(expiredFiles) > 0 {
		logrus.Infof("Cleaned up %d expired temp files", len(expiredFiles))
	}

	return nil
}