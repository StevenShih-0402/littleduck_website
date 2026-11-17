// backend/internal/service/payment_service_test.go
// payment service 單元測試：建立付款（開關、金額與文字邊界、交易編號、重複重試）、
// 綠界通知處理（判斷順序、冪等、paid 不被改回）、狀態查詢（mock repository，真實 ECPayService 驗簽）

package service

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
	"littleduck-backend/internal/service/payment"
)

// mockDonationRepo 測試用 DonationRepository：以記憶體實作與 SQL 相同的條件更新語意
type mockDonationRepo struct {
	rows    map[string]*model.Donation
	created []*model.Donation

	createErrs []error // 依序回傳的 Create 錯誤，用完後成功
	findErr    error
	markErr    error

	markPaidCalls   int
	markFailedCalls int
	paidArgs        struct {
		paidAt       time.Time
		ecpayTradeNo string
		paymentType  string
	}
}

func newMockDonationRepo(rows ...model.Donation) *mockDonationRepo {
	m := &mockDonationRepo{rows: map[string]*model.Donation{}}
	for i := range rows {
		r := rows[i]
		m.rows[r.ID] = &r
	}
	return m
}

func (m *mockDonationRepo) FindAll() ([]model.Donation, error) { return nil, nil }
func (m *mockDonationRepo) Count() (int64, error)              { return int64(len(m.rows)), nil }
func (m *mockDonationRepo) Create(d *model.Donation) error {
	if len(m.createErrs) > 0 {
		err := m.createErrs[0]
		m.createErrs = m.createErrs[1:]
		if err != nil {
			return err
		}
	}
	cp := *d
	m.rows[d.ID] = &cp
	m.created = append(m.created, &cp)
	return nil
}
func (m *mockDonationRepo) FindByID(id string) (*model.Donation, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if d, ok := m.rows[id]; ok {
		cp := *d
		return &cp, nil
	}
	return nil, repository.ErrNotFound
}
func (m *mockDonationRepo) MarkPaid(id string, paidAt time.Time, ecpayTradeNo, paymentType string) (bool, error) {
	m.markPaidCalls++
	if m.markErr != nil {
		return false, m.markErr
	}
	d, ok := m.rows[id]
	if !ok || d.Status == model.DonationStatusPaid {
		return false, nil
	}
	d.Status, d.PaidAt, d.EcpayTradeNo, d.PaymentType = model.DonationStatusPaid, &paidAt, ecpayTradeNo, paymentType
	m.paidArgs.paidAt, m.paidArgs.ecpayTradeNo, m.paidArgs.paymentType = paidAt, ecpayTradeNo, paymentType
	return true, nil
}
func (m *mockDonationRepo) MarkFailed(id string) (bool, error) {
	m.markFailedCalls++
	if m.markErr != nil {
		return false, m.markErr
	}
	d, ok := m.rows[id]
	if !ok || d.Status != model.DonationStatusPending {
		return false, nil
	}
	d.Status = model.DonationStatusFailed
	return true, nil
}

// 測試用綠界設定（官方文件範例金鑰）
func paymentTestConfig() config.ECPayConfig {
	return config.ECPayConfig{
		MerchantID:    "3002607",
		HashKey:       "pwFHCqoQZGmho4w6",
		HashIV:        "EkRm7iFT261dpevs",
		Env:           "stage",
		ReturnURL:     "https://api.example.com/api/v1/payments/ecpay/callback",
		ClientBackURL: "https://www.example.com/sponsor/result",
	}
}

var payNow = time.Date(2026, 11, 17, 7, 30, 12, 0, time.UTC) // UTC+8：2026/11/17 15:30:12

func newTestPaymentService(repo *mockDonationRepo, features FeatureChecker, cfg config.ECPayConfig) *paymentService {
	svc := NewPaymentService(repo, features, payment.NewECPayService(cfg), cfg).(*paymentService)
	svc.now = func() time.Time { return payNow }
	return svc
}

// bothOn 兩個開關皆啟用
func bothOn() *stubFeatures {
	return &stubFeatures{states: map[string]bool{FeatureSponsor: true, FeatureSponsorPayment: true}}
}

