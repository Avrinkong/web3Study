package main

import (
	"log"
	"os"

	"taskFour/config"
	"taskFour/controllers"
	"taskFour/middleware"
	"taskFour/models"

	_ "taskFour/docs" // 重要：导入自动生成的docs包

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// 添加全局db变量声明
var db *gorm.DB

// @title 个人博客系统 API
// @version 1.0
// @description 基于Go、Gin和GORM构建的个人博客系统后端API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT认证令牌，格式: "Bearer {token}"

func main() {
	// 初始化数据库
	var err error
	err = config.ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db = config.GetDB()

	// 自动迁移数据库表
	err = db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 设置日志
	setupLogger()

	// 初始化Gin路由
	router := setupRouter()

	// 启动服务器
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// setupRouter 配置路由
func setupRouter() *gin.Engine {
	router := gin.Default()

	// 全局中间件
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.ErrorHandler())

	// Swagger路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	router.GET("/health", healthCheck)

	// API路由分组 - 添加这部分缺失的路由配置
	api := router.Group("/api")
	{
		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		// 文章路由
		posts := api.Group("/posts")
		{
			posts.GET("", controllers.GetPosts)
			posts.GET("/:id", controllers.GetPost)
			posts.GET("/:id/comments", controllers.GetPostComments)

			// 需要认证的路由
			authPosts := posts.Group("")
			authPosts.Use(middleware.AuthMiddleware())
			{
				authPosts.POST("", controllers.CreatePost)
				authPosts.PUT("/:id", controllers.UpdatePost)
				authPosts.DELETE("/:id", controllers.DeletePost)
			}
		}

		// 评论路由
		comments := api.Group("/comments")
		comments.Use(middleware.AuthMiddleware())
		{
			comments.POST("", controllers.CreateComment)
		}
	}

	return router
}

// healthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "服务状态"
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "服务运行正常",
	})
}

// setupLogger 设置日志
func setupLogger() {
	// 创建日志文件
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	// 设置日志输出
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
