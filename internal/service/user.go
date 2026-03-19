package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/example/go-api-starter/internal/config"
	"github.com/example/go-api-starter/internal/model"
	"github.com/example/go-api-starter/internal/repository"
	jwtutil "github.com/example/go-api-starter/pkg/jwt"
	"github.com/example/go-api-starter/pkg/logger"
	"github.com/example/go-api-starter/pkg/wechat"
)

// UserService 用户业务逻辑
type UserService struct {
	userRepo     *repository.UserRepository
	wechatClient *wechat.MiniProgramClient
	cfg          *config.Config
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo *repository.UserRepository,
	wechatClient *wechat.MiniProgramClient,
	cfg *config.Config,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		wechatClient: wechatClient,
		cfg:          cfg,
	}
}

// WeChatLoginRequest 微信登录请求
type WeChatLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User  *model.User        `json:"user"`
	Token *jwtutil.TokenPair `json:"token"`
}

// WeChatLogin 微信小程序登录
func (s *UserService) WeChatLogin(ctx context.Context, req *WeChatLoginRequest) (*LoginResponse, error) {
	session, err := s.wechatClient.Code2Session(ctx, req.Code)
	if err != nil {
		logger.Error("微信code2session失败", zap.String("code", req.Code), zap.Error(err))
		return nil, err
	}

	user := &model.User{
		OpenID:    session.OpenID,
		UnionID:   session.UnionID,
		Nickname:  req.Nickname,
		AvatarURL: req.AvatarURL,
		Role:      "user",
		Status:    1,
	}

	if err = s.userRepo.UpsertByOpenID(ctx, user); err != nil {
		logger.Error("更新或创建用户失败", zap.String("open_id", session.OpenID), zap.Error(err))
		return nil, err
	}

	if err = s.userRepo.UpdateLastLogin(ctx, user.ID, time.Now().Unix()); err != nil {
		logger.Warn("更新最后登录时间失败", zap.Uint("user_id", user.ID), zap.Error(err))
	}

	token, err := jwtutil.GenerateTokenPair(
		user.ID,
		user.OpenID,
		user.Nickname,
		user.Role,
		s.cfg.JWT.Secret,
		s.cfg.JWT.ExpireHours,
		s.cfg.JWT.RefreshExpireHours,
	)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{User: user, Token: token}, nil
}

// GetUserByID 根据ID获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// UpdateProfileRequest 更新用户资料请求。
// Gender 使用 *int8（指针），区分"未传递（nil，不更新）"与"传递了 0（更新为未知）"，
// 避免用户仅更新昵称时 gender 被意外重置为 0。
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname" binding:"max=32"`
	AvatarURL string `json:"avatar_url" binding:"max=512"`
	Gender    *int8  `json:"gender" binding:"omitempty,oneof=0 1 2"`
}

// UpdateProfile 更新用户资料，只更新请求中明确提供的字段
func (s *UserService) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*model.User, error) {
	fields := map[string]interface{}{}
	if req.Nickname != "" {
		fields["nickname"] = req.Nickname
	}
	if req.AvatarURL != "" {
		fields["avatar_url"] = req.AvatarURL
	}
	// 只有显式传递了 gender 字段才更新，nil 表示客户端未传此字段
	if req.Gender != nil {
		fields["gender"] = *req.Gender
	}

	if len(fields) == 0 {
		return s.userRepo.GetByID(ctx, userID)
	}

	if err := s.userRepo.UpdateFields(ctx, userID, fields); err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, userID)
}

// BindPhone 绑定手机号。
//
// 安全说明：废弃先查询再更新的 check-then-act 模式（存在 TOCTOU 竞争条件），
// 改为直接 UPDATE 并依赖数据库 uniqueIndex 做原子性保证，通过错误码 1062 识别冲突。
func (s *UserService) BindPhone(ctx context.Context, userID uint, phone string) error {
	err := s.userRepo.UpdateFields(ctx, userID, map[string]interface{}{
		"phone": phone,
	})
	if err != nil {
		if isMySQLDuplicateKeyError(err) {
			return ErrPhoneAlreadyBound
		}
		return err
	}
	return nil
}

// RefreshToken 刷新访问令牌。
//
// 安全说明：ParseToken 只做签名验证，此处必须显式检查 claims.Role == "refresh"，
// 确保 access token 无法被用来换取新令牌，双 token 机制才有意义。
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*jwtutil.TokenPair, error) {
	claims, err := jwtutil.ParseToken(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return nil, err
	}

	// 严格校验：只接受专用的 refresh token，拒绝 access token 冒充
	if claims.Role != "refresh" {
		return nil, jwtutil.ErrTokenInvalid
	}

	user, err := s.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	return jwtutil.GenerateTokenPair(
		user.ID,
		user.OpenID,
		user.Nickname,
		user.Role,
		s.cfg.JWT.Secret,
		s.cfg.JWT.ExpireHours,
		s.cfg.JWT.RefreshExpireHours,
	)
}

// isMySQLDuplicateKeyError 判断是否为 MySQL 唯一键冲突错误（错误码 1062）
func isMySQLDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrPhoneAlreadyBound = errors.New("该手机号已被绑定")
)
