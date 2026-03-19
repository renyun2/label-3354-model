package middleware

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/example/go-api-starter/pkg/logger"
)

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 生成TraceID并注入上下文
		traceID := generateTraceID()
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}

		fields := []zap.Field{
			zap.String("trace_id", traceID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// 记录错误信息
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		switch {
		case statusCode >= 500:
			logger.Error("请求处理失败", fields...)
		case statusCode >= 400:
			logger.Warn("请求参数错误", fields...)
		default:
			logger.Info("请求完成", fields...)
		}
	}
}

func generateTraceID() string {
	return fmt.Sprintf("%d-%08x", time.Now().UnixNano(), rand.Uint32())
}
