package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dz-poker/server/internal/config"
	"github.com/dz-poker/server/internal/handler"
	"github.com/dz-poker/server/internal/repository"
)

func main() {
	cfg := config.Load()

	// 创建 Hub 和存储
	roomStore := repository.NewRoomStore()
	hub := handler.NewHub(roomStore)

	// 启动 Hub
	go hub.Run()

	// 设置 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// WebSocket 端点
	r.GET("/ws", hub.HandleWebSocket)

	log.Printf("Server starting on WS port %s, HTTP port %s", cfg.WSPort, cfg.HTTPPort)

	// 启动 HTTP 服务（包含 WebSocket）
	go func() {
		if err := r.Run(":" + cfg.WSPort); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 阻塞
	select {}
}
