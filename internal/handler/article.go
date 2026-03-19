package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/example/go-api-starter/internal/service"
	"github.com/example/go-api-starter/pkg/response"
)

// ArticleHandler 文章处理器
type ArticleHandler struct {
	articleSvc *service.ArticleService
}

// NewArticleHandler 创建文章处理器
func NewArticleHandler(articleSvc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{articleSvc: articleSvc}
}

// ListArticles godoc
// @Summary     获取文章列表
// @Tags        文章
// @Produce     json
// @Param       page      query int false "页码" default(1)
// @Param       page_size query int false "每页数量" default(10)
// @Success     200 {object} response.Response{data=response.PageData}
// @Router      /api/v1/articles [get]
func (h *ArticleHandler) ListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	articles, total, err := h.articleSvc.ListArticles(c.Request.Context(), parsePagination(page, pageSize))
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Page(c, articles, total, page, pageSize)
}

// GetArticle godoc
// @Summary     获取文章详情
// @Tags        文章
// @Produce     json
// @Param       id path int true "文章ID"
// @Success     200 {object} response.Response{data=model.Article}
// @Router      /api/v1/articles/{id} [get]
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的文章ID")
		return
	}

	article, err := h.articleSvc.GetArticle(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrArticleNotFound) {
			response.NotFound(c, "文章不存在")
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, article)
}

// CreateArticle godoc
// @Summary     创建文章
// @Tags        文章
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body service.CreateArticleRequest true "文章内容"
// @Success     200 {object} response.Response{data=model.Article}
// @Router      /api/v1/articles [post]
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		response.Unauthorized(c)
		return
	}

	var req service.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	article, err := h.articleSvc.CreateArticle(c.Request.Context(), claims.UserID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, article)
}

// UpdateArticle godoc
// @Summary     更新文章
// @Tags        文章
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path int true "文章ID"
// @Param       body body service.UpdateArticleRequest true "文章内容"
// @Success     200 {object} response.Response{data=model.Article}
// @Router      /api/v1/articles/{id} [put]
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		response.Unauthorized(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的文章ID")
		return
	}

	var req service.UpdateArticleRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	article, err := h.articleSvc.UpdateArticle(c.Request.Context(), uint(id), claims.UserID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrArticleNotFound):
			response.NotFound(c, "文章不存在")
		case errors.Is(err, service.ErrForbidden):
			response.Forbidden(c, "无权修改他人文章")
		default:
			response.InternalError(c)
		}
		return
	}

	response.Success(c, article)
}

// DeleteArticle godoc
// @Summary     删除文章
// @Tags        文章
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "文章ID"
// @Success     200 {object} response.Response
// @Router      /api/v1/articles/{id} [delete]
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		response.Unauthorized(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的文章ID")
		return
	}

	if err = h.articleSvc.DeleteArticle(c.Request.Context(), uint(id), claims.UserID); err != nil {
		switch {
		case errors.Is(err, service.ErrArticleNotFound):
			response.NotFound(c, "文章不存在")
		case errors.Is(err, service.ErrForbidden):
			response.Forbidden(c, "无权删除他人文章")
		default:
			response.InternalError(c)
		}
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
