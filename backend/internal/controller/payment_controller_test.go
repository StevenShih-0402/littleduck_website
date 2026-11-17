// backend/internal/controller/payment_controller_test.go
// payment controller 單元測試：建立付款（成功、400、403、503、管理員旗標）、
// callback 回純文字 200、狀態查詢 404（mock service）

package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"littleduck-backend/internal/middleware"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/service"
)

// mockPaymentService 測試用 PaymentService
type mockPaymentService struct {
	createView *service.CreatePaymentView
	createErr  error
	gotInput   service.PaymentInput
	gotAdmin   *bool

	callbackResult string
	gotParams      map[string]string

	statusView *service.PaymentStatusView
	statusErr  error
	gotTradeNo string
}

func (m *mockPaymentService) Create(in service.PaymentInput, asAdmin bool) (*service.CreatePaymentView, error) {
	m.gotInput, m.gotAdmin = in, &asAdmin
	return m.createView, m.createErr
}
func (m *mockPaymentService) HandleECPayCallback(p map[string]string) string {
	m.gotParams = p
	return m.callbackResult
}
func (m *mockPaymentService) Status(tradeNo string) (*service.PaymentStatusView, error) {
	m.gotTradeNo = tradeNo
	return m.statusView, m.statusErr
}

func newPaymentRouter(svc service.PaymentService, adminSubject *string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	if adminSubject != nil {
		r.Use(func(c *gin.Context) { c.Set(middleware.ContextKeyAdmin, *adminSubject) })
	}
	ctrl := NewPaymentController(svc)
	r.POST("/api/v1/payments", ctrl.Create)
	r.POST("/api/v1/payments/ecpay/callback", ctrl.ECPayCallback)
	r.GET("/api/v1/payments/:trade_no", ctrl.Status)
	return r
}

func postJSON(r http.Handler, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("回應不是合法 JSON: %v（%s）", err, rec.Body.String())
	}
	code, _ := body["code"].(string)
	return code
}

