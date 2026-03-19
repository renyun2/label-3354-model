package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	code2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"
	accessTokenURL  = "https://api.weixin.qq.com/cgi-bin/token"
)

// MiniProgramClient 微信小程序客户端
type MiniProgramClient struct {
	AppID     string
	AppSecret string
	httpClient *http.Client
}

// NewMiniProgramClient 创建微信小程序客户端
func NewMiniProgramClient(appID, appSecret string) *MiniProgramClient {
	return &MiniProgramClient{
		AppID:     appID,
		AppSecret: appSecret,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Code2SessionResponse 微信code换取session响应
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Code2Session 微信小程序登录：通过code获取openid和session_key
func (c *MiniProgramClient) Code2Session(ctx context.Context, code string) (*Code2SessionResponse, error) {
	url := fmt.Sprintf("%s?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		code2SessionURL, c.AppID, c.AppSecret, code)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求微信API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result Code2SessionResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信API错误 [%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// AccessTokenResponse 获取access_token响应
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// GetAccessToken 获取微信接口调用凭证
func (c *MiniProgramClient) GetAccessToken(ctx context.Context) (*AccessTokenResponse, error) {
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s",
		accessTokenURL, c.AppID, c.AppSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求微信API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result AccessTokenResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信API错误 [%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// PhoneNumberResponse 获取手机号响应
type PhoneNumberResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     string `json:"countryCode"`
	} `json:"phone_info"`
}

// GetPhoneNumber 获取用户手机号（需要用户授权）
func (c *MiniProgramClient) GetPhoneNumber(ctx context.Context, accessToken, code string) (*PhoneNumberResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", accessToken)

	payload := fmt.Sprintf(`{"code":"%s"}`, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url,
		newStringReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result PhoneNumberResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信API错误 [%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
