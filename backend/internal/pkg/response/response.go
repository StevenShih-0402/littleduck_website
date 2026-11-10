// backend/internal/pkg/response/response.go
// 統一 JSON 回應 envelope：{code, message, data}，成功與錯誤回應皆經此套件輸出

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/apperror"
)

// CodeSuccess 成功回應的固定 code
const CodeSuccess = "SUCCESS"

// Envelope 統一回應格式
type Envelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// OK 輸出 200 成功回應
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{
		Code:    CodeSuccess,
		Message: "ok",
		Data:    data,
	})
}

// Created 輸出 201 成功回應
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{
		Code:    CodeSuccess,
		Message: "created",
		Data:    data,
	})
}

// NoContentOK 輸出 204（無內容；用於非阻塞回報類端點）
func NoContentOK(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error 依 AppError 輸出對應狀態碼的錯誤回應
func Error(c *gin.Context, appErr *apperror.AppError) {
	c.JSON(appErr.Status, Envelope{
		Code:    appErr.Code,
		Message: appErr.Message,
	})
}
