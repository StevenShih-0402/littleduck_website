// backend/internal/middleware/middleware.go
// Gin middleware 集合：request logger、panic recovery、統一 error handler 與 CORS

package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/pkg/response"
)

// RequestLogger 記錄每個請求的 method、path、status、latency 與 client IP
func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// Recovery 攔截 handler panic：記錄 stack trace 後回應 500，伺服器繼續服務
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("path", c.Request.URL.Path),
					zap.Stack("stack"),
				)
				response.Error(c, apperror.New(
					apperror.CodeInternal, http.StatusInternalServerError, "internal server error",
				))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ErrorHandler 統一轉換 handler 掛入 c.Errors 的錯誤：
// AppError 依其狀態碼回應，未知錯誤一律 500 且不洩漏內部細節
func ErrorHandler(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		if appErr, ok := apperror.AsAppError(err); ok {
			if appErr.Err != nil {
				log.Error("app error", zap.String("code", appErr.Code), zap.Error(appErr.Err))
			}
			response.Error(c, appErr)
			return
		}

		log.Error("unhandled error", zap.Error(err))
		response.Error(c, apperror.New(
			apperror.CodeInternal, http.StatusInternalServerError, "internal server error",
		))
	}
}

// CORS 允許跨來源請求（骨架階段全開，正式部署時再收斂 origin 白名單）
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
