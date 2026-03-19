package handler

import (
	"github.com/example/go-api-starter/internal/repository"
	"github.com/example/go-api-starter/internal/service"
	"github.com/example/go-api-starter/pkg/jwt"
)

// Handler 聚合所有处理器
type Handler struct {
	User    *UserHandler
	Article *ArticleHandler
	WeChat  *WeChatHandler
}

// NewHandler 创建处理器聚合
func NewHandler(
	userSvc *service.UserService,
	articleSvc *service.ArticleService,
	wechatSvc *service.WeChatService,
) *Handler {
	return &Handler{
		User:    NewUserHandler(userSvc),
		Article: NewArticleHandler(articleSvc),
		WeChat:  NewWeChatHandler(wechatSvc),
	}
}

// parsePagination 从请求中解析分页参数
func parsePagination(page, pageSize int) *repository.Pagination {
	p := &repository.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
	p.Normalize()
	return p
}

// getUserIDFromContext 从Gin上下文中获取当前用户ID
func getUserIDFromContext(c interface{ GetUint(string) (uint, bool) }) (uint, bool) {
	return c.GetUint("user_id")
}

// claimsKey 在上下文中存储JWT claims的key
const claimsKey = "claims"

// getClaims 从Gin上下文获取JWT声明
func getClaims(c interface {
	Get(string) (interface{}, bool)
}) (*jwt.Claims, bool) {
	v, exists := c.Get(claimsKey)
	if !exists {
		return nil, false
	}
	claims, ok := v.(*jwt.Claims)
	return claims, ok
}
