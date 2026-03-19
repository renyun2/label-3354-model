package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/example/go-api-starter/internal/service"
	pkgwechat "github.com/example/go-api-starter/pkg/wechat"
	"github.com/example/go-api-starter/pkg/response"
)

// WeChatHandler 微信相关处理器
type WeChatHandler struct {
	wechatSvc *service.WeChatService
}

// NewWeChatHandler 创建微信处理器
func NewWeChatHandler(wechatSvc *service.WeChatService) *WeChatHandler {
	return &WeChatHandler{wechatSvc: wechatSvc}
}

// SendSubscribeMessage godoc
// @Summary     发送微信订阅消息
// @Description 向指定用户发送微信订阅消息（需用户授权订阅）
// @Tags        微信
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body service.SendSubscribeMessageRequest true "消息内容"
// @Success     200 {object} response.Response
// @Router      /api/v1/wechat/message/subscribe [post]
func (h *WeChatHandler) SendSubscribeMessage(c *gin.Context) {
	var req service.SendSubscribeMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	msgReq := &pkgwechat.SubscribeMessageRequest{
		ToUser:     req.OpenID,
		TemplateID: req.TemplateID,
		Page:       req.Page,
		Data:       req.Data,
	}

	if err := h.wechatSvc.SendSubscribeMessage(c.Request.Context(), msgReq); err != nil {
		response.Fail(c, response.CodeWeChatAPIError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "消息发送成功", nil)
}

// GetPhoneNumber godoc
// @Summary     获取微信手机号
// @Description 小程序端调用 wx.getPhoneNumber() 获取 code，再通过此接口换取真实手机号
// @Tags        微信
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body object{code=string} true "小程序 phoneCode"
// @Success     200 {object} response.Response{data=wechat.PhoneNumberResponse}
// @Router      /api/v1/wechat/phone [post]
func (h *WeChatHandler) GetPhoneNumber(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	result, err := h.wechatSvc.GetPhoneNumber(c.Request.Context(), req.Code)
	if err != nil {
		response.Fail(c, response.CodeWeChatAPIError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"phone_number":      result.PhoneInfo.PhoneNumber,
		"pure_phone_number": result.PhoneInfo.PurePhoneNumber,
		"country_code":      result.PhoneInfo.CountryCode,
	})
}

// SendCustomerMessage godoc
// @Summary     发送客服消息
// @Description 向指定用户发送客服文字消息
// @Tags        微信
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body object{open_id=string,content=string} true "消息内容"
// @Success     200 {object} response.Response
// @Router      /api/v1/wechat/message/customer [post]
func (h *WeChatHandler) SendCustomerMessage(c *gin.Context) {
	var req struct {
		OpenID  string `json:"open_id" binding:"required"`
		Content string `json:"content" binding:"required,max=2048"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if err := h.wechatSvc.SendCustomerServiceMessage(c.Request.Context(), req.OpenID, req.Content); err != nil {
		response.Fail(c, response.CodeWeChatAPIError, err.Error())
		return
	}

	response.SuccessWithMessage(c, "客服消息发送成功", nil)
}
