// backend/internal/pkg/apperror/apperror.go
// 自訂 AppError：攜帶錯誤碼、HTTP 狀態與訊息，供 error-handling middleware 統一轉為 JSON 回應

package apperror

import (
	"errors"
	"net/http"
)

// 業務錯誤碼常數
const (
	CodeBadRequest      = "BAD_REQUEST"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeNotFound        = "NOT_FOUND"
	CodeInternal        = "INTERNAL_ERROR"
	CodeNotImplemented  = "NOT_IMPLEMENTED"
	CodeFeatureDisabled = "FEATURE_DISABLED"
)

// AppError 統一的應用層錯誤，實作 error 介面
type AppError struct {
	Code    string // 業務錯誤碼
	Status  int    // 對應的 HTTP 狀態碼
	Message string // 對外可見的錯誤訊息
	Err     error  // 被包裝的底層錯誤（僅寫入日誌，不對外洩漏）
}

// Error 實作 error 介面
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap 支援 errors.Is / errors.As 追溯底層錯誤
func (e *AppError) Unwrap() error {
	return e.Err
}

// New 建立不含底層錯誤的 AppError
func New(code string, status int, message string) *AppError {
	return &AppError{Code: code, Status: status, Message: message}
}

// Wrap 包裝底層錯誤為 AppError
func Wrap(err error, code string, status int, message string) *AppError {
	return &AppError{Code: code, Status: status, Message: message, Err: err}
}

// NotFound 建立 404 錯誤
func NotFound(message string) *AppError {
	return New(CodeNotFound, http.StatusNotFound, message)
}

// BadRequest 建立 400 錯誤
func BadRequest(message string) *AppError {
	return New(CodeBadRequest, http.StatusBadRequest, message)
}

// Unauthorized 建立 401 錯誤
func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, http.StatusUnauthorized, message)
}

// Internal 包裝底層錯誤為 500 錯誤
func Internal(err error, message string) *AppError {
	return Wrap(err, CodeInternal, http.StatusInternalServerError, message)
}

// NotImplemented 建立 501 錯誤
func NotImplemented(message string) *AppError {
	return New(CodeNotImplemented, http.StatusNotImplemented, message)
}

// FeatureDisabled 建立 403 錯誤：功能已被管理員停用
func FeatureDisabled(message string) *AppError {
	return New(CodeFeatureDisabled, http.StatusForbidden, message)
}

// AsAppError 從錯誤鏈中取出 AppError
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
