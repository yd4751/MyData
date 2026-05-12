package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"resource-station/internal/api"
	"resource-station/internal/database"
	"resource-station/internal/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := utils.LoadConfig("configs/config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	config := utils.GetConfig()

	// 设置Gin模式
	gin.SetMode(config.Server.Mode)

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.CloseDB()

	// 自动迁移数据库表
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建Gin引擎
	router := gin.Default()

	// 设置中间件
	setupMiddleware(router)

	// 设置路由
	setupRoutes(router)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         config.Server.Host + ":" + strconv.Itoa(config.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动在 %s:%d", config.Server.Host, config.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("服务器已关闭")
}

func setupMiddleware(router *gin.Engine) {
	// 恢复中间件
	router.Use(gin.Recovery())

	// 日志中间件
	router.Use(gin.Logger())

	// CORS中间件
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

func setupRoutes(router *gin.Engine) {
	// 提供静态文件（前端资源）
	router.Static("/static", "../frontend")
	
	// 创建API处理器
	userHandler := api.NewUserHandler()
	resourceHandler, err := api.NewResourceHandler()
	if err != nil {
		log.Fatalf("创建资源处理器失败: %v", err)
	}

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 公开路由
		public := v1.Group("/auth")
		{
			public.POST("/register", userHandler.Register)
			public.POST("/login", userHandler.Login)
		}

		// 需要认证的路由
		auth := v1.Group("/")
		auth.Use(api.AuthMiddleware())
		{
			// 用户相关
			user := auth.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PUT("/profile", userHandler.UpdateProfile)
			}

			// 资源相关
			resources := auth.Group("/resources")
			{
				resources.POST("", resourceHandler.CreateResource)
				resources.GET("", resourceHandler.GetResources)
				resources.POST("/batch", resourceHandler.BatchUpload)
				
				// 单个资源操作
				resource := resources.Group("/:id")
				{
					resource.GET("", resourceHandler.GetResource)
					resource.PUT("", resourceHandler.UpdateResource)
					resource.DELETE("", resourceHandler.DeleteResource)
					resource.GET("/progress", resourceHandler.GetUploadProgress)
					resource.POST("/complete", resourceHandler.CompleteUpload)
					
					// 分片上传
					resource.POST("/chunks/:chunkIndex", resourceHandler.UploadChunk)
				}
			}

			// 下载资源（公开资源不需要认证）
			v1.GET("/resources/:id/download", resourceHandler.DownloadResource)
		}

		// 管理员路由
		admin := auth.Group("/admin")
		admin.Use(api.AdminMiddleware())
		{
			// 这里可以添加管理员专用路由
			admin.GET("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "管理员功能"})
			})
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 根路径 - 提供前端首页
	router.GET("/", func(c *gin.Context) {
		c.File("../frontend/index.html")
	})
	
	// 处理前端路由 - 所有未匹配的路由都返回前端首页（支持前端路由）
	router.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API endpoint not found",
			})
			return
		}
		// 否则返回前端首页
		c.File("../frontend/index.html")
	})
}
