package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	jwtutil "github.com/example/go-api-starter/pkg/jwt"
	"github.com/example/go-api-starter/internal/middleware"
)

const testJWTSecret = "test-jwt-secret-that-is-long-enough-32chars"

func newAuthEngine(secret string) *gin.Engine {
	r := gin.New()
	r.Use(middleware.JWTAuth(secret))
	r.GET("/protected", func(c *gin.Context) {
		claims, _ := c.Get("claims")
		c.JSON(http.StatusOK, gin.H{"user_id": claims.(*jwtutil.Claims).UserID})
	})
	return r
}

func makeAccessToken(t *testing.T, userID uint, role string) string {
	t.Helper()
	pair, err := jwtutil.GenerateTokenPair(userID, "openid", "name", role, testJWTSecret, 72, 168)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}
	return pair.AccessToken
}

func makeRefreshToken(t *testing.T, userID uint) string {
	t.Helper()
	pair, err := jwtutil.GenerateTokenPair(userID, "openid", "name", "user", testJWTSecret, 72, 168)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}
	return pair.RefreshToken
}

// TestJWTAuth_ValidToken 有效 token 应通过鉴权
func TestJWTAuth_ValidToken(t *testing.T) {
	r := newAuthEngine(testJWTSecret)
	token := makeAccessToken(t, 1, "user")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("有效 token 应返回 200，实际: %d，body: %s", w.Code, w.Body.String())
	}
}

// TestJWTAuth_NoToken 无 token 应返回 401
func TestJWTAuth_NoToken(t *testing.T) {
	r := newAuthEngine(testJWTSecret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("无 token 应返回 401，实际: %d", w.Code)
	}
}

// TestJWTAuth_WrongSecret 错误 secret 签发的 token 应返回 401
func TestJWTAuth_WrongSecret(t *testing.T) {
	r := newAuthEngine(testJWTSecret)

	// 用另一个 secret 生成 token
	pair, _ := jwtutil.GenerateTokenPair(1, "", "", "user", "other-secret-key-that-is-long-enough", 72, 168)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("错误 secret 应返回 401，实际: %d", w.Code)
	}
}

// TestJWTAuth_RefreshTokenRejected refresh token 不能用于鉴权
func TestJWTAuth_RefreshTokenRejected(t *testing.T) {
	r := newAuthEngine(testJWTSecret)
	refreshToken := makeRefreshToken(t, 1)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("refresh token 不应通过鉴权，应返回 401，实际: %d，body: %s", w.Code, w.Body.String())
	}
}

// TestJWTAuth_MalformedHeader Authorization 头格式错误应返回 401
func TestJWTAuth_MalformedHeader(t *testing.T) {
	r := newAuthEngine(testJWTSecret)
	token := makeAccessToken(t, 1, "user")

	cases := []struct {
		name  string
		value string
	}{
		{"无Bearer前缀", token},
		{"前缀错误", "Token " + token},
		{"只有Bearer", "Bearer"},
		{"空字符串", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			if tc.value != "" {
				req.Header.Set("Authorization", tc.value)
			}
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s: 期望 401，实际 %d", tc.name, w.Code)
			}
		})
	}
}

// TestJWTAuth_QueryTokenRejected 普通请求不能通过 query 参数传 token
func TestJWTAuth_QueryTokenRejected(t *testing.T) {
	r := newAuthEngine(testJWTSecret)
	token := makeAccessToken(t, 1, "user")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/protected?token=%s", token), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("普通请求不应支持 query token，期望 401，实际 %d", w.Code)
	}
}

// TestJWTAuth_WebSocketQueryToken WebSocket 升级请求允许 query 参数 token
func TestJWTAuth_WebSocketQueryToken(t *testing.T) {
	r := newAuthEngine(testJWTSecret)
	token := makeAccessToken(t, 1, "user")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/protected?token=%s", token), nil)
	// 模拟 WebSocket 升级请求
	req.Header.Set("Upgrade", "websocket")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("WebSocket 升级请求应支持 query token，期望 200，实际 %d，body: %s", w.Code, w.Body.String())
	}
}

// TestRequireRole_Pass 用户具有所需角色时通过
func TestRequireRole_Pass(t *testing.T) {
	r := gin.New()
	r.Use(middleware.JWTAuth(testJWTSecret))
	r.GET("/admin", middleware.RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	token := makeAccessToken(t, 1, "admin")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("admin 角色应通过，期望 200，实际 %d", w.Code)
	}
}

// TestRequireRole_Forbidden 用户角色不足时应返回 403
func TestRequireRole_Forbidden(t *testing.T) {
	r := gin.New()
	r.Use(middleware.JWTAuth(testJWTSecret))
	r.GET("/admin", middleware.RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	token := makeAccessToken(t, 1, "user") // 普通用户

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// Forbidden 使用 c.AbortWithStatusJSON(403)，HTTP 状态码为 403
	if w.Code != http.StatusForbidden {
		t.Errorf("user 角色访问 admin 路由应返回 HTTP 403，实际 %d，body: %s", w.Code, w.Body.String())
	}
}
