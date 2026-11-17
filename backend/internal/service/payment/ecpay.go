// backend/internal/service/payment/ecpay.go
// 綠界（ECPay）全方位金流（AIO）適配：組付款表單欄位與 CheckMacValue、驗證付款結果通知

package payment

import (
	"strconv"
	"time"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/pkg/apperror"
)

// 綠界 AIO 付款網址
const (
	ecpayStageURL      = "https://payment-stage.ecpay.com.tw/Cashier/AioCheckOut/V5"
	ecpayProductionURL = "https://payment.ecpay.com.tw/Cashier/AioCheckOut/V5"
)

// ChoosePaymentCredit 只開信用卡
const ChoosePaymentCredit = "Credit"

// ecpayTradeDateLayout 綠界 MerchantTradeDate 格式
const ecpayTradeDateLayout = "2006/01/02 15:04:05"

// taipeiZone 綠界交易時間使用的固定 UTC+8 時區
var taipeiZone = time.FixedZone("Asia/Taipei", 8*60*60)

// ECPayService PaymentProvider 的綠界實作
type ECPayService struct {
	cfg config.ECPayConfig
}

// NewECPayService 以 config 中的綠界設定（ECPAY_*）建立 ECPayService
func NewECPayService(cfg config.ECPayConfig) *ECPayService {
	return &ECPayService{cfg: cfg}
}

// Configured 設定是否齊全
func (s *ECPayService) Configured() bool {
	return s.cfg.Configured()
}

// actionURL 依 ECPAY_ENV 選擇綠界網址：production 用正式環境，其餘一律測試環境
func (s *ECPayService) actionURL() string {
	if s.cfg.Env == "production" {
		return ecpayProductionURL
	}
	return ecpayStageURL
}

// CreatePayment 組出前端表單 POST 到綠界所需的網址與欄位（含 CheckMacValue）
func (s *ECPayService) CreatePayment(input CreatePaymentInput) (*CreatePaymentResult, error) {
	if !s.Configured() {
		return nil, apperror.PaymentUnavailable("付款功能暫時無法使用")
	}
	choose := input.ChoosePayment
	if choose == "" {
		choose = ChoosePaymentCredit
	}
	params := map[string]string{
		"MerchantID":        s.cfg.MerchantID,
		"MerchantTradeNo":   input.OrderID,
		"MerchantTradeDate": input.TradeDate.In(taipeiZone).Format(ecpayTradeDateLayout),
		"PaymentType":       "aio",
		"TotalAmount":       strconv.Itoa(input.Amount),
		"TradeDesc":         input.Description,
		"ItemName":          input.ItemName,
		"ReturnURL":         s.cfg.ReturnURL,
		"ClientBackURL":     input.ClientBackURL,
		"ChoosePayment":     choose,
		"EncryptType":       "1",
	}
	params["CheckMacValue"] = CheckMacValue(params, s.cfg.HashKey, s.cfg.HashIV)
	return &CreatePaymentResult{ActionURL: s.actionURL(), Params: params}, nil
}

// VerifyCallback 驗證綠界 callback：先驗 CheckMacValue，再比對 MerchantID
func (s *ECPayService) VerifyCallback(params map[string]string) error {
	if !VerifyCheckMac(params, s.cfg.HashKey, s.cfg.HashIV) {
		return ErrCheckMacMismatch
	}
	if params["MerchantID"] != s.cfg.MerchantID {
		return ErrMerchantMismatch
	}
	return nil
}

// MerchantID 回傳設定的商店代號
func (s *ECPayService) MerchantID() string {
	return s.cfg.MerchantID
}
