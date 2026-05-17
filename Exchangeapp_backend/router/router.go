package router

import (
	"exchangeapp/controllers"
	"exchangeapp/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 注册全部 HTTP 路由与中间件。
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 允许本地前端开发环境跨域访问。
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 鉴权相关公开接口。
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/register", controllers.Register)
	}

	// 智能助手公开接口（内部自行做降级与容错）。
	rag := r.Group("/api/assistant")
	{
		rag.GET("/status", controllers.AssistantStatus)
		rag.POST("/chat", controllers.AssistantChat)
	}

	// 业务接口分组：先开放读取汇率，再挂载鉴权中间件保护写接口。
	api := r.Group("/api")
	api.GET("/exchangeRates", controllers.GetExchangeRates)
	api.Use(middlewares.AuthMiddleware())
	{
		api.POST("/exchangeRates", controllers.CreateExchangeRate)
		api.POST("/articles", controllers.CreateArticle)
		api.GET("/my-articles", controllers.GetMyArticles)
		api.GET("/articles", controllers.GetArticles)
		api.GET("/articles/:id", controllers.GetArticlesByID)
		api.PUT("/articles/:id", controllers.UpdateArticle)
		api.DELETE("/articles/:id", controllers.DeleteArticle)

		api.POST("/articles/:id/like", controllers.LikeArticle)
		api.GET("/articles/:id/like", controllers.GetArticleLikes)
	}
	return r
}
