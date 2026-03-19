package wechat

import (
	"net/http"
	"time"
)

// Client 微信综合客户端（整合小程序和消息功能）
type Client struct {
	MiniProgram *MiniProgramClient
	Message     *MessageClient
}

// NewClient 创建微信综合客户端
func NewClient(appID, appSecret string) *Client {
	return &Client{
		MiniProgram: NewMiniProgramClient(appID, appSecret),
		Message:     NewMessageClient(),
	}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}