func assertAppErr(t *testing.T, err error, status int, code string) {
	t.Helper()
	appErr, ok := apperror.AsAppError(err)
	if !ok {
		t.Fatalf("應為 AppError，得到 %v", err)
	}
	if appErr.Status != status || appErr.Code != code {
		t.Errorf("得到 %d %s，want %d %s", appErr.Status, appErr.Code, status, code)
	}
}

// ---- Create ----

func TestPaymentCreate_SuccessCreatesPendingDonation(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo()
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	view, err := svc.Create(PaymentInput{Amount: 100, Message: "加油"}, false)

	// Assert
	if err != nil {
		t.Fatalf("建立付款應成功: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("應新增 1 筆紀錄，得到 %d", len(repo.created))
	}
	d := repo.created[0]
	if d.ID != view.TradeNo || d.Status != model.DonationStatusPending || d.Amt != 100 ||
		d.Note != "加油" || d.Name != "匿名訪客" || d.Date != "2026 / 11 / 17" {
		t.Errorf("紀錄錯誤: %+v", d)
	}
	if view.ActionURL != "https://payment-stage.ecpay.com.tw/Cashier/AioCheckOut/V5" {
		t.Errorf("ActionURL = %q", view.ActionURL)
	}
	p := view.Params
	if p["TotalAmount"] != "100" || p["MerchantTradeNo"] != view.TradeNo || p["ChoosePayment"] != "Credit" ||
		p["ReturnURL"] != paymentTestConfig().ReturnURL || p["TradeDesc"] != "LittleDuck 贊助" ||
		p["ItemName"] != "請小鴨喝蜂蜜檸檬 NT$100" || p["MerchantTradeDate"] != "2026/11/17 15:30:12" {
		t.Errorf("Params 錯誤: %v", p)
	}
	if p["ClientBackURL"] != "https://www.example.com/sponsor/result?trade_no="+view.TradeNo {
		t.Errorf("ClientBackURL = %q", p["ClientBackURL"])
	}
	if !payment.VerifyCheckMac(p, paymentTestConfig().HashKey, paymentTestConfig().HashIV) {
		t.Error("CheckMacValue 應有效")
	}
	if strings.Contains(p["ItemName"], "加油") {
		t.Error("留言不應送綠界")
	}
}

func TestPaymentCreate_TradeNoFormat(t *testing.T) {
	// Arrange
	svc := newTestPaymentService(newMockDonationRepo(), bothOn(), paymentTestConfig())

	// Act
	view, err := svc.Create(PaymentInput{Amount: 100}, false)

	// Assert
	if err != nil {
		t.Fatalf("應成功: %v", err)
	}
	if len(view.TradeNo) != 20 || !regexp.MustCompile(`^LD261117153012[A-Z0-9]{6}$`).MatchString(view.TradeNo) {
		t.Errorf("交易編號格式錯誤（應為 LD+UTC+8 時間+6 碼英數）: %q", view.TradeNo)
	}
}

func TestPaymentCreate_TradeNosDiffer(t *testing.T) {
	// Arrange
	svc := newTestPaymentService(newMockDonationRepo(), bothOn(), paymentTestConfig())

	// Act
	a, _ := svc.Create(PaymentInput{Amount: 100}, false)
	b, _ := svc.Create(PaymentInput{Amount: 100}, false)

	// Assert
	if a.TradeNo == b.TradeNo {
		t.Errorf("同一秒內兩次建立的交易編號不應相同: %s", a.TradeNo)
	}
}

func TestPaymentCreate_ClientBackURLWithExistingQuery(t *testing.T) {
	// Arrange
	cfg := paymentTestConfig()
	cfg.ClientBackURL = "https://www.example.com/sponsor/result?from=ecpay"
	svc := newTestPaymentService(newMockDonationRepo(), bothOn(), cfg)

	// Act
	view, _ := svc.Create(PaymentInput{Amount: 100}, false)

	// Assert
	if want := "https://www.example.com/sponsor/result?from=ecpay&trade_no=" + view.TradeNo; view.Params["ClientBackURL"] != want {
		t.Errorf("ClientBackURL = %q, want %q", view.Params["ClientBackURL"], want)
	}
}

func TestPaymentCreate_FeatureGate(t *testing.T) {
	tests := []struct {
		name          string
		sponsor, pay  bool
		asAdmin       bool
		wantForbidden bool
	}{
		{"兩者皆啟用", true, true, false, false},
		{"sponsor 停用", false, true, false, true},
		{"sponsor_payment 停用", true, false, false, true},
		{"兩者皆停用", false, false, false, true},
		{"管理員預覽（sponsor_payment 停用）", true, false, true, false},
		{"管理員預覽（兩者皆停用）", false, false, true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			repo := newMockDonationRepo()
			feats := &stubFeatures{states: map[string]bool{FeatureSponsor: tc.sponsor, FeatureSponsorPayment: tc.pay}}
			svc := newTestPaymentService(repo, feats, paymentTestConfig())

			// Act
			_, err := svc.Create(PaymentInput{Amount: 100}, tc.asAdmin)

			// Assert
			if tc.wantForbidden {
				assertAppErr(t, err, http.StatusForbidden, apperror.CodeFeatureDisabled)
				if len(repo.created) != 0 {
					t.Error("開關停用不應建立紀錄")
				}
				return
			}
			if err != nil {
				t.Fatalf("應成功: %v", err)
			}
			if tc.asAdmin && len(feats.queried) != 0 {
				t.Errorf("管理員不應查詢開關，查了 %v", feats.queried)
			}
		})
	}
}

