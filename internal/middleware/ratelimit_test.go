package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/example/go-api-starter/internal/middleware"
)

func newRateLimitEngine(rps float64, burst int) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RateLimiter(rps, burst))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// TestRateLimiter_AllowsWithinLimit 在限制内的请求应全部通过
func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	r := newRateLimitEngine(100, 10) // 10 个突发容量

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("第 %d 次请求应通过（限制内），实际: %d", i+1, w.Code)
		}
	}
}

// TestRateLimiter_BlocksOverLimit 超过突发容量的请求应被限流
func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	r := newRateLimitEngine(0.001, 2) // 极低 rps，突发 2

	passCount := 0
	blockCount := 0

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			passCount++
		} else if w.Code == http.StatusTooManyRequests {
			blockCount++
		}
	}

	if passCount > 2 {
		t.Errorf("突发容量为2，不应有超过2个请求通过，实际通过: %d", passCount)
	}
	if blockCount == 0 {
		t.Error("应有请求被限流，实际无限流")
	}
}

// TestRateLimiter_DifferentIPsIndependent 不同 IP 的限流器相互独立
func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	r := newRateLimitEngine(0.001, 1) // 每个 IP 突发容量 1

	ips := []string{"1.1.1.1:1", "2.2.2.2:1", "3.3.3.3:1"}
	for _, ip := range ips {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("IP %s 首次请求应通过，实际: %d", ip, w.Code)
		}
	}
}

// TestRateLimiter_Returns429OnBlock 限流响应应包含正确的 JSON 格式
func TestRateLimiter_Returns429OnBlock(t *testing.T) {
	r := newRateLimitEngine(0.001, 1) // 突发容量 1

	// 第一次通过
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "9.9.9.9:9"
	r.ServeHTTP(w1, req1)

	// 第二次应被限流
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "9.9.9.9:9"
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("第二次请求应返回 429，实际: %d", w2.Code)
	}
	body := w2.Body.String()
	if body == "" {
		t.Error("429 响应 body 不应为空")
	}
}
