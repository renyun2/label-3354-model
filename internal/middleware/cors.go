package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件，基于来源白名单。
//
// 安全规则：
//   - 白名单非空时：仅对白名单内的 Origin 设置 Allow-Credentials: true，并通过 Vary: Origin 防止缓存投毒
//   - 白名单为空时（开发环境）：允许所有来源，但明确不携带凭证（无 Allow-Credentials 头）
//   - 不在白名单中的 Origin：不返回任何 CORS 头，浏览器将阻止该跨域请求
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Next()
			return
		}

		if len(allowedOrigins) == 0 {
			// 开发模式回退：允许任意来源，但不允许携带凭证（Cookie/Authorization）
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			_, allowed := originSet[origin]
			if !allowed {
				// 非白名单来源：拒绝 CORS，OPTIONS 直接返回 403
				if c.Request.Method == http.MethodOptions {
					c.AbortWithStatus(http.StatusForbidden)
				} else {
					c.Next()
				}
				return
			}
			// 白名单来源：精确反射 Origin，允许携带凭证
			// Vary: Origin 告知缓存层此响应因 Origin 不同而不同，防止缓存投毒
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Request-ID, X-Trace-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID, X-Trace-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