func TestPaymentCreate_FeatureReadErrorReturns500(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo()
	feats := &stubFeatures{states: map[string]bool{}, err: errors.New("db down")}
	svc := newTestPaymentService(repo, feats, paymentTestConfig())

	// Act
	_, err := svc.Create(PaymentInput{Amount: 100}, false)

	// Assert
	assertAppErr(t, err, http.StatusInternalServerError, apperror.CodeInternal)
	if len(repo.created) != 0 {
		t.Error("讀取開關失敗不應建立紀錄（fail-closed）")
	}
}

func TestPaymentCreate_UnconfiguredReturns503(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo()
	svc := newTestPaymentService(repo, bothOn(), config.ECPayConfig{})

	// Act
	_, err := svc.Create(PaymentInput{Amount: 100}, false)

	// Assert
	assertAppErr(t, err, http.StatusServiceUnavailable, apperror.CodePaymentUnavailable)
	if len(repo.created) != 0 {
		t.Error("設定不完整不應建立紀錄")
	}
}

func TestPaymentCreate_UnconfiguredAdminAlsoReturns503(t *testing.T) {
	// Arrange
	svc := newTestPaymentService(newMockDonationRepo(), bothOn(), config.ECPayConfig{})

	// Act
	_, err := svc.Create(PaymentInput{Amount: 100}, true)

	// Assert
	assertAppErr(t, err, http.StatusServiceUnavailable, apperror.CodePaymentUnavailable)
}

func TestPaymentCreate_AmountBoundaries(t *testing.T) {
	tests := []struct {
		amount int
		ok     bool
	}{
		{-1, false}, {0, false}, {9, false}, {10, true}, {11, true},
		{123, true}, {19999, true}, {20000, true}, {20001, false},
	}
	for _, tc := range tests {
		repo := newMockDonationRepo()
		svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

		_, err := svc.Create(PaymentInput{Amount: tc.amount}, false)

		if tc.ok && err != nil {
			t.Errorf("amount=%d 應成功: %v", tc.amount, err)
		}
		if !tc.ok {
			assertAppErr(t, err, http.StatusBadRequest, apperror.CodeBadRequest)
			if len(repo.created) != 0 {
				t.Errorf("amount=%d 不應建立紀錄", tc.amount)
			}
		}
	}
}

func TestPaymentCreate_TextLimitsCountRunes(t *testing.T) {
	tests := []struct {
		name string
		in   PaymentInput
		ok   bool
	}{
		{"message 100 個中文字", PaymentInput{Amount: 100, Message: strings.Repeat("鴨", 100)}, true},
		{"message 101 個中文字", PaymentInput{Amount: 100, Message: strings.Repeat("鴨", 101)}, false},
		{"name 50 個中文字", PaymentInput{Amount: 100, Name: strings.Repeat("鴨", 50)}, true},
		{"name 51 個中文字", PaymentInput{Amount: 100, Name: strings.Repeat("鴨", 51)}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			repo := newMockDonationRepo()
			svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

			// Act
			_, err := svc.Create(tc.in, false)

			// Assert
			if tc.ok && err != nil {
				t.Errorf("應成功: %v", err)
			}
			if !tc.ok {
				assertAppErr(t, err, http.StatusBadRequest, apperror.CodeBadRequest)
				if len(repo.created) != 0 {
					t.Error("不應建立紀錄")
				}
			}
		})
	}
}

