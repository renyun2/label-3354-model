// @title           Go API Starter
// @version         1.0
// @description     基于 Gin + GORM + Redis 的 Golang RESTful API 启动套件，内置微信小程序开发支持。
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/example/go-api-starter
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 格式：Bearer {token}

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "github.com/example/go-api-starter/docs" // swagger 生成的文档包
	"github.com/example/go-api-starter/internal/config"
	"github.com/example/go-api-starter/internal/handler"
	"github.com/example/go-api-starter/internal/model"
	"github.com/example/go-api-starter/internal/repository"
	"github.com/example/go-api-starter/internal/router"
	"github.com/example/go-api-starter/internal/service"
	"github.com/example/go-api-starter/pkg/cache"
	"github.com/example/go-api-starter/pkg/database"
	"github.com/example/go-api-starter/pkg/logger"
	pkgwechat "github.com/example/go-api-starter/pkg/wechat"
)

var configPath = flag.String("config", "", "配置文件路径（默认：./configs/config.yaml）")

func main() {
	flag.Parse()

	// 1. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	logger.Init(cfg.Logger)
	defer logger.Sync()

	logger.Info("服务启动中",
		zap.String("name", cfg.App.Name),
		zap.String("version", cfg.App.Version),
		zap.String("mode", cfg.App.Mode),
	)

	// 3. 初始化数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")

	// 4. 自动迁移表结构
	if err = database.AutoMigrate(
		&model.User{},
		&model.UserSession{},
		&model.Article{},
	); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}
	logger.Info("数据库表结构迁移完成")

	// 5. 初始化Redis
	rdb, err := cache.Init(cfg.Redis)
	if err != nil {
		logger.Fatal("初始化Redis失败", zap.Error(err))
	}
	logger.Info("Redis连接成功")

	// 6. 初始化微信客户端
	wechatClient := pkgwechat.NewClient(
		cfg.WeChat.MiniProgram.AppID,
		cfg.WeChat.MiniProgram.AppSecret,
	)

	// 7. 初始化仓库层
	userRepo := repository.NewUserRepository(db)
	articleRepo := repository.NewArticleRepository(db)

	// 8. 初始化服务层
	userSvc := service.NewUserService(userRepo, wechatClient.MiniProgram, cfg)
	articleSvc := service.NewArticleService(articleRepo)
	wechatSvc := service.NewWeChatService(wechatClient.MiniProgram, wechatClient.Message, rdb)

	// 9. 初始化处理器
	h := handler.NewHandler(userSvc, articleSvc, wechatSvc)

	// 10. 设置路由
	engine := router.Setup(cfg, h)

	// 11. 启动HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 异步启动服务
	go func() {
		logger.Info("HTTP服务器启动", zap.Int("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP服务器启动失败", zap.Error(err))
		}
	}()

	// 12. 优雅关闭（监听系统信号）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在优雅关闭服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		logger.Error("服务关闭超时", zap.Error(err))
	}

	logger.Info("服务已安全退出")
}
