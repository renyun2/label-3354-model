package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/example/go-api-starter/internal/model"
)

// UserRepository 用户仓库
type UserRepository struct {
	*BaseRepository[model.User]
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository[model.User](db),
	}
}

// GetByOpenID 通过微信OpenID查询用户
func (r *UserRepository) GetByOpenID(ctx context.Context, openID string) (*model.User, error) {
	var user model.User
	err := r.DB(ctx).Where("open_id = ?", openID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone 通过手机号查询用户（phone 为指针，NULL 值不参与唯一索引）
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.DB(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers 分页查询用户列表（管理员使用）
func (r *UserRepository) ListUsers(ctx context.Context, page *Pagination, status int8) ([]model.User, int64, error) {
	page.Normalize()
	var users []model.User
	var total int64

	query := r.DB(ctx).Model(&model.User{})
	if status != 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(page.Offset()).Limit(page.PageSize).
		Order("created_at DESC").Find(&users).Error
	return users, total, err
}

// UpsertByOpenID 通过OpenID更新或创建用户
func (r *UserRepository) UpsertByOpenID(ctx context.Context, user *model.User) error {
	return r.DB(ctx).
		Where(model.User{OpenID: user.OpenID}).
		Assign(model.User{
			Nickname:  user.Nickname,
			AvatarURL: user.AvatarURL,
			UnionID:   user.UnionID,
		}).
		FirstOrCreate(user).Error
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uint, timestamp int64) error {
	return r.DB(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("last_login", timestamp).Error
}
