// backend/internal/service/payment/ecpay_test.go
// ECPay service 單元測試：設定載入、stage／production 網址、付款欄位完整、callback 驗證

package payment

import (
	"net/http"
	"testing"
	"time"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/pkg/apperror"
)

func testECPayConfig() config.ECPayConfig {
	return config.ECPayConfig{
		MerchantID:    "3002607",
		HashKey:       docHashKey,
		HashIV:        docHashIV,
		Env:           "stage",
		ReturnURL:     "https://api.example.com/api/v1/payments/ecpay/callback",
		ClientBackURL: "https://www.example.com/sponsor/result",
	}
}

func testInput() CreatePaymentInput {
	return CreatePaymentInput{
		OrderID:       "LD261117153012A9K2QZ",
		Amount:        100,
		Description:   "LittleDuck 贊助",
		ItemName:      "請小鴨喝蜂蜜檸檬 NT$100",
		TradeDate:     time.Date(2026, 11, 17, 7, 30, 12, 0, time.UTC), // UTC+8 為 15:30:12
		ClientBackURL: "https://www.example.com/sponsor/result?trade_no=LD261117153012A9K2QZ",
	}
}

func TestNewECPayService_LoadsConfigFromEnvironment(t *testing.T) {
	// Arrange & Act
	svc := NewECPayService(testECPayConfig())

	// Assert
	if svc.MerchantID() != "3002607" {
		t.Errorf("MerchantID = %q", svc.MerchantID())
	}
	if !svc.Configured() {
		t.Error("設定齊全時 Configured 應為 true")
	}
	if NewECPayService(config.ECPayConfig{}).Configured() {
		t.Error("空設定時 Configured 應為 false")
	}
}

func TestECPayCreatePayment_StageURLAndCompleteParams(t *testing.T) {
	// Arrange
	svc := NewECPayService(testECPayConfig())

	// Act
	res, err := svc.CreatePayment(testInput())

	// Assert
	if err != nil {
		t.Fatalf("CreatePayment 應成功: %v", err)
	}
	if res.ActionURL != "https://payment-stage.ecpay.com.tw/Cashier/AioCheckOut/V5" {
		t.Errorf("ActionURL = %q", res.ActionURL)
	}
	want := map[string]string{
		"MerchantID":        "3002607",
		"MerchantTradeNo":   "LD261117153012A9K2QZ",
		"MerchantTradeDate": "2026/11/17 15:30:12",
		"PaymentType":       "aio",
		"TotalAmount":       "100",
		"TradeDesc":         "LittleDuck 贊助",
		"ItemName":          "請小鴨喝蜂蜜檸檬 NT$100",
		"ReturnURL":         "https://api.example.com/api/v1/payments/ecpay/callback",
		"ClientBackURL":     "https://www.example.com/sponsor/result?trade_no=LD261117153012A9K2QZ",
		"ChoosePayment":     "Credit",
		"EncryptType":       "1",
	}
	for k, v := range want {
		if res.Params[k] != v {
			t.Errorf("Params[%s] = %q, want %q", k, res.Params[k], v)
		}
	}
	if len(res.Params) != len(want)+1 {
		t.Errorf("Params 應為 %d 個欄位（含 CheckMacValue），得到 %d", len(want)+1, len(res.Params))
	}
	if len(res.Params["CheckMacValue"]) != 64 {
		t.Errorf("CheckMacValue 應為 64 碼，得到 %q", res.Params["CheckMacValue"])
	}
	if !VerifyCheckMac(res.Params, docHashKey, docHashIV) {
		t.Error("產生的 CheckMacValue 應能被同一組 HashKey／HashIV 驗證")
	}
}

func TestECPayCreatePayment_ProductionURL(t *testing.T) {
	// Arrange
	cfg := testECPayConfig()
	cfg.Env = "production"

	// Act
	res, err := NewECPayService(cfg).CreatePayment(testInput())

	// Assert
	if err != nil {
		t.Fatalf("CreatePayment 應成功: %v", err)
	}
	if res.ActionURL != "https://payment.ecpay.com.tw/Cashier/AioCheckOut/V5" {
		t.Errorf("ActionURL = %q", res.ActionURL)
	}
}

