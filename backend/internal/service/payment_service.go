// backend/internal/service/payment_service.go
// Payment service：贊助付款業務邏輯——建立付款（開關、驗證、交易編號、pending 紀錄）、
// 處理綠界付款結果通知（冪等狀態轉換）、查詢付款狀態

package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
	"littleduck-backend/internal/service/payment"
)

// 贊助付款的輸入限制
const (
	PaymentAmountMin  = 10    // 金額下限（新台幣元）
	PaymentAmountMax  = 20000 // 金額上限（新台幣元）
	PaymentMessageMax = 100   // 留言上限（字）
	PaymentNameMax    = 50    // 贊助者名稱上限（字）
)

// 贊助付款的固定文字
const (
	anonymousDonorName = "匿名訪客"
	paymentTradeDesc   = "LittleDuck 贊助"
)

// 綠界 callback 的回應字串（綠界要求純文字；非 1|OK 時綠界會重送）
const (
	callbackOK            = "1|OK"
	callbackMacError      = "0|CheckMacValue Error"
	callbackMerchantError = "0|MerchantID Error"
	callbackNotFound      = "0|Order Not Found"
	callbackAmountError   = "0|Amount Mismatch"
	callbackInternalError = "0|Internal Error"
)

// 交易編號：LD + yyMMddHHmmss + 6 碼英數，共 20 碼（綠界上限）
const (
	tradeNoPrefix      = "LD"
	tradeNoRandomLen   = 6
	tradeNoAlphabet    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tradeNoMaxAttempts = 2 // 主鍵碰撞時重產一次
)

// PaymentInput 建立贊助付款的輸入
type PaymentInput struct {
	Amount  int    `json:"amount"`  // 新台幣元，10～20000
	Message string `json:"message"` // 想對小鴨說的話，存入 note，不送綠界
	Name    string `json:"name"`    // 贊助者名稱；空白記為「匿名訪客」
}

// CreatePaymentView 建立付款的回應：前端以表單 POST 到 ActionURL
type CreatePaymentView struct {
	TradeNo   string            `json:"trade_no"`   // 綠界 MerchantTradeNo，同時是贊助紀錄 id
	ActionURL string            `json:"action_url"` // 綠界 AIO 付款網址
	Params    map[string]string `json:"params"`     // 表單欄位（含 CheckMacValue）
}

// PaymentStatusView 公開的付款狀態：不含名稱、留言與綠界交易細節
type PaymentStatusView struct {
	TradeNo string `json:"trade_no"`
	Amount  int    `json:"amount"`
	Status  string `json:"status"`
}

// PaymentService 贊助付款業務邏輯介面
type PaymentService interface {
	// Create 建立 pending 贊助紀錄並回傳綠界表單；asAdmin 為 true（管理員預覽）時略過功能開關檢查
	Create(in PaymentInput, asAdmin bool) (*CreatePaymentView, error)
	// HandleECPayCallback 處理綠界付款結果通知，回傳要寫給綠界的純文字回應（1|OK 或 0|原因）
	HandleECPayCallback(params map[string]string) string
	// Status 以交易編號查詢付款狀態；查無回 404
	Status(tradeNo string) (*PaymentStatusView, error)
}

// paymentService PaymentService 實作
type paymentService struct {
	repo          repository.DonationRepository
	features      FeatureChecker
	provider      payment.PaymentProvider
	clientBackURL string
	now           func() time.Time
	random        io.Reader
}

// NewPaymentService 建立 PaymentService
func NewPaymentService(
	repo repository.DonationRepository,
	features FeatureChecker,
	provider payment.PaymentProvider,
	cfg config.ECPayConfig,
) PaymentService {
	return &paymentService{
		repo:          repo,
		features:      features,
		provider:      provider,
		clientBackURL: cfg.ClientBackURL,
		now:           time.Now,
		random:        rand.Reader,
	}
}

// Create 依序檢查：輸入（400）→ 功能開關（非管理員；403，讀取失敗 500）→ 金流設定（503）。
func (s *paymentService) Create(in PaymentInput, asAdmin bool) (*CreatePaymentView, error) {
	name, err := validatePaymentInput(in)
	if err != nil {
		return nil, err
	}
	if !asAdmin {
		if err := s.ensureEnabled(); err != nil {
			return nil, err
		}
	}
	if !s.provider.Configured() {
		return nil, apperror.PaymentUnavailable("付款功能暫時無法使用")
	}

	for attempt := 0; attempt < tradeNoMaxAttempts; attempt++ {
		now := s.now()
		tradeNo, err := s.newTradeNo(now)
		if err != nil {
			return nil, apperror.Internal(err, "建立付款失敗")
		}
		res, err := s.provider.CreatePayment(payment.CreatePaymentInput{
			OrderID:       tradeNo,
			Amount:        in.Amount,
			Description:   paymentTradeDesc,
			ItemName:      fmt.Sprintf("請小鴨喝蜂蜜檸檬 NT$%d", in.Amount),
			TradeDate:     now,
			ClientBackURL: appendTradeNo(s.clientBackURL, tradeNo),
			ChoosePayment: payment.ChoosePaymentCredit,
		})
		if err != nil {
			if _, ok := apperror.AsAppError(err); ok {
				return nil, err
			}
			return nil, apperror.Internal(err, "建立付款失敗")
		}
		err = s.repo.Create(&model.Donation{
			ID:     tradeNo,
			Date:   now.In(taipeiZone).Format("2006 / 01 / 02"),
			Amt:    in.Amount,
			Name:   name,
			Note:   in.Message,
			Status: model.DonationStatusPending,
		})
		if errors.Is(err, repository.ErrDuplicate) {
			continue
		}
		if err != nil {
			return nil, apperror.Internal(err, "建立付款失敗")
		}
		return &CreatePaymentView{TradeNo: tradeNo, ActionURL: res.ActionURL, Params: res.Params}, nil
	}
	return nil, apperror.Internal(errors.New("交易編號重複"), "建立付款失敗")
}

