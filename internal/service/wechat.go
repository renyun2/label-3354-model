package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/example/go-api-starter/pkg/cache"
	"github.com/example/go-api-starter/pkg/logger"
	pkgwechat "github.com/example/go-api-starter/pkg/wechat"
)

const (
	accessTokenCacheKey    = "wechat:access_token"
	accessTokenCacheBuffer = 5 * time.Minute // 提前5分钟刷新
)

// WeChatService 微信业务服务
type WeChatService struct {
	miniClient    *pkgwechat.MiniProgramClient
	messageClient *pkgwechat.MessageClient
	rdb           *redis.Client
}

// NewWeChatService 创建微信服务
func NewWeChatService(
	miniClient *pkgwechat.MiniProgramClient,
	messageClient *pkgwechat.MessageClient,
	rdb *redis.Client,
) *WeChatService {
	return &WeChatService{
		miniClient:    miniClient,
		messageClient: messageClient,
		rdb:           rdb,
	}
}

// GetAccessToken 获取微信access_token（带Redis缓存）
func (s *WeChatService) GetAccessToken(ctx context.Context) (string, error) {
	// 先从缓存取
	token, err := cache.GetString(ctx, accessTokenCacheKey)
	if err == nil && token != "" {
		return token, nil
	}

	// 缓存未命中，从微信API获取
	result, err := s.miniClient.GetAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("获取access_token失败: %w", err)
	}

	// 存入缓存，TTL减去缓冲时间
	ttl := time.Duration(result.ExpiresIn)*time.Second - accessTokenCacheBuffer
	if ttl > 0 {
		if err = cache.SetString(ctx, accessTokenCacheKey, result.AccessToken, ttl); err != nil {
			logger.Warn("缓存access_token失败", zap.Error(err))
		}
	}

	return result.AccessToken, nil
}

// SendSubscribeMessage 发送订阅消息
func (s *WeChatService) SendSubscribeMessage(ctx context.Context, req *pkgwechat.SubscribeMessageRequest) error {
	accessToken, err := s.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	if err = s.messageClient.SendSubscribeMessage(ctx, accessToken, req); err != nil {
		logger.Error("发送订阅消息失败",
			zap.String("to_user", req.ToUser),
			zap.String("template_id", req.TemplateID),
			zap.Error(err),
		)
		return err
	}

	logger.Info("订阅消息发送成功",
		zap.String("to_user", req.ToUser),
		zap.String("template_id", req.TemplateID),
	)
	return nil
}

// SendCustomerServiceMessage 发送客服文字消息
func (s *WeChatService) SendCustomerServiceMessage(ctx context.Context, openID, content string) error {
	accessToken, err := s.GetAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.messageClient.SendCustomerServiceTextMessage(ctx, accessToken, openID, content)
}

// GetPhoneNumber 通过小程序 getPhoneNumber 返回的 code 获取用户手机号
//
// 小程序端流程：
//  1. 调用 wx.getPhoneNumber() 获取 phoneCode
//  2. 将 phoneCode POST 到此接口
//  3. 服务端用 access_token + phoneCode 换取真实手机号
func (s *WeChatService) GetPhoneNumber(ctx context.Context, phoneCode string) (*pkgwechat.PhoneNumberResponse, error) {
	accessToken, err := s.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return s.miniClient.GetPhoneNumber(ctx, accessToken, phoneCode)
}

// SendSubscribeMessageRequest 发送订阅消息的业务请求
type SendSubscribeMessageRequest struct {
	OpenID     string                                    `json:"open_id" binding:"required"`
	TemplateID string                                    `json:"template_id" binding:"required"`
	Page       string                                    `json:"page"`
	Data       map[string]pkgwechat.SubscribeMessageData `json:"data" binding:"required"`
}
