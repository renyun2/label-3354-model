package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	subscribeMessageURL = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send"
	uniformMessageURL   = "https://api.weixin.qq.com/cgi-bin/message/wxopen/template/uniform_send"
)

// SubscribeMessageData 订阅消息数据项
type SubscribeMessageData struct {
	Value string `json:"value"`
}

// SubscribeMessageRequest 发送订阅消息请求体
type SubscribeMessageRequest struct {
	ToUser           string                          `json:"touser"`            // 接收者openid
	TemplateID       string                          `json:"template_id"`       // 模板ID
	Page             string                          `json:"page,omitempty"`    // 跳转页面
	Data             map[string]SubscribeMessageData `json:"data"`              // 消息数据
	MiniprogramState string                          `json:"miniprogram_state"` // developer|trial|formal
	Lang             string                          `json:"lang,omitempty"`    // zh_CN
}

// WeChatAPIResponse 微信通用API响应
type WeChatAPIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// MessageClient 微信消息客户端
type MessageClient struct {
	httpClient *http.Client
}

// NewMessageClient 创建消息客户端
func NewMessageClient() *MessageClient {
	return &MessageClient{
		httpClient: defaultHTTPClient(),
	}
}

// SendSubscribeMessage 发送订阅消息
//
// 示例：
//
//	err := client.SendSubscribeMessage(ctx, accessToken, &SubscribeMessageRequest{
//	    ToUser:     "openid_xxx",
//	    TemplateID: "template_id_xxx",
//	    Page:       "pages/index/index",
//	    Data: map[string]SubscribeMessageData{
//	        "thing1": {Value: "您的订单已发货"},
//	        "time2":  {Value: "2024-01-01 12:00"},
//	    },
//	    MiniprogramState: "formal",
//	})
func (c *MessageClient) SendSubscribeMessage(ctx context.Context, accessToken string, msg *SubscribeMessageRequest) error {
	if msg.Lang == "" {
		msg.Lang = "zh_CN"
	}
	if msg.MiniprogramState == "" {
		msg.MiniprogramState = "formal"
	}

	url := fmt.Sprintf("%s?access_token=%s", subscribeMessageURL, accessToken)
	return c.postJSON(ctx, url, msg)
}

// CustomerServiceMessage 客服消息
type CustomerServiceMessage struct {
	ToUser  string                 `json:"touser"`
	MsgType string                 `json:"msgtype"`
	Text    *CustomerServiceText   `json:"text,omitempty"`
	Image   *CustomerServiceImage  `json:"image,omitempty"`
}

type CustomerServiceText struct {
	Content string `json:"content"`
}

type CustomerServiceImage struct {
	MediaID string `json:"media_id"`
}

// SendCustomerServiceTextMessage 发送客服文字消息
func (c *MessageClient) SendCustomerServiceTextMessage(ctx context.Context, accessToken, openID, content string) error {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", accessToken)
	msg := &CustomerServiceMessage{
		ToUser:  openID,
		MsgType: "text",
		Text:    &CustomerServiceText{Content: content},
	}
	return c.postJSON(ctx, url, msg)
}

func (c *MessageClient) postJSON(ctx context.Context, url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var result WeChatAPIResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("微信API错误 [%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return nil
}

// newStringReader 辅助创建字符串读取器
func newStringReader(s string) io.Reader {
	return strings.NewReader(s)
}
