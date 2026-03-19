package service

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/example/go-api-starter/internal/model"
	"github.com/example/go-api-starter/internal/repository"
	"github.com/example/go-api-starter/pkg/logger"
)

// ArticleService 文章业务逻辑
type ArticleService struct {
	articleRepo *repository.ArticleRepository
}

// NewArticleService 创建文章服务
func NewArticleService(articleRepo *repository.ArticleRepository) *ArticleService {
	return &ArticleService{articleRepo: articleRepo}
}

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title   string `json:"title" binding:"required,max=256"`
	Content string `json:"content" binding:"required"`
	Status  int8   `json:"status" binding:"oneof=0 1"`
}

// CreateArticle 创建文章
func (s *ArticleService) CreateArticle(ctx context.Context, authorID uint, req *CreateArticleRequest) (*model.Article, error) {
	article := &model.Article{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID,
		Status:   req.Status,
	}

	if err := s.articleRepo.Create(ctx, article); err != nil {
		return nil, err
	}
	return article, nil
}

// GetArticle 获取文章详情
func (s *ArticleService) GetArticle(ctx context.Context, id uint) (*model.Article, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}

	// 异步增加浏览次数。
	// 使用独立 context（不依赖请求 ctx，避免请求结束后 ctx 已取消导致操作失败），
	// 但保留 traceID 以便链路追踪，并记录错误而非静默忽略。
	traceID, _ := ctx.Value("trace_id").(string)
	go func() {
		bCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if traceID != "" {
			bCtx = context.WithValue(bCtx, "trace_id", traceID) //nolint:staticcheck
		}
		if err := s.articleRepo.IncrViewCount(bCtx, id); err != nil {
			logger.Warn("增加浏览次数失败",
				zap.Uint("article_id", id),
				zap.String("trace_id", traceID),
				zap.Error(err),
			)
		}
	}()

	return article, nil
}

// ListArticles 分页获取已发布文章列表
func (s *ArticleService) ListArticles(ctx context.Context, page *repository.Pagination) ([]model.Article, int64, error) {
	return s.articleRepo.ListPublished(ctx, page)
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title   string `json:"title" binding:"max=256"`
	Content string `json:"content"`
	Status  *int8  `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateArticle 更新文章
func (s *ArticleService) UpdateArticle(ctx context.Context, id, authorID uint, req *UpdateArticleRequest) (*model.Article, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}

	if article.AuthorID != authorID {
		return nil, ErrForbidden
	}

	fields := map[string]interface{}{}
	if req.Title != "" {
		fields["title"] = req.Title
	}
	if req.Content != "" {
		fields["content"] = req.Content
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}

	if err = s.articleRepo.UpdateFields(ctx, id, fields); err != nil {
		return nil, err
	}

	return s.articleRepo.GetByID(ctx, id)
}

// DeleteArticle 删除文章（软删除）
func (s *ArticleService) DeleteArticle(ctx context.Context, id, authorID uint) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrArticleNotFound
		}
		return err
	}

	if article.AuthorID != authorID {
		return ErrForbidden
	}

	return s.articleRepo.Delete(ctx, id)
}

var (
	ErrArticleNotFound = errors.New("文章不存在")
	ErrForbidden       = errors.New("无权限操作")
)
