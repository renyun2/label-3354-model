package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/example/go-api-starter/internal/model"
)

// ArticleRepository 文章仓库
type ArticleRepository struct {
	*BaseRepository[model.Article]
}

// NewArticleRepository 创建文章仓库
func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{
		BaseRepository: NewBaseRepository[model.Article](db),
	}
}

// ListByAuthor 分页查询某用户的文章
func (r *ArticleRepository) ListByAuthor(ctx context.Context, authorID uint, page *Pagination) ([]model.Article, int64, error) {
	page.Normalize()
	var articles []model.Article
	var total int64

	query := r.DB(ctx).Model(&model.Article{}).Where("author_id = ?", authorID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Author").
		Order("created_at DESC").
		Offset(page.Offset()).Limit(page.PageSize).
		Find(&articles).Error
	return articles, total, err
}

// ListPublished 查询已发布文章列表
func (r *ArticleRepository) ListPublished(ctx context.Context, page *Pagination) ([]model.Article, int64, error) {
	page.Normalize()
	var articles []model.Article
	var total int64

	query := r.DB(ctx).Model(&model.Article{}).Where("status = ?", 1)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Author").
		Order("created_at DESC").
		Offset(page.Offset()).Limit(page.PageSize).
		Find(&articles).Error
	return articles, total, err
}

// IncrViewCount 原子性增加浏览次数
func (r *ArticleRepository) IncrViewCount(ctx context.Context, id uint) error {
	return r.DB(ctx).Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}