func TestPaymentCreate_NameHandling(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", "匿名訪客"},
		{"   ", "匿名訪客"},
		{"小明", "小明"},
		{"  小明  ", "小明"},
	}
	for _, tc := range tests {
		repo := newMockDonationRepo()
		svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

		if _, err := svc.Create(PaymentInput{Amount: 100, Name: tc.in}, false); err != nil {
			t.Fatalf("應成功: %v", err)
		}

		if got := repo.created[0].Name; got != tc.want {
			t.Errorf("name %q → %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPaymentCreate_ValidationBeforeFeatureGate(t *testing.T) {
	// Arrange：開關停用且金額不合法
	feats := &stubFeatures{states: map[string]bool{}}
	svc := newTestPaymentService(newMockDonationRepo(), feats, paymentTestConfig())

	// Act
	_, err := svc.Create(PaymentInput{Amount: 1}, false)

	// Assert
	assertAppErr(t, err, http.StatusBadRequest, apperror.CodeBadRequest)
}

func TestPaymentCreate_RetriesOnceOnDuplicateTradeNo(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo()
	repo.createErrs = []error{repository.ErrDuplicate}
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	view, err := svc.Create(PaymentInput{Amount: 100}, false)

	// Assert
	if err != nil {
		t.Fatalf("碰撞後重產一次應成功: %v", err)
	}
	if len(repo.created) != 1 || repo.created[0].ID != view.TradeNo {
		t.Errorf("應只留下 1 筆且 id 等於回傳 trade_no: %+v", repo.created)
	}
}

func TestPaymentCreate_DuplicateTwiceReturns500(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo()
	repo.createErrs = []error{repository.ErrDuplicate, repository.ErrDuplicate}
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	_, err := svc.Create(PaymentInput{Amount: 100}, false)

	// Assert
	assertAppErr(t, err, http.StatusInternalServerError, apperror.CodeInternal)
}

func TestPaymentCreate_RepoErrorReturns500(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo()
	repo.createErrs = []error{errors.New("db down")}
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	_, err := svc.Create(PaymentInput{Amount: 100}, false)

	// Assert
	assertAppErr(t, err, http.StatusInternalServerError, apperror.CodeInternal)
}

func TestRandomString_UsesAlphabetAndRejectsBiasedBytes(t *testing.T) {
	// Arrange：前 4 個位元組 >= 252 會被拒絕；其後 0,1,35,36 對應 A,B,9,A
	src := strings.NewReader("\xfc\xfd\xfe\xff" + "\x00\x01\x23\x24" + strings.Repeat("\x00", 8))

	// Act
	got, err := randomString(src, 4)

	// Assert
	if err != nil {
		t.Fatalf("不應失敗: %v", err)
	}
	if got != "AB9A" {
		t.Errorf("得到 %q, want %q", got, "AB9A")
	}
}

func TestRandomString_ReaderErrorPropagates(t *testing.T) {
	// Arrange & Act
	_, err := randomString(strings.NewReader(""), 6)

	// Assert
	if err == nil {
		t.Error("亂數來源失敗應回傳錯誤")
	}
}

// ---- HandleECPayCallback ----

func pendingDonation() model.Donation {
	return model.Donation{ID: "LD261117153012A9K2QZ", Amt: 100, Status: model.DonationStatusPending}
}

// signedParams 產生以測試金鑰簽章的通知參數
func signedParams(mutate func(p map[string]string)) map[string]string {
	cfg := paymentTestConfig()
	p := map[string]string{
		"MerchantID":      cfg.MerchantID,
		"MerchantTradeNo": "LD261117153012A9K2QZ",
		"RtnCode":         "1",
		"RtnMsg":          "交易成功",
		"TradeNo":         "2611171530120001",
		"TradeAmt":        "100",
		"PaymentDate":     "2026/11/17 15:31:00",
		"PaymentType":     "Credit_CreditCard",
		"SimulatePaid":    "0",
	}
	if mutate != nil {
		mutate(p)
	}
	p["CheckMacValue"] = payment.CheckMacValue(p, cfg.HashKey, cfg.HashIV)
	return p
}

func TestCallback_PaidNotificationMarksPaid(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo(pendingDonation())
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	got := svc.HandleECPayCallback(signedParams(nil))

	// Assert
	if got != "1|OK" {
		t.Fatalf("回應 = %q, want 1|OK", got)
	}
	d := repo.rows["LD261117153012A9K2QZ"]
	if d.Status != model.DonationStatusPaid || d.EcpayTradeNo != "2611171530120001" || d.PaymentType != "Credit_CreditCard" {
		t.Errorf("紀錄錯誤: %+v", d)
	}
	wantPaidAt := time.Date(2026, 11, 17, 7, 31, 0, 0, time.UTC)
	if d.PaidAt == nil || !d.PaidAt.Equal(wantPaidAt) {
		t.Errorf("paid_at 應為 PaymentDate（UTC+8）= %v，得到 %v", wantPaidAt, d.PaidAt)
	}
}

func TestCallback_InvalidPaymentDateFallsBackToNow(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo(pendingDonation())
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	got := svc.HandleECPayCallback(signedParams(func(p map[string]string) { p["PaymentDate"] = "bad" }))

	// Assert
	if got != "1|OK" || repo.rows["LD261117153012A9K2QZ"].PaidAt == nil || !repo.rows["LD261117153012A9K2QZ"].PaidAt.Equal(payNow) {
		t.Errorf("PaymentDate 無法解析時 paid_at 應為目前時間，回應 %q，紀錄 %+v", got, repo.rows["LD261117153012A9K2QZ"])
	}
}

func TestCallback_DuplicateNotificationKeepsPaidAt(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo(pendingDonation())
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())
	svc.HandleECPayCallback(signedParams(nil))
	firstPaidAt := *repo.rows["LD261117153012A9K2QZ"].PaidAt

	// Act：同一筆成功通知再送一次（PaymentDate 不同也不應改動）
	got := svc.HandleECPayCallback(signedParams(func(p map[string]string) { p["PaymentDate"] = "2026/11/18 10:00:00" }))

	// Assert
	d := repo.rows["LD261117153012A9K2QZ"]
	if got != "1|OK" || d.Status != model.DonationStatusPaid || !d.PaidAt.Equal(firstPaidAt) {
		t.Errorf("重複通知應回 1|OK 且 paid_at 不變: %q %+v", got, d)
	}
	if repo.markPaidCalls != 2 {
		t.Errorf("重複通知仍應走條件更新，MarkPaid 呼叫 %d 次", repo.markPaidCalls)
	}
}

func TestCallback_CheckMacError(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo(pendingDonation())
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())
	p := signedParams(nil)
	p["CheckMacValue"] = strings.Repeat("A", 64)

	// Act
	got := svc.HandleECPayCallback(p)

	// Assert
	if got != "0|CheckMacValue Error" {
		t.Errorf("回應 = %q", got)
	}
	if repo.rows["LD261117153012A9K2QZ"].Status != model.DonationStatusPending || repo.markPaidCalls+repo.markFailedCalls != 0 {
		t.Error("檢查碼錯誤不得改變狀態")
	}
}

