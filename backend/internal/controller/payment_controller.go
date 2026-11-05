// backend/internal/controller/payment_controller.go
// Payment controller：金流端點骨架，正式串接前一律回應 501 Not Implemented

package controller

import (
	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/service/payment"
)

// PaymentController 金流 controller
type PaymentController struct {
	provider payment.PaymentProvider
}

// NewPaymentController 建立 PaymentController
func NewPaymentController(provider payment.PaymentProvider) *PaymentController {
	return &PaymentController{provider: provider}
}

// Create 建立付款
//
//	@Summary		建立付款（未實作）
//	@Description	建立綠界付款交易；骨架階段回傳 501
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Failure		501	{object}	response.Envelope
//	@Router			/api/v1/payments [post]
func (ctrl *PaymentController) Create(c *gin.Context) {
	_ = c.Error(apperror.NotImplemented("付款功能尚未實作"))
}

// ECPayCallback 綠界付款結果 callback
//
//	@Summary		綠界付款 callback（未實作）
//	@Description	接收綠界付款結果通知；骨架階段回傳 501
//	@Tags			payments
//	@Accept			x-www-form-urlencoded
//	@Produce		json
//	@Failure		501	{object}	response.Envelope
//	@Router			/api/v1/payments/ecpay/callback [post]
func (ctrl *PaymentController) ECPayCallback(c *gin.Context) {
	_ = c.Error(apperror.NotImplemented("綠界 callback 尚未實作"))
}
