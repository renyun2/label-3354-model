package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT载荷结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	OpenID   string `json:"open_id,omitempty"`
	Username string `json:"username,omitempty"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair 访问令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Unix时间戳
}

var (
	ErrTokenExpired   = errors.New("token已过期")
	ErrTokenInvalid   = errors.New("token无效")
	ErrTokenMalformed = errors.New("token格式错误")
	ErrTokenSignature = errors.New("token签名错误")
	ErrEmptySecret    = errors.New("JWT secret不能为空")
)

// GenerateTokenPair 生成访问令牌和刷新令牌
func GenerateTokenPair(userID uint, openID, username, role, secret string, expireHours, refreshExpireHours int) (*TokenPair, error) {
	// HMAC-SHA256 允许空 key，必须在此处主动拒绝
	if secret == "" {
		return nil, ErrEmptySecret
	}

	now := time.Now()
	accessExpire := now.Add(time.Duration(expireHours) * time.Hour)
	refreshExpire := now.Add(time.Duration(refreshExpireHours) * time.Hour)

	accessClaims := &Claims{
		UserID:   userID,
		OpenID:   openID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpire),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	// refresh token 使用固定 role="refresh"，与 access token 严格区分
	refreshClaims := &Claims{
		UserID: userID,
		Role:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpire),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    accessExpire.Unix(),
	}, nil
}

// ParseToken 解析并验证JWT令牌
func ParseToken(tokenString, secret string) (*Claims, error) {
	if secret == "" {
		return nil, ErrEmptySecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenSignature
		}
		return []byte(secret), nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, ErrTokenMalformed
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, ErrTokenSignature
		default:
			return nil, ErrTokenInvalid
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ParseAccessToken 解析 access token，明确拒绝 refresh token 被用于鉴权
func ParseAccessToken(tokenString, secret string) (*Claims, error) {
	claims, err := ParseToken(tokenString, secret)
	if err != nil {
		return nil, err
	}
	if claims.Role == "refresh" {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
