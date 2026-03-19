package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/example/go-api-starter/internal/service"
	"github.com/example/go-api-starter/pkg/response"
)

// UserHandler 用户处理器
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// WeChatLogin godoc
// @Summary     微信小程序登录
// @Description 通过微信code换取用户token
// @Tags        用户
// @Accept      json
// @Produce     json
// @Param       body body service.WeChatLoginRequest true "登录参数"
// @Success     200 {object} response.Response{data=service.LoginResponse}
// @Router      /api/v1/auth/wechat-login [post]
func (h *UserHandler) WeChatLogin(c *gin.Context) {
	var req service.WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	result, err := h.userSvc.WeChatLogin(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, response.CodeWeChatAuthFailed, err.Error())
		return
	}

	response.Success(c, result)
}

// RefreshToken godoc
// @Summary     刷新访问令牌
// @Tags        用户
// @Accept      json
// @Produce     json
// @Param       body body object{refresh_token=string} true "刷新token"
// @Success     200 {object} response.Response
// @Router      /api/v1/auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	token, err := h.userSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Fail(c, response.CodeTokenInvalid, err.Error())
		return
	}

	response.Success(c, token)
}

// GetProfile godoc
// @Summary     获取当前用户信息
// @Tags        用户
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.Response{data=model.User}
// @Router      /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		response.Unauthorized(c)
		return
	}

	user, err := h.userSvc.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, user)
}

// UpdateProfile godoc
// @Summary     更新用户资料
// @Tags        用户
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body service.UpdateProfileRequest true "用户资料"
// @Success     200 {object} response.Response{data=model.User}
// @Router      /api/v1/user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		response.Unauthorized(c)
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	user, err := h.userSvc.UpdateProfile(c.Request.Context(), claims.UserID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, user)
}

// BindPhone godoc
// @Summary     绑定手机号
// @Tags        用户
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body object{phone=string} true "手机号"
// @Success     200 {object} response.Response
// @Router      /api/v1/user/bind-phone [post]
func (h *UserHandler) BindPhone(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		response.Unauthorized(c)
		return
	}

	var req struct {
		Phone string `json:"phone" binding:"required,len=11"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if err := h.userSvc.BindPhone(c.Request.Context(), claims.UserID, req.Phone); err != nil {
		if errors.Is(err, service.ErrPhoneAlreadyBound) {
			response.Fail(c, response.CodeParamError, "该手机号已被绑定")
			return
		}
		response.InternalError(c)
		return
	}

	response.SuccessWithMessage(c, "手机号绑定成功", nil)
}
