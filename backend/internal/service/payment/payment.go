// backend/internal/service/payment/payment.go
// 金流抽象層：PaymentProvider interface 定義建立付款與驗證 callback 的行為（只做金流商適配，不碰資料庫）

package payment

import (
	"errors"
	"time"
)

// callback 驗證失敗的原因；service 依此對應綠界要求的 0|<原因> 回應
var (
	// ErrCheckMacMismatch callback 的 CheckMacValue 與重新計算的值不符
	ErrCheckMacMismatch = errors.New("CheckMacValue 不符")
	// ErrMerchantMismatch callback 的 MerchantID 與設定不同
	ErrMerchantMismatch = errors.New("MerchantID 不符")
)

// CreatePaymentInput 建立付款的輸入參數
type CreatePaymentInput struct {
	OrderID       string    // 訂單編號（綠界 MerchantTradeNo，最長 20 碼英數）
	Amount        int       // 金額（新台幣元）
	Description   string    // 交易描述（TradeDesc）
	ItemName      string    // 商品名稱（ItemName）
	TradeDate     time.Time // 交易時間（MerchantTradeDate，以 UTC+8 格式化）
	ClientBackURL string    // 付款後導回前台的完整網址
	ChoosePayment string    // 付款方式；空值時為 Credit
}

// CreatePaymentResult 建立付款的結果：前端以表單 POST 到 ActionURL，欄位為 Params
type CreatePaymentResult struct {
	ActionURL string            // 綠界 AIO 付款網址
	Params    map[string]string // 要放進表單 hidden input 的全部欄位（含 CheckMacValue）
}

// PaymentProvider 金流服務介面：各金流商（如綠界）實作此介面
type PaymentProvider interface {
	// Configured 金流設定是否齊全（不齊全時不應呼叫 CreatePayment）
	Configured() bool
	// CreatePayment 建立付款交易，回傳導向金流頁面所需資訊
	CreatePayment(input CreatePaymentInput) (*CreatePaymentResult, error)
	// VerifyCallback 驗證金流商 callback 的合法性；
	// 失敗時回傳 ErrCheckMacMismatch 或 ErrMerchantMismatch
	VerifyCallback(params map[string]string) error
}
