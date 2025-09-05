package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化Gin路由器
	r := gin.Default()

	// 健康檢查端點
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "genVideoSub API is running",
		})
	})

	// 啟動服務器
	log.Println("Starting genVideoSub API server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}