// validatePaymentInput 驗證金額與文字長度（以字元數計），回傳整理後的贊助者名稱
func validatePaymentInput(in PaymentInput) (string, error) {
	if in.Amount < PaymentAmountMin || in.Amount > PaymentAmountMax {
		return "", apperror.BadRequest(fmt.Sprintf("贊助金額需介於 %d～%d 元", PaymentAmountMin, PaymentAmountMax))
	}
	if utf8.RuneCountInString(in.Message) > PaymentMessageMax {
		return "", apperror.BadRequest(fmt.Sprintf("留言不可超過 %d 字", PaymentMessageMax))
	}
	name := strings.TrimSpace(in.Name)
	if utf8.RuneCountInString(name) > PaymentNameMax {
		return "", apperror.BadRequest(fmt.Sprintf("名稱不可超過 %d 字", PaymentNameMax))
	}
	if name == "" {
		name = anonymousDonorName
	}
	return name, nil
}

// ensureEnabled 檢查 sponsor 與 sponsor_payment 皆啟用；任一停用回 403，讀取失敗回 500（fail-closed）
func (s *paymentService) ensureEnabled() error {
	for _, f := range []struct{ key, label string }{
		{FeatureSponsor, "贊助"},
		{FeatureSponsorPayment, "贊助付款"},
	} {
		enabled, err := s.features.IsEnabled(f.key)
		if err != nil {
			if _, ok := apperror.AsAppError(err); ok {
				return err
			}
			return apperror.Internal(err, "取得功能開關失敗")
		}
		if !enabled {
			return apperror.FeatureDisabled(f.label + "功能目前已停用")
		}
	}
	return nil
}

// newTradeNo 產生 LD + yyMMddHHmmss（UTC+8）+ 6 碼 crypto/rand 英數
func (s *paymentService) newTradeNo(now time.Time) (string, error) {
	suffix, err := randomString(s.random, tradeNoRandomLen)
	if err != nil {
		return "", err
	}
	return tradeNoPrefix + now.In(taipeiZone).Format("060102150405") + suffix, nil
}

// randomString 以拒絕取樣從 tradeNoAlphabet 取 n 碼，避免取餘數造成的分布偏差
func randomString(r io.Reader, n int) (string, error) {
	const limit = 256 - 256%len(tradeNoAlphabet) // 252
	out := make([]byte, 0, n)
	buf := make([]byte, n*2)
	for len(out) < n {
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if int(b) >= limit {
				continue
			}
			out = append(out, tradeNoAlphabet[int(b)%len(tradeNoAlphabet)])
			if len(out) == n {
				break
			}
		}
	}
	return string(out), nil
}

// appendTradeNo 在返回網址加上 trade_no 查詢參數（已有查詢字串時用 &）
func appendTradeNo(base, tradeNo string) string {
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + "trade_no=" + tradeNo
}

// HandleECPayCallback 依序：驗證 CheckMacValue → MerchantID → 查交易 → 金額 → SimulatePaid → RtnCode。
// 驗證通過後只做條件更新，影響列數 0（重複通知、paid 不被改回）仍回 1|OK。
func (s *paymentService) HandleECPayCallback(params map[string]string) string {
	if !s.provider.Configured() {
		// 設定不完整時無法驗證，一律拒絕（避免以空金鑰通過檢查）
		return callbackMacError
	}
	if err := s.provider.VerifyCallback(params); err != nil {
		switch {
		case errors.Is(err, payment.ErrCheckMacMismatch):
			return callbackMacError
		case errors.Is(err, payment.ErrMerchantMismatch):
			return callbackMerchantError
		default:
			return callbackInternalError
		}
	}

	d, err := s.repo.FindByID(params["MerchantTradeNo"])
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return callbackNotFound
		}
		return callbackInternalError
	}

	amount, err := strconv.Atoi(params["TradeAmt"])
	if err != nil || amount != d.Amt {
		return callbackAmountError
	}

	if params["SimulatePaid"] == "1" {
		return callbackOK
	}

	if params["RtnCode"] == "1" {
		if _, err := s.repo.MarkPaid(d.ID, s.paymentTime(params["PaymentDate"]), params["TradeNo"], params["PaymentType"]); err != nil {
			return callbackInternalError
		}
		return callbackOK
	}
	if _, err := s.repo.MarkFailed(d.ID); err != nil {
		return callbackInternalError
	}
	return callbackOK
}

// paymentTime 解析綠界 PaymentDate（yyyy/MM/dd HH:mm:ss，UTC+8）；格式不符時用目前時間
func (s *paymentService) paymentTime(raw string) time.Time {
	if t, err := time.ParseInLocation("2006/01/02 15:04:05", raw, taipeiZone); err == nil {
		return t
	}
	return s.now()
}

// Status 查詢付款狀態
func (s *paymentService) Status(tradeNo string) (*PaymentStatusView, error) {
	d, err := s.repo.FindByID(tradeNo)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.NotFound("付款紀錄不存在")
		}
		return nil, apperror.Internal(err, "取得付款狀態失敗")
	}
	return &PaymentStatusView{TradeNo: d.ID, Amount: d.Amt, Status: d.Status}, nil
}
