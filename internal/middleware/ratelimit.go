package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	// limiterTTL IP限流器在无请求后的最大保留时间
	limiterTTL = 10 * time.Minute
	// cleanupInterval 定期清理间隔
	cleanupInterval = 5 * time.Minute
)

// ipLimiter 每个IP的限流器，记录最后访问时间以便清理
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// limiterStore 持有所有 IP 的限流器，并支持定期清理
type limiterStore struct {
	mu       sync.Mutex
	entries  map[string]*ipLimiter
	rps      float64
	burst    int
}

func newLimiterStore(rps float64, burst int) *limiterStore {
	s := &limiterStore{
		entries: make(map[string]*ipLimiter),
		rps:     rps,
		burst:   burst,
	}
	go s.cleanupLoop()
	return s
}

func (s *limiterStore) get(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[ip]
	if !exists {
		entry = &ipLimiter{
			limiter: rate.NewLimiter(rate.Limit(s.rps), s.burst),
		}
		s.entries[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

// cleanupLoop 定期删除长时间无请求的 IP 条目，防止内存无限增长
func (s *limiterStore) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.cleanup()
	}
}

func (s *limiterStore) cleanup() {
	threshold := time.Now().Add(-limiterTTL)
	s.mu.Lock()
	defer s.mu.Unlock()
	for ip, entry := range s.entries {
		if entry.lastSeen.Before(threshold) {
			delete(s.entries, ip)
		}
	}
}

// RateLimiter 创建 IP 级别速率限制中间件。
// 每个中间件实例持有独立的 limiterStore（含后台清理 goroutine）。
func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	store := newLimiterStore(rps, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !store.get(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

// Recovery 异常恢复中间件（防止panic崩溃）
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "服务器内部错误",
		})
	})
}
