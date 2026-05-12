package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"resource-station/internal/models"
	"resource-station/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(user *models.User) (string, error) {
	config := utils.GetConfig().JWT

	// 设置token过期时间
	expireTime := time.Now().Add(time.Duration(config.ExpireHours) * time.Hour)

	// 创建声明
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "resource-station",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString([]byte(config.Secret))
	if err != nil {
		return "", fmt.Errorf("生成token失败: %v", err)
	}

	return tokenString, nil
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*JWTClaims, error) {
	config := utils.GetConfig().JWT

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名方法: %v", token.Header["alg"])
		}
		return []byte(config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析token失败: %v", err)
	}

	// 验证token
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// ValidateToken 验证token有效性
func ValidateToken(tokenString string) (bool, *JWTClaims) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return false, nil
	}
	return true, claims
}

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("加密密码失败: %v", err)
	}
	return string(hashedBytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetUserFromToken 从token中获取用户信息
func GetUserFromToken(tokenString string) (*models.User, error) {
	valid, claims := ValidateToken(tokenString)
	if !valid {
		return nil, errors.New("无效的token")
	}

	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}

	return user, nil
}

// MiddlewareAuth 认证中间件（将在API层实现）
func MiddlewareAuth() {
	log.Println("认证中间件已加载")
}