func TestECPayCreatePayment_UnknownEnvFallsBackToStage(t *testing.T) {
	// Arrange
	cfg := testECPayConfig()
	cfg.Env = "whatever"

	// Act
	res, _ := NewECPayService(cfg).CreatePayment(testInput())

	// Assert
	if res.ActionURL != ecpayStageURL {
		t.Errorf("非 production 應用測試環境網址，得到 %q", res.ActionURL)
	}
}

func TestECPayCreatePayment_ExplicitChoosePayment(t *testing.T) {
	// Arrange
	in := testInput()
	in.ChoosePayment = "ALL"

	// Act
	res, _ := NewECPayService(testECPayConfig()).CreatePayment(in)

	// Assert
	if res.Params["ChoosePayment"] != "ALL" {
		t.Errorf("ChoosePayment = %q", res.Params["ChoosePayment"])
	}
}

func TestECPayCreatePayment_UnconfiguredReturns503(t *testing.T) {
	// Arrange
	svc := NewECPayService(config.ECPayConfig{})

	// Act
	res, err := svc.CreatePayment(testInput())

	// Assert
	if res != nil {
		t.Errorf("設定不完整不應回傳結果: %+v", res)
	}
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusServiceUnavailable || appErr.Code != apperror.CodePaymentUnavailable {
		t.Errorf("應回傳 503 PAYMENT_UNAVAILABLE，得到 %v", err)
	}
}

func signedCallback(cfg config.ECPayConfig, mutate func(map[string]string)) map[string]string {
	p := map[string]string{
		"MerchantID":      cfg.MerchantID,
		"MerchantTradeNo": "LD261117153012A9K2QZ",
		"RtnCode":         "1",
		"TradeAmt":        "100",
		"TradeNo":         "2611171530120001",
		"PaymentType":     "Credit_CreditCard",
	}
	if mutate != nil {
		mutate(p)
	}
	p["CheckMacValue"] = CheckMacValue(p, cfg.HashKey, cfg.HashIV)
	return p
}

func TestECPayVerifyCallback_Valid(t *testing.T) {
	// Arrange
	cfg := testECPayConfig()

	// Act
	err := NewECPayService(cfg).VerifyCallback(signedCallback(cfg, nil))

	// Assert
	if err != nil {
		t.Errorf("合法通知應驗證成功: %v", err)
	}
}

func TestECPayVerifyCallback_BadCheckMac(t *testing.T) {
	// Arrange
	cfg := testECPayConfig()
	p := signedCallback(cfg, nil)
	p["TradeAmt"] = "1" // 竄改後檢查碼失效

	// Act
	err := NewECPayService(cfg).VerifyCallback(p)

	// Assert
	if err != ErrCheckMacMismatch {
		t.Errorf("err = %v, want ErrCheckMacMismatch", err)
	}
}

func TestECPayVerifyCallback_MerchantMismatch(t *testing.T) {
	// Arrange：檢查碼正確但商店代號不同
	cfg := testECPayConfig()
	p := signedCallback(cfg, func(p map[string]string) { p["MerchantID"] = "9999999" })

	// Act
	err := NewECPayService(cfg).VerifyCallback(p)

	// Assert
	if err != ErrMerchantMismatch {
		t.Errorf("err = %v, want ErrMerchantMismatch", err)
	}
}

func TestECPayVerifyCallback_MissingCheckMac(t *testing.T) {
	// Arrange
	cfg := testECPayConfig()
	p := signedCallback(cfg, nil)
	delete(p, "CheckMacValue")

	// Act
	err := NewECPayService(cfg).VerifyCallback(p)

	// Assert
	if err != ErrCheckMacMismatch {
		t.Errorf("err = %v, want ErrCheckMacMismatch", err)
	}
}
