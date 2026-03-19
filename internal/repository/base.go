package repository

import (
	"context"

	"gorm.io/gorm"
)

// Pagination 分页参数
type Pagination struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// Normalize 规范化分页参数
func (p *Pagination) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 计算数据库 OFFSET，在未调用 Normalize() 时也保证结果 >= 0
func (p *Pagination) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// BaseRepository 基础仓库（提供通用CRUD操作）
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础仓库
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// DB 获取数据库连接（支持传入事务）
func (r *BaseRepository[T]) DB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// Create 创建记录
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// GetByID 根据ID查询
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Update 更新记录
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// UpdateFields 部分字段更新
func (r *BaseRepository[T]) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	var entity T
	return r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Updates(fields).Error
}

// Delete 软删除
func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}

// Count 统计数量
func (r *BaseRepository[T]) Count(ctx context.Context, conditions ...interface{}) (int64, error) {
	var entity T
	var count int64
	query := r.db.WithContext(ctx).Model(&entity)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	err := query.Count(&count).Error
	return count, err
}

// List 分页查询（通用，子类可覆盖以添加特定条件）
func (r *BaseRepository[T]) List(ctx context.Context, page *Pagination) ([]T, int64, error) {
	page.Normalize()
	var entities []T
	var total int64

	query := r.db.WithContext(ctx).Model(new(T))
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(page.Offset()).Limit(page.PageSize).Find(&entities).Error
	return entities, total, err
}
