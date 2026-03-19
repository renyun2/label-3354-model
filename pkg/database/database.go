package database

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	mysqldrv "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/example/go-api-starter/internal/config"
)

var (
	once sync.Once
	db   *gorm.DB
)

// Init 初始化数据库连接
func Init(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var err error
	once.Do(func() {
		db, err = newDB(cfg)
	})
	return db, err
}

// Get 获取数据库实例
func Get() *gorm.DB {
	return db
}

func newDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := buildDSN(cfg)
	gormLogger := newGormLogger(cfg.LogLevel)

	gormCfg := &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	database, err := gorm.Open(mysql.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接池失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetimeHours) * time.Hour)

	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库ping失败: %w", err)
	}

	return database, nil
}

// buildDSN 使用 go-sql-driver 官方 Config 构建 DSN，正确处理密码中的 @、:、/ 等特殊字符。
// 禁止直接拼接 fmt.Sprintf("user:pass@tcp(...)") —— 密码中的特殊字符会导致解析错误。
func buildDSN(cfg config.DatabaseConfig) string {
	drvCfg := mysqldrv.Config{
		User:                 cfg.Username,
		Passwd:               cfg.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DBName:               cfg.DBName,
		Params:               map[string]string{"charset": cfg.Charset},
		ParseTime:            true,
		Loc:                  time.Local,
		AllowNativePasswords: true,
	}
	return drvCfg.FormatDSN()
}

func newGormLogger(level string) logger.Interface {
	logLevel := map[string]logger.LogLevel{
		"silent": logger.Silent,
		"error":  logger.Error,
		"warn":   logger.Warn,
		"info":   logger.Info,
	}

	lvl, ok := logLevel[level]
	if !ok {
		lvl = logger.Info
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  lvl,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(models ...interface{}) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	return db.AutoMigrate(models...)
}

// Transaction 事务封装
func Transaction(fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}
