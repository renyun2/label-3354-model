package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一API响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// 业务错误码定义
const (
	CodeSuccess           = 0
	CodeParamError        = 400
	CodeUnauthorized      = 401
	CodeForbidden         = 403
	CodeNotFound          = 404
	CodeInternalError     = 500

	// 微信相关错误码
	CodeWeChatAuthFailed  = 10001
	CodeWeChatAPIError    = 10002

	// 用户相关错误码
	CodeUserNotFound      = 20001
	CodeUserAlreadyExists = 20002
	CodePasswordError     = 20003
	CodeTokenExpired      = 20004
	CodeTokenInvalid      = 20005
)

var codeMessages = map[int]string{
	CodeSuccess:           "成功",
	CodeParamError:        "参数错误",
	CodeUnauthorized:      "未授权",
	CodeForbidden:         "无权限",
	CodeNotFound:          "资源不存在",
	CodeInternalError:     "服务器内部错误",
	CodeWeChatAuthFailed:  "微信授权失败",
	CodeWeChatAPIError:    "微信API调用失败",
	CodeUserNotFound:      "用户不存在",
	CodeUserAlreadyExists: "用户已存在",
	CodePasswordError:     "密码错误",
	CodeTokenExpired:      "Token已过期",
	CodeTokenInvalid:      "Token无效",
}

func getMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "成功",
		Data:    data,
		TraceID: getTraceID(c),
	})
}

// SuccessWithMessage 自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
		TraceID: getTraceID(c),
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, message ...string) {
	msg := getMessage(code)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
		TraceID: getTraceID(c),
	})
}

// ParamError 参数错误
func ParamError(c *gin.Context, message ...string) {
	Fail(c, CodeParamError, message...)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message ...string) {
	msg := "未授权，请先登录"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: msg,
		TraceID: getTraceID(c),
	})
}

// Forbidden 无权限（HTTP 403 + Abort，与 Unauthorized 行为一致，防止后续 handler 继续执行）
func Forbidden(c *gin.Context, message ...string) {
	msg := getMessage(CodeForbidden)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	c.AbortWithStatusJSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: msg,
		TraceID: getTraceID(c),
	})
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message ...string) {
	Fail(c, CodeNotFound, message...)
}

// InternalError 服务器内部错误
func InternalError(c *gin.Context, message ...string) {
	Fail(c, CodeInternalError, message...)
}

// Page 分页响应
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