func TestPaymentCreate_Returns200WithFormFields(t *testing.T) {
	// Arrange
	svc := &mockPaymentService{createView: &service.CreatePaymentView{
		TradeNo: "LD1", ActionURL: "https://payment-stage.ecpay.com.tw/Cashier/AioCheckOut/V5",
		Params: map[string]string{"TotalAmount": "100"},
	}}
	r := newPaymentRouter(svc, nil)

	// Act
	rec := postJSON(r, "/api/v1/payments", `{"amount":100,"message":"加油","name":"小明"}`)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Code string `json:"code"`
		Data struct {
			TradeNo   string            `json:"trade_no"`
			ActionURL string            `json:"action_url"`
			Params    map[string]string `json:"params"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Code != "SUCCESS" || body.Data.TradeNo != "LD1" || body.Data.Params["TotalAmount"] != "100" ||
		!strings.HasPrefix(body.Data.ActionURL, "https://payment-stage") {
		t.Errorf("回應錯誤: %s", rec.Body.String())
	}
	if svc.gotInput.Amount != 100 || svc.gotInput.Message != "加油" || svc.gotInput.Name != "小明" {
		t.Errorf("輸入未正確傳給 service: %+v", svc.gotInput)
	}
	if svc.gotAdmin == nil || *svc.gotAdmin {
		t.Error("未帶管理員權杖時 asAdmin 應為 false")
	}
}

func TestPaymentCreate_AdminContextPassesAsAdmin(t *testing.T) {
	// Arrange
	svc := &mockPaymentService{createView: &service.CreatePaymentView{}}
	admin := "admin"
	r := newPaymentRouter(svc, &admin)

	// Act
	postJSON(r, "/api/v1/payments", `{"amount":100}`)

	// Assert
	if svc.gotAdmin == nil || !*svc.gotAdmin {
		t.Error("有管理員 context 時 asAdmin 應為 true")
	}
}

func TestPaymentCreate_ErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		svcErr   error
		wantCode int
		wantErr  string
	}{
		{"JSON 格式錯誤", `not json`, nil, http.StatusBadRequest, apperror.CodeBadRequest},
		{"amount 非整數", `{"amount":1.5}`, nil, http.StatusBadRequest, apperror.CodeBadRequest},
		{"amount 為字串", `{"amount":"abc"}`, nil, http.StatusBadRequest, apperror.CodeBadRequest},
		{"service 回 400", `{"amount":1}`, apperror.BadRequest("金額"), http.StatusBadRequest, apperror.CodeBadRequest},
		{"功能停用", `{"amount":100}`, apperror.FeatureDisabled("停用"), http.StatusForbidden, apperror.CodeFeatureDisabled},
		{"設定不完整", `{"amount":100}`, apperror.PaymentUnavailable("暫時無法使用"), http.StatusServiceUnavailable, apperror.CodePaymentUnavailable},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			svc := &mockPaymentService{createErr: tc.svcErr}
			r := newPaymentRouter(svc, nil)

			// Act
			rec := postJSON(r, "/api/v1/payments", tc.body)

			// Assert
			if rec.Code != tc.wantCode || errorCode(t, rec) != tc.wantErr {
				t.Errorf("得到 %d %s，want %d %s", rec.Code, rec.Body.String(), tc.wantCode, tc.wantErr)
			}
		})
	}
}

func postForm(r http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(rec, req)
	return rec
}

func TestECPayCallback_ReturnsPlainText200(t *testing.T) {
	tests := []string{"1|OK", "0|CheckMacValue Error", "0|Internal Error"}
	for _, want := range tests {
		// Arrange
		svc := &mockPaymentService{callbackResult: want}
		r := newPaymentRouter(svc, nil)

		// Act
		rec := postForm(r, "/api/v1/payments/ecpay/callback", url.Values{
			"MerchantTradeNo": {"LD1"}, "RtnCode": {"1"}, "CheckMacValue": {"ABC"},
		})

		// Assert：一律 200 純文字，不走 JSON envelope
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
			t.Errorf("Content-Type = %q, want text/plain", ct)
		}
		if rec.Body.String() != want {
			t.Errorf("body = %q, want %q", rec.Body.String(), want)
		}
		if svc.gotParams["MerchantTradeNo"] != "LD1" || svc.gotParams["RtnCode"] != "1" || svc.gotParams["CheckMacValue"] != "ABC" {
			t.Errorf("表單欄位未正確轉成 map: %v", svc.gotParams)
		}
	}
}

func TestECPayCallback_IgnoresQueryString(t *testing.T) {
	// Arrange：查詢字串的參數不應被當成通知欄位
	svc := &mockPaymentService{callbackResult: "1|OK"}
	r := newPaymentRouter(svc, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/ecpay/callback?RtnCode=1&Evil=1", strings.NewReader("MerchantTradeNo=LD1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Act
	r.ServeHTTP(rec, req)

	// Assert
	if _, ok := svc.gotParams["Evil"]; ok || svc.gotParams["MerchantTradeNo"] != "LD1" {
		t.Errorf("只應採用表單本文: %v", svc.gotParams)
	}
}

func TestPaymentStatus_ReturnsView(t *testing.T) {
	// Arrange
	svc := &mockPaymentService{statusView: &service.PaymentStatusView{TradeNo: "LD1", Amount: 50, Status: "pending"}}
	r := newPaymentRouter(svc, nil)

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/payments/LD1", nil))

	// Assert
	if rec.Code != http.StatusOK || svc.gotTradeNo != "LD1" {
		t.Fatalf("status = %d，trade_no = %q", rec.Code, svc.gotTradeNo)
	}
	var body struct {
		Data map[string]any `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Data["status"] != "pending" || body.Data["amount"] != float64(50) || body.Data["trade_no"] != "LD1" || len(body.Data) != 3 {
		t.Errorf("回應錯誤: %s", rec.Body.String())
	}
}

func TestPaymentStatus_NotFoundReturns404(t *testing.T) {
	// Arrange
	svc := &mockPaymentService{statusErr: apperror.NotFound("付款紀錄不存在")}
	r := newPaymentRouter(svc, nil)

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/payments/NOPE", nil))

	// Assert
	if rec.Code != http.StatusNotFound || errorCode(t, rec) != apperror.CodeNotFound {
		t.Errorf("得到 %d %s", rec.Code, rec.Body.String())
	}
}
