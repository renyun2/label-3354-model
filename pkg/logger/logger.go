package logger

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/example/go-api-starter/internal/config"
)

var (
	once   sync.Once
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// Init 初始化日志系统（单例）
func Init(cfg config.LoggerConfig) {
	once.Do(func() {
		Logger = newLogger(cfg)
		Sugar = Logger.Sugar()
	})
}

func newLogger(cfg config.LoggerConfig) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEncoderConfig := encoderConfig
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	level := parseLevel(cfg.Level)
	var cores []zapcore.Core

	// 始终输出到 stdout（JSON 或彩色 console）
	var stdoutEncoder zapcore.Encoder
	if cfg.Format == "json" {
		stdoutEncoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		stdoutEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
	}
	cores = append(cores, zapcore.NewCore(stdoutEncoder, zapcore.AddSync(os.Stdout), level))

	// 文件输出：通过 lumberjack 实现按大小/时间滚动
	if cfg.FilePath != "" {
		roller := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSizeMB,  // 单个文件最大 MB 数，超出后滚动
			MaxBackups: cfg.MaxBackups, // 保留旧文件数量
			MaxAge:     cfg.MaxAgeDays, // 旧文件最长保留天数
			Compress:   cfg.Compress,   // 是否 gzip 压缩旧日志
			LocalTime:  true,           // 备份文件名使用本地时间
		}
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig), // 文件日志始终用 JSON 便于采集
			zapcore.AddSync(roller),
			level,
		)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func Debug(msg string, fields ...zap.Field)  { Logger.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)   { Logger.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)   { Logger.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field)  { Logger.Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field)  { Logger.Fatal(msg, fields...) }
func With(fields ...zap.Field) *zap.Logger   { return Logger.With(fields...) }

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
