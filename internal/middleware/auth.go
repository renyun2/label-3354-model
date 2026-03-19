package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	jwtutil "github.com/example/go-api-starter/pkg/jwt"
	"github.com/example/go-api-starter/pkg/response"
)

// JWTAuth JWT鉴权中间件，仅接受 access token（refresh token 被拒绝）
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Unauthorized(c, "请求头缺少Authorization")
			return
		}

		// ParseAccessToken 会拒绝 role=="refresh" 的 token，防止双token机制被绕过
		claims, err := jwtutil.ParseAccessToken(token, secret)
		if err != nil {
			switch err {
			case jwtutil.ErrTokenExpired:
				response.Unauthorized(c, "Token已过期，请重新登录")
			case jwtutil.ErrEmptySecret:
				response.InternalError(c, "服务端配置错误")
			default:
				response.Unauthorized(c, "Token无效")
			}
			return
		}

		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("open_id", claims.OpenID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole 角色权限中间件（用法：RequireRole("admin")）
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c)
			return
		}

		roleStr, _ := role.(string)
		for _, r := range roles {
			if roleStr == r {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "权限不足")
	}
}

// extractToken 从请求中提取 Bearer Token。
//
// 安全说明：
//   - 标准场景：仅从 Authorization 请求头提取
//   - query 参数 token 会出现在服务器日志、浏览器历史、Referer 头中，存在泄露风险
//   - 例外：WebSocket 握手无法自定义请求头，允许通过 ?token= 传递，其他请求一律拒绝
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
		return ""
	}

	// 仅 WebSocket 升级请求允许通过 query 参数传 token
	if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
		return c.Query("token")
	}

	return ""
}
