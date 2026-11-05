// backend/internal/service/payment/payment.go
// 金流抽象層：PaymentProvider interface 定義建立付款與驗證 callback 的行為

package payment

// CreatePaymentInput 建立付款的輸入參數
type CreatePaymentInput struct {
	OrderID     string // 訂單編號
	Amount      int    // 金額（新台幣元）
	Description string // 交易描述
}

// CreatePaymentResult 建立付款的結果
type CreatePaymentResult struct {
	PaymentURL string // 導向金流頁面的 URL
}

// PaymentProvider 金流服務介面：各金流商（如綠界）實作此介面
type PaymentProvider interface {
	// CreatePayment 建立付款交易，回傳導向金流頁面所需資訊
	CreatePayment(input CreatePaymentInput) (*CreatePaymentResult, error)
	// VerifyCallback 驗證金流商 callback 的合法性（如綠界 CheckMacValue）
	VerifyCallback(params map[string]string) error
}
