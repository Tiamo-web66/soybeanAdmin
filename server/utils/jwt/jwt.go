package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"soybean-admin-go/global"
)

// CustomClaims 自定义声明
type CustomClaims struct {
	UserID   uint     `json:"userId"`
	UserName string   `json:"userName"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// TokenType 令牌类型
type TokenType string

const (
	// AccessToken 访问令牌
	AccessToken TokenType = "access"
	// RefreshToken 刷新令牌
	RefreshToken TokenType = "refresh"
)

var (
	// TokenExpired 令牌已过期
	TokenExpired = errors.New("token is expired")
	// TokenMalformed 令牌格式错误
	TokenMalformed = errors.New("token is malformed")
	// TokenInvalid 令牌无效
	TokenInvalid = errors.New("token is invalid")
	// TokenNotValidYet 令牌尚未生效
	TokenNotValidYet = errors.New("token not active yet")
)

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, userName string, roles []string, tokenType TokenType) (string, error) {
	// 确定过期时间
	var expDuration time.Duration
	if tokenType == AccessToken {
		expDuration = time.Duration(global.CONFIG.JWT.Expiration) * time.Hour
	} else {
		// RefreshToken 有效期通常更长，这里设置为AccessToken的2倍
		expDuration = time.Duration(global.CONFIG.JWT.Expiration*2) * time.Hour
	}

	// 添加随机抖动，确保每次生成的token不同（最多10分钟）
	jitter := time.Duration(time.Now().UnixNano()%600) * time.Second
	expTime := time.Now().Add(expDuration).Add(jitter)

	claims := CustomClaims{
		UserID:   userID,
		UserName: userName,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   string(tokenType),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(global.CONFIG.JWT.Secret))
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(global.CONFIG.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, TokenExpired
		}
		return nil, TokenInvalid
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, TokenInvalid
}