func TestCallback_MerchantIDError(t *testing.T) {
	// Arrange：檢查碼正確但 MerchantID 不同
	repo := newMockDonationRepo(pendingDonation())
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	got := svc.HandleECPayCallback(signedParams(func(p map[string]string) { p["MerchantID"] = "9999999" }))

	// Assert
	if got != "0|MerchantID Error" {
		t.Errorf("回應 = %q", got)
	}
	if repo.markPaidCalls+repo.markFailedCalls != 0 {
		t.Error("商店代號不符不得改變狀態")
	}
}

func TestCallback_MacCheckedBeforeMerchant(t *testing.T) {
	// Arrange：檢查碼與商店代號都錯
	svc := newTestPaymentService(newMockDonationRepo(pendingDonation()), bothOn(), paymentTestConfig())
	p := signedParams(func(p map[string]string) { p["MerchantID"] = "9999999" })
	p["CheckMacValue"] = "BAD"

	// Act & Assert
	if got := svc.HandleECPayCallback(p); got != "0|CheckMacValue Error" {
		t.Errorf("應先回檢查碼錯誤，得到 %q", got)
	}
}

func TestCallback_OrderNotFound(t *testing.T) {
	// Arrange
	svc := newTestPaymentService(newMockDonationRepo(), bothOn(), paymentTestConfig())

	// Act & Assert
	if got := svc.HandleECPayCallback(signedParams(nil)); got != "0|Order Not Found" {
		t.Errorf("回應 = %q", got)
	}
}

