// backend/internal/service/auth/auth.go
// Auth service：以環境變數設定的單一管理員帳密驗證，簽發與解析 HS256 JWT

package auth

import (
	"crypto/subtle"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/pkg/apperror"
)

// Token 登入成功後回傳的權杖資訊
type Token struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Service 認證業務邏輯介面（controller／middleware 依賴此介面，測試以 mock 替換）
type Service interface {
	// Login 以帳密換發 JWT；帳密不符回傳 401 AppError
	Login(username, password string) (*Token, error)
	// Parse 驗證並解析權杖；無效／過期回傳 401 AppError，成功回傳 subject（管理員帳號）
	Parse(tokenString string) (string, error)
}

// authService Service 的實作，憑證與秘鑰來自 config
type authService struct {
	username string
	password string
	secret   []byte
	ttl      time.Duration
	now      func() time.Time // 可注入時間來源，供測試
}

// NewService 以 AuthConfig 建立 Service
func NewService(cfg config.AuthConfig) Service {
	return &authService{
		username: cfg.AdminUsername,
		password: cfg.AdminPassword,
		secret:   []byte(cfg.JWTSecret),
		ttl:      cfg.JWTTTL,
		now:      time.Now,
	}
}

// Login 以 constant-time 比對帳密，成功簽發 HS256 權杖
func (s *authService) Login(username, password string) (*Token, error) {
	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(s.username)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(s.password)) == 1
	if !userOK || !passOK {
		return nil, apperror.Unauthorized("帳號或密碼錯誤")
	}

	now := s.now()
	expiresAt := now.Add(s.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   s.username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return nil, apperror.Internal(err, "權杖簽發失敗")
	}
	return &Token{Token: signed, ExpiresAt: expiresAt}, nil
}

// Parse 驗證簽章與效期，成功回傳 subject
func (s *authService) Parse(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperror.Unauthorized("權杖簽章方法不符")
		}
		return s.secret, nil
	}, jwt.WithTimeFunc(s.now))
	if err != nil || !token.Valid {
		return "", apperror.Unauthorized("權杖無效或已過期")
	}
	return claims.Subject, nil
}
