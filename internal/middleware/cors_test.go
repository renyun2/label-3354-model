package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/example/go-api-starter/internal/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newCORSEngine(allowedOrigins []string) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS(allowedOrigins))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.OPTIONS("/test", func(c *gin.Context) {})
	return r
}

// TestCORS_AllowedOrigin 白名单内的 Origin 应获得正确的 CORS 头
func TestCORS_AllowedOrigin(t *testing.T) {
	r := newCORSEngine([]string{"https://example.com"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("Allow-Origin: 期望 'https://example.com'，实际 '%s'", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Allow-Credentials: 期望 'true'，实际 '%s'", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary: 期望 'Origin'，实际 '%s'", got)
	}
}

// TestCORS_ForbiddenOrigin 不在白名单的 Origin 不应获得 CORS 头
func TestCORS_ForbiddenOrigin(t *testing.T) {
	r := newCORSEngine([]string{"https://example.com"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("非白名单 Origin 不应设置 Allow-Origin，实际: '%s'", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("非白名单 Origin 不应设置 Allow-Credentials，实际: '%s'", got)
	}
}

// TestCORS_ForbiddenOrigin_Options 非白名单 OPTIONS 请求应返回 403
func TestCORS_ForbiddenOrigin_Options(t *testing.T) {
	r := newCORSEngine([]string{"https://example.com"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("非白名单 OPTIONS 应返回 403，实际: %d", w.Code)
	}
}

// TestCORS_Preflight_ReturnsNoContent 预检请求应返回 204
func TestCORS_Preflight_ReturnsNoContent(t *testing.T) {
	r := newCORSEngine([]string{"https://example.com"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("预检请求应返回 204，实际: %d", w.Code)
	}
}

// TestCORS_EmptyAllowedOrigins_WildcardNoCredentials 空白名单（开发模式）允许所有来源但不携带凭证
func TestCORS_EmptyAllowedOrigins_WildcardNoCredentials(t *testing.T) {
	r := newCORSEngine([]string{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("空白名单应返回 '*'，实际: '%s'", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got == "true" {
		t.Error("空白名单模式不应设置 Allow-Credentials: true（安全风险）")
	}
}

// TestCORS_NoOriginHeader 无 Origin 头的请求不应添加 CORS 头（非跨域请求）
func TestCORS_NoOriginHeader(t *testing.T) {
	r := newCORSEngine([]string{"https://example.com"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("无 Origin 头时不应设置 CORS 头，实际: '%s'", got)
	}
}

// TestCORS_MultipleAllowedOrigins 多个白名单域名各自通过
func TestCORS_MultipleAllowedOrigins(t *testing.T) {
	origins := []string{"https://a.com", "https://b.com", "https://c.com"}
	r := newCORSEngine(origins)

	for _, origin := range origins {
		t.Run(origin, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", origin)
			r.ServeHTTP(w, req)

			if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Errorf("Origin %s: 期望 Allow-Origin=%s，实际 %s", origin, origin, got)
			}
		})
	}
}
