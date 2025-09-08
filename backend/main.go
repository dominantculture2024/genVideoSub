package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"genVideoSub/config"
	"genVideoSub/handlers"
	"genVideoSub/services"
	"genVideoSub/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 設置日誌格式
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// 加載測試配置
	testConfig := config.LoadTestConfig()
	logrus.Infof("Environment: %s, Mock enabled: %v", testConfig.GetEnvironment(), testConfig.IsMockEnabled())

	// 創建服務工廠
	serviceFactory := services.NewServiceFactory(testConfig)

	// 初始化存儲
	jsonStorage, err := storage.NewJSONStorage(testConfig.TestDataPath)
	if err != nil {
		logrus.Fatalf("Failed to initialize storage: %v", err)
	}
	logrus.Info("JSON storage initialized")

	// 獲取fal.ai服務（自動選擇mock或真實服務）
	falService := serviceFactory.GetFalAIService()
	logrus.Infof("FalAI service initialized (type: %s)", testConfig.GetEnvironment())

	// 初始化任務服務
	taskService := services.NewTaskService(jsonStorage, falService, testConfig.MaxConcurrent)
	logrus.Info("Task service initialized")

	// 啟動工作協程池
	taskService.StartWorkers()
	logrus.Infof("Started %d task workers", testConfig.MaxConcurrent)

	// 初始化處理器
	taskHandler := handlers.NewTaskHandler(taskService)
	fileHandler := handlers.NewFileHandler(jsonStorage, "./temp", 104857600) // 100MB
	logrus.Info("Handlers initialized")

	// 創建 Gin 路由器
	if testConfig.GetEnvironment() == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.Default()

	// 配置CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(corsConfig))

	// 健康檢查端點
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"message":   "genVideoSub backend is running",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// API路由組
	api := r.Group("/api")
	{
		// 任務相關路由
		tasks := api.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("", taskHandler.ListTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.GET("/:id/status", taskHandler.GetTaskStatus)
			tasks.GET("/:id/result", taskHandler.GetTaskResult)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
		}

		// 文件相關路由
		files := api.Group("/files")
		{
			files.POST("/upload", fileHandler.UploadFile)
			files.GET("/:id", fileHandler.GetFile)
			files.GET("/:id/info", fileHandler.GetFileInfo)
			files.DELETE("/:id", fileHandler.DeleteFile)
		}
	}

	// 創建HTTP服務器
	addr := "localhost:8080"
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 在goroutine中啟動服務器
	go func() {
		logrus.Infof("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中斷信號以優雅地關閉服務器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// 停止工作協程池
	taskService.StopWorkers()

	// 給服務器5秒時間完成現有請求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}