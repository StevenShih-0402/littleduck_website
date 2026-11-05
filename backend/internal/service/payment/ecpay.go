// backend/internal/service/payment/ecpay.go
// 綠界（ECPay）金流服務骨架：由環境變數載入設定，正式串接（CheckMacValue、交易流程）留待後續 change

package payment

import (
	"littleduck-backend/internal/config"
	"littleduck-backend/internal/pkg/apperror"
)

// ECPayService PaymentProvider 的綠界實作骨架
type ECPayService struct {
	cfg config.ECPayConfig
}

// NewECPayService 以 config 中的綠界設定（ECPAY_*）建立 ECPayService
func NewECPayService(cfg config.ECPayConfig) *ECPayService {
	return &ECPayService{cfg: cfg}
}

// CreatePayment 建立綠界付款交易（骨架階段回傳未實作）
func (s *ECPayService) CreatePayment(input CreatePaymentInput) (*CreatePaymentResult, error) {
	return nil, apperror.NotImplemented("綠界付款尚未實作")
}

// VerifyCallback 驗證綠界 callback 的 CheckMacValue（骨架階段回傳未實作）
func (s *ECPayService) VerifyCallback(params map[string]string) error {
	return apperror.NotImplemented("綠界 callback 驗證尚未實作")
}

// MerchantID 回傳設定的商店代號（供後續實作與測試檢視設定載入）
func (s *ECPayService) MerchantID() string {
	return s.cfg.MerchantID
}
