// backend/internal/middleware/auth.go
// Auth middleware：驗證 Authorization: Bearer <jwt>，掛在受保護的後台路由群組

package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/service/auth"
)

// ContextKeyAdmin gin.Context 中存放通過驗證之管理員帳號的鍵
const ContextKeyAdmin = "admin_subject"

// Auth 以 auth.Service 驗證 Bearer 權杖；缺失／格式錯誤／過期一律 401 並中止
func Auth(svc auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			abortUnauthorized(c, "缺少 Authorization 標頭")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			abortUnauthorized(c, "Authorization 標頭格式錯誤")
			return
		}

		subject, err := svc.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Set(ContextKeyAdmin, subject)
		c.Next()
	}
}

// abortUnauthorized 以統一錯誤格式回 401 並中止
func abortUnauthorized(c *gin.Context, message string) {
	_ = c.Error(apperror.Unauthorized(message))
	c.Abort()
}
