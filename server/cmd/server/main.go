package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/dz-poker/server/internal/config"
	"github.com/dz-poker/server/internal/handler"
	"github.com/dz-poker/server/internal/repository"
	"github.com/dz-poker/server/internal/service"
)

func main() {
	cfg := config.Load()

	// 初始化 SQLite 数据库
	if err := repository.InitDB(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repository.CloseDB()

	// 创建服务
	authService := service.NewAuthService()
	authHandler := handler.NewAuthHandler(authService)

	// 创建 Hub 和存储
	roomStore := repository.NewRoomStore()
	hub := handler.NewHub(roomStore, authService)

	// 让 authHandler 能访问 Hub 用于推送消息
	authHandler.SetHub(hub)

	// 启动 Hub
	go hub.Run()

	// 设置 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// CORS 配置
	allowOrigins := []string{
		"http://localhost:1420",
		"http://127.0.0.1:1420",
		"http://localhost:1421",
		"http://127.0.0.1:1421",
	}
	// 生产环境添加部署地址
	if os.Getenv("GIN_MODE") == "release" {
		allowOrigins = append(allowOrigins, "http://8.145.38.16:8088")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:          12 * time.Hour,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 认证路由
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// 需要认证的路由
	userGroup := r.Group("/api/user")
	userGroup.Use(handler.AuthMiddleware(authService))
	{
		userGroup.GET("/info", authHandler.GetUserInfo)
		userGroup.GET("/history", authHandler.GetGameHistory)
	}

	// 管理员路由
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(handler.AuthMiddleware(authService))
	{
		adminGroup.GET("/users", authHandler.GetAllUsers)
		adminGroup.POST("/chips", authHandler.UpdateUserChips)
	}

	// WebSocket 端点
	r.GET("/ws", hub.HandleWebSocket)

	log.Printf("Server starting on WS port %s, HTTP port %s", cfg.WSPort, cfg.HTTPPort)

	// 启动 HTTP 服务（包含 WebSocket）
	go func() {
		if err := r.Run(":" + cfg.WSPort); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
