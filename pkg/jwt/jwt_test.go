package jwt_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	jwtutil "github.com/example/go-api-starter/pkg/jwt"
)

const (
	testSecret  = "test-secret-key-at-least-32-chars-long"
	testUserID  = uint(42)
	testOpenID  = "oXtest123"
	testRole    = "user"
)

// TestGenerateTokenPair_Success 正常生成令牌对
func TestGenerateTokenPair_Success(t *testing.T) {
	pair, err := jwtutil.GenerateTokenPair(testUserID, testOpenID, "testuser", testRole, testSecret, 72, 168)
	if err != nil {
		t.Fatalf("生成TokenPair失败: %v", err)
	}
	if pair.AccessToken == "" {
		t.Error("AccessToken 不应为空")
	}
	if pair.RefreshToken == "" {
		t.Error("RefreshToken 不应为空")
	}
	if pair.ExpiresIn <= time.Now().Unix() {
		t.Error("ExpiresIn 应为将来的时间戳")
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Error("AccessToken 和 RefreshToken 不应相同")
	}
}

// TestGenerateTokenPair_EmptySecret 空 secret 必须拒绝
func TestGenerateTokenPair_EmptySecret(t *testing.T) {
	_, err := jwtutil.GenerateTokenPair(testUserID, testOpenID, "", testRole, "", 72, 168)
	if err != jwtutil.ErrEmptySecret {
		t.Errorf("期望 ErrEmptySecret，实际: %v", err)
	}
}

// TestParseToken_ValidAccessToken 正常解析 access token
func TestParseToken_ValidAccessToken(t *testing.T) {
	pair, _ := jwtutil.GenerateTokenPair(testUserID, testOpenID, "testuser", testRole, testSecret, 72, 168)

	claims, err := jwtutil.ParseToken(pair.AccessToken, testSecret)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if claims.UserID != testUserID {
		t.Errorf("UserID: 期望 %d, 实际 %d", testUserID, claims.UserID)
	}
	if claims.OpenID != testOpenID {
		t.Errorf("OpenID: 期望 %s, 实际 %s", testOpenID, claims.OpenID)
	}
	if claims.Role != testRole {
		t.Errorf("Role: 期望 %s, 实际 %s", testRole, claims.Role)
	}
}

// TestParseToken_RefreshTokenHasCorrectRole refresh token 的 role 必须是 "refresh"
func TestParseToken_RefreshTokenHasCorrectRole(t *testing.T) {
	pair, _ := jwtutil.GenerateTokenPair(testUserID, testOpenID, "", testRole, testSecret, 72, 168)

	claims, err := jwtutil.ParseToken(pair.RefreshToken, testSecret)
	if err != nil {
		t.Fatalf("解析 refresh token 失败: %v", err)
	}
	if claims.Role != "refresh" {
		t.Errorf("refresh token 的 role 应为 'refresh'，实际: %s", claims.Role)
	}
}

// TestParseAccessToken_RefreshTokenRejected access token 解析器必须拒绝 refresh token
func TestParseAccessToken_RefreshTokenRejected(t *testing.T) {
	pair, _ := jwtutil.GenerateTokenPair(testUserID, testOpenID, "", testRole, testSecret, 72, 168)

	_, err := jwtutil.ParseAccessToken(pair.RefreshToken, testSecret)
	if err != jwtutil.ErrTokenInvalid {
		t.Errorf("ParseAccessToken 应拒绝 refresh token，实际错误: %v", err)
	}
}

// TestParseAccessToken_AccessTokenAccepted ParseAccessToken 应接受正常 access token
func TestParseAccessToken_AccessTokenAccepted(t *testing.T) {
	pair, _ := jwtutil.GenerateTokenPair(testUserID, testOpenID, "", testRole, testSecret, 72, 168)

	claims, err := jwtutil.ParseAccessToken(pair.AccessToken, testSecret)
	if err != nil {
		t.Errorf("ParseAccessToken 应接受 access token，实际错误: %v", err)
	}
	if claims.UserID != testUserID {
		t.Errorf("UserID 不匹配")
	}
}

// TestParseToken_WrongSecret 错误 secret 必须返回签名错误
func TestParseToken_WrongSecret(t *testing.T) {
	pair, _ := jwtutil.GenerateTokenPair(testUserID, testOpenID, "", testRole, testSecret, 72, 168)

	_, err := jwtutil.ParseToken(pair.AccessToken, "wrong-secret")
	if err != jwtutil.ErrTokenSignature {
		t.Errorf("期望 ErrTokenSignature，实际: %v", err)
	}
}

// TestParseToken_MalformedToken 格式错误的 token
func TestParseToken_MalformedToken(t *testing.T) {
	_, err := jwtutil.ParseToken("not.a.valid.token.string", testSecret)
	if err == nil {
		t.Error("期望返回错误")
	}
}

// TestParseToken_EmptySecret 解析时空 secret 必须拒绝
func TestParseToken_EmptySecret(t *testing.T) {
	pair, _ := jwtutil.GenerateTokenPair(testUserID, testOpenID, "", testRole, testSecret, 72, 168)

	_, err := jwtutil.ParseToken(pair.AccessToken, "")
	if err != jwtutil.ErrEmptySecret {
		t.Errorf("期望 ErrEmptySecret，实际: %v", err)
	}
}

// TestParseToken_ExpiredToken 过期 token 必须返回 ErrTokenExpired
func TestParseToken_ExpiredToken(t *testing.T) {
	// 直接构造一个已过期的 token（expireHours = -1）
	claims := &jwtutil.Claims{
		UserID: testUserID,
		Role:   testRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	tokenStr, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))

	_, err := jwtutil.ParseToken(tokenStr, testSecret)
	if err != jwtutil.ErrTokenExpired {
		t.Errorf("期望 ErrTokenExpired，实际: %v", err)
	}
}

// TestTokenPair_Symmetry access/refresh token 应互不相通
func TestTokenPair_Symmetry(t *testing.T) {
	pair, _ := jwtutil.GenerateTokenPair(testUserID, testOpenID, "", testRole, testSecret, 72, 168)

	// access token 不能当 refresh token 用（ParseToken 可解析但 role 不是 refresh）
	aClaims, err := jwtutil.ParseToken(pair.AccessToken, testSecret)
	if err != nil {
		t.Fatalf("ParseToken(access) failed: %v", err)
	}
	if aClaims.Role == "refresh" {
		t.Error("access token 的 role 不应为 'refresh'")
	}
}