func TestCallback_FindErrorReturnsInternalError(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo(pendingDonation())
	repo.findErr = errors.New("db down")
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act & Assert
	if got := svc.HandleECPayCallback(signedParams(nil)); got != "0|Internal Error" {
		t.Errorf("回應 = %q", got)
	}
}

func TestCallback_AmountMismatch(t *testing.T) {
	tests := []struct{ name, amt string }{
		{"金額不同", "101"},
		{"金額為 0", "0"},
		{"非數字", "abc"},
		{"空字串", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			repo := newMockDonationRepo(pendingDonation())
			svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

			// Act
			got := svc.HandleECPayCallback(signedParams(func(p map[string]string) { p["TradeAmt"] = tc.amt }))

			// Assert
			if got != "0|Amount Mismatch" {
				t.Errorf("回應 = %q", got)
			}
			if repo.rows["LD261117153012A9K2QZ"].Status != model.DonationStatusPending {
				t.Error("金額不符不得改變狀態")
			}
		})
	}
}

func TestCallback_AmountCheckedBeforeSimulatePaid(t *testing.T) {
	// Arrange：模擬付款但金額不符
	svc := newTestPaymentService(newMockDonationRepo(pendingDonation()), bothOn(), paymentTestConfig())

	// Act
	got := svc.HandleECPayCallback(signedParams(func(p map[string]string) { p["SimulatePaid"] = "1"; p["TradeAmt"] = "5" }))

	// Assert
	if got != "0|Amount Mismatch" {
		t.Errorf("回應 = %q", got)
	}
}

func TestCallback_SimulatePaidDoesNotChangeState(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo(pendingDonation())
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	got := svc.HandleECPayCallback(signedParams(func(p map[string]string) { p["SimulatePaid"] = "1" }))

	// Assert
	if got != "1|OK" {
		t.Errorf("回應 = %q", got)
	}
	if repo.rows["LD261117153012A9K2QZ"].Status != model.DonationStatusPending || repo.markPaidCalls+repo.markFailedCalls != 0 {
		t.Error("SimulatePaid=1 不得改變狀態")
	}
}

func TestCallback_FailedNotificationMarksPendingFailed(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo(pendingDonation())
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	got := svc.HandleECPayCallback(signedParams(func(p map[string]string) { p["RtnCode"] = "10100058" }))

	// Assert
	if got != "1|OK" {
		t.Errorf("回應 = %q", got)
	}
	if repo.rows["LD261117153012A9K2QZ"].Status != model.DonationStatusFailed || repo.markPaidCalls != 0 {
		t.Errorf("應轉 failed 且不呼叫 MarkPaid: %+v", repo.rows["LD261117153012A9K2QZ"])
	}
}

func TestCallback_PaidIsNotOverwrittenByFailedNotification(t *testing.T) {
	// Arrange
	d := pendingDonation()
	d.Status = model.DonationStatusPaid
	repo := newMockDonationRepo(d)
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	got := svc.HandleECPayCallback(signedParams(func(p map[string]string) { p["RtnCode"] = "0" }))

	// Assert
	if got != "1|OK" || repo.rows["LD261117153012A9K2QZ"].Status != model.DonationStatusPaid {
		t.Errorf("paid 不應被改回: %q %+v", got, repo.rows["LD261117153012A9K2QZ"])
	}
}

