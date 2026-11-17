// backend/internal/controller/payment_controller.go
// Payment controller：建立贊助付款、接收綠界付款結果通知（純文字回應）、查詢付款狀態

package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/middleware"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/pkg/response"
	"littleduck-backend/internal/service"
)

// PaymentController 金流 controller
type PaymentController struct {
	svc service.PaymentService
}

// NewPaymentController 建立 PaymentController
func NewPaymentController(svc service.PaymentService) *PaymentController {
	return &PaymentController{svc: svc}
}

// Create 建立贊助付款
//
//	@Summary		建立贊助付款
//	@Description	建立一筆 pending 贊助紀錄，回傳綠界 AIO 的表單網址與欄位（前端以 POST 表單導向）。金額 10～20000、message ≤100 字、name ≤50 字。
//	@Description	非管理員時 sponsor 與 sponsor_payment 皆需啟用，否則 403 FEATURE_DISABLED；綠界設定不完整回 503 PAYMENT_UNAVAILABLE。
//	@Description	可選帶管理員 Authorization: Bearer；權杖有效時略過功能開關（管理員預覽，會建立真實訂單），無效或過期的權杖視同訪客、不回 401。
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					false	"選用：Bearer <管理員 JWT>"
//	@Param			body			body		service.PaymentInput	true	"贊助內容"
//	@Success		200				{object}	response.Envelope{data=service.CreatePaymentView}
//	@Failure		400				{object}	response.Envelope
//	@Failure		403				{object}	response.Envelope	"功能已停用（FEATURE_DISABLED）"
//	@Failure		500				{object}	response.Envelope
//	@Failure		503				{object}	response.Envelope	"綠界設定不完整（PAYMENT_UNAVAILABLE）"
//	@Router			/api/v1/payments [post]
func (ctrl *PaymentController) Create(c *gin.Context) {
	var in service.PaymentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperror.BadRequest("付款內容格式錯誤"))
		return
	}
	_, asAdmin := c.Get(middleware.ContextKeyAdmin)
	view, err := ctrl.svc.Create(in, asAdmin)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, view)
}

// ECPayCallback 綠界付款結果通知（ReturnURL）
//
//	@Summary		綠界付款結果通知
//	@Description	由綠界 server-to-server 呼叫，不受功能開關影響。一律 HTTP 200 text/plain：成功或重複通知回 1|OK；
//	@Description	驗證失敗回 0|CheckMacValue Error、0|MerchantID Error、0|Order Not Found、0|Amount Mismatch；資料庫錯誤回 0|Internal Error（綠界會重送）。
//	@Tags			payments
//	@Accept			x-www-form-urlencoded
//	@Produce		plain
//	@Success		200	{string}	string	"1|OK 或 0|原因"
//	@Router			/api/v1/payments/ecpay/callback [post]
func (ctrl *PaymentController) ECPayCallback(c *gin.Context) {
	result := "0|Internal Error"
	if err := c.Request.ParseForm(); err == nil {
		params := make(map[string]string, len(c.Request.PostForm))
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
		result = ctrl.svc.HandleECPayCallback(params)
	}
	c.String(http.StatusOK, result)
}

// Status 查詢付款狀態
//
//	@Summary		查詢付款狀態
//	@Description	以交易編號查詢，只回傳 trade_no、amount、status，不含贊助者名稱與留言；不受功能開關影響
//	@Tags			payments
//	@Produce		json
//	@Param			trade_no	path		string	true	"交易編號（建立付款回傳的 trade_no）"
//	@Success		200			{object}	response.Envelope{data=service.PaymentStatusView}
//	@Failure		404			{object}	response.Envelope
//	@Failure		500			{object}	response.Envelope
//	@Router			/api/v1/payments/{trade_no} [get]
func (ctrl *PaymentController) Status(c *gin.Context) {
	view, err := ctrl.svc.Status(c.Param("trade_no"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, view)
}
