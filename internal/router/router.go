package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/example/go-api-starter/internal/config"
	"github.com/example/go-api-starter/internal/handler"
	"github.com/example/go-api-starter/internal/middleware"
)

// Setup 初始化并返回Gin路由
func Setup(cfg *config.Config, h *handler.Handler) *gin.Engine {
	gin.SetMode(cfg.App.Mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(cfg.App.CORSAllowedOrigins))
	r.Use(middleware.RequestLogger())

	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst))
	}

	// 健康检查（不需要鉴权）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": cfg.App.Version,
			"service": cfg.App.Name,
		})
	})

	// Swagger UI（仅非生产环境暴露，避免接口文档泄露）
	if cfg.App.Mode != "release" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL("/swagger/doc.json"),
			ginSwagger.DefaultModelsExpandDepth(-1),
		))
	}

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证相关（无需Token）
		auth := v1.Group("/auth")
		{
			auth.POST("/wechat-login", h.User.WeChatLogin)
			auth.POST("/refresh", h.User.RefreshToken)
		}

		// 需要JWT鉴权的路由
		authorized := v1.Group("")
		authorized.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// 用户信息
			user := authorized.Group("/user")
			{
				user.GET("/profile", h.User.GetProfile)
				user.PUT("/profile", h.User.UpdateProfile)
				user.POST("/bind-phone", h.User.BindPhone)
			}

			// 文章（登录用户可操作）
			articleAuth := authorized.Group("/articles")
			{
				articleAuth.POST("", h.Article.CreateArticle)
				articleAuth.PUT("/:id", h.Article.UpdateArticle)
				articleAuth.DELETE("/:id", h.Article.DeleteArticle)
			}

			// 微信数据交互（登录用户可用）
			wechatUser := authorized.Group("/wechat")
			{
				// 获取手机号：小程序端 wx.getPhoneNumber() 返回 code 后调用此接口
				wechatUser.POST("/phone", h.WeChat.GetPhoneNumber)
			}

			// 微信消息推送（管理员权限）
			wechatAdmin := authorized.Group("/wechat")
			wechatAdmin.Use(middleware.RequireRole("admin"))
			{
				wechatAdmin.POST("/message/subscribe", h.WeChat.SendSubscribeMessage)
				wechatAdmin.POST("/message/customer", h.WeChat.SendCustomerMessage)
			}
		}

		// 公开路由（无需Token）
		v1.GET("/articles", h.Article.ListArticles)
		v1.GET("/articles/:id", h.Article.GetArticle)
	}

	return r
}