func TestCallback_FailedCanLaterBecomePaid(t *testing.T) {
	// Arrange
	d := pendingDonation()
	d.Status = model.DonationStatusFailed
	repo := newMockDonationRepo(d)
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	got := svc.HandleECPayCallback(signedParams(nil))

	// Assert
	if got != "1|OK" || repo.rows["LD261117153012A9K2QZ"].Status != model.DonationStatusPaid {
		t.Errorf("failed 後續成功通知應轉 paid: %q %+v", got, repo.rows["LD261117153012A9K2QZ"])
	}
}

func TestCallback_UpdateErrorReturnsInternalError(t *testing.T) {
	for _, rtn := range []string{"1", "0"} {
		// Arrange
		repo := newMockDonationRepo(pendingDonation())
		repo.markErr = errors.New("db down")
		svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

		// Act
		got := svc.HandleECPayCallback(signedParams(func(p map[string]string) { p["RtnCode"] = rtn }))

		// Assert
		if got != "0|Internal Error" {
			t.Errorf("RtnCode=%s 更新失敗回應 = %q, want 0|Internal Error", rtn, got)
		}
	}
}

func TestCallback_IgnoresFeatureFlags(t *testing.T) {
	// Arrange：兩個開關皆停用，且讀取會失敗
	repo := newMockDonationRepo(pendingDonation())
	feats := &stubFeatures{states: map[string]bool{}, err: errors.New("db down")}
	svc := newTestPaymentService(repo, feats, paymentTestConfig())

	// Act
	got := svc.HandleECPayCallback(signedParams(nil))

	// Assert
	if got != "1|OK" || repo.rows["LD261117153012A9K2QZ"].Status != model.DonationStatusPaid {
		t.Errorf("開關不應影響通知處理: %q", got)
	}
	if len(feats.queried) != 0 {
		t.Errorf("通知處理不應查詢開關: %v", feats.queried)
	}
}

func TestCallback_UnconfiguredRejectsEvenWithEmptyKeySignature(t *testing.T) {
	// Arrange：設定為空，攻擊者以空金鑰簽章
	repo := newMockDonationRepo(pendingDonation())
	svc := newTestPaymentService(repo, bothOn(), config.ECPayConfig{})
	p := map[string]string{"MerchantID": "", "MerchantTradeNo": "LD261117153012A9K2QZ", "RtnCode": "1", "TradeAmt": "100"}
	p["CheckMacValue"] = payment.CheckMacValue(p, "", "")

	// Act
	got := svc.HandleECPayCallback(p)

	// Assert
	if got != "0|CheckMacValue Error" || repo.markPaidCalls != 0 {
		t.Errorf("設定不完整應拒絕: %q", got)
	}
}

// ---- Status ----

func TestPaymentStatus_ReturnsOnlyPublicFields(t *testing.T) {
	// Arrange
	d := pendingDonation()
	d.Name, d.Note, d.EcpayTradeNo = "小明", "私密留言", "T1"
	svc := newTestPaymentService(newMockDonationRepo(d), bothOn(), paymentTestConfig())

	// Act
	v, err := svc.Status("LD261117153012A9K2QZ")

	// Assert
	if err != nil {
		t.Fatalf("應成功: %v", err)
	}
	if *v != (PaymentStatusView{TradeNo: "LD261117153012A9K2QZ", Amount: 100, Status: "pending"}) {
		t.Errorf("得到 %+v", *v)
	}
}

func TestPaymentStatus_NotFoundReturns404(t *testing.T) {
	// Arrange
	svc := newTestPaymentService(newMockDonationRepo(), bothOn(), paymentTestConfig())

	// Act
	_, err := svc.Status("NOPE")

	// Assert
	assertAppErr(t, err, http.StatusNotFound, apperror.CodeNotFound)
}

func TestPaymentStatus_RepoErrorReturns500(t *testing.T) {
	// Arrange
	repo := newMockDonationRepo()
	repo.findErr = errors.New("db down")
	svc := newTestPaymentService(repo, bothOn(), paymentTestConfig())

	// Act
	_, err := svc.Status("X")

	// Assert
	assertAppErr(t, err, http.StatusInternalServerError, apperror.CodeInternal)
}
