// backend/internal/middleware/middleware_test.go
// middleware 套件單元測試：驗證 request logger、recovery、error handler 與 CORS 行為

package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"littleduck-backend/internal/pkg/apperror"
)

// newObservedLogger 建立可檢視輸出紀錄的測試 logger
func newObservedLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

// perform 建立掛好 middleware 的 gin engine 並發送請求
func perform(handlers []gin.HandlerFunc, register func(*gin.Engine), method, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(handlers...)
	register(r)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	r.ServeHTTP(rec, req)
	return rec
}

func TestRequestLogger_LogsMethodPathStatusLatency(t *testing.T) {
	// Arrange
	log, logs := newObservedLogger()
	register := func(r *gin.Engine) {
		r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	}

	// Act
	perform([]gin.HandlerFunc{RequestLogger(log)}, register, http.MethodGet, "/ping")

	// Assert
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("應輸出 1 筆請求日誌，得到 %d 筆", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["method"] != http.MethodGet || fields["path"] != "/ping" || fields["status"] != int64(http.StatusOK) {
		t.Errorf("日誌欄位錯誤: %+v", fields)
	}
	if _, ok := fields["latency"]; !ok {
		t.Error("日誌應包含 latency 欄位")
	}
}

func TestRecovery_PanicReturns500AndLogsStack(t *testing.T) {
	// Arrange
	log, logs := newObservedLogger()
	register := func(r *gin.Engine) {
		r.GET("/boom", func(c *gin.Context) { panic("kaboom") })
		r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })
	}
	handlers := []gin.HandlerFunc{Recovery(log)}

	// Act
	recBoom := perform(handlers, register, http.MethodGet, "/boom")
	recOK := perform(handlers, register, http.MethodGet, "/ok")

	// Assert
	if recBoom.Code != http.StatusInternalServerError {
		t.Errorf("panic 請求 status = %d, want 500", recBoom.Code)
	}
	if strings.Contains(recBoom.Body.String(), "kaboom") {
		t.Error("回應不應洩漏 panic 內容")
	}
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("應輸出 1 筆 panic 日誌，得到 %d 筆", len(entries))
	}
	if _, ok := entries[0].ContextMap()["stack"]; !ok {
		t.Error("panic 日誌應包含 stack 欄位")
	}
	if recOK.Code != http.StatusOK {
		t.Errorf("panic 後續請求 status = %d, want 200", recOK.Code)
	}
}

func TestErrorHandler_AppErrorMapsToHTTPResponse(t *testing.T) {
	// Arrange
	log, _ := newObservedLogger()
	register := func(r *gin.Engine) {
		r.GET("/missing", func(c *gin.Context) {
			_ = c.Error(apperror.NotFound("文章不存在"))
		})
	}

	// Act
	rec := perform([]gin.HandlerFunc{ErrorHandler(log)}, register, http.MethodGet, "/missing")

	// Assert
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("回應不是合法 JSON: %v", err)
	}
	if body["code"] != apperror.CodeNotFound || body["message"] != "文章不存在" {
		t.Errorf("envelope 錯誤: %+v", body)
	}
}

func TestErrorHandler_UnknownErrorReturns500WithoutDetails(t *testing.T) {
	// Arrange
	log, logs := newObservedLogger()
	register := func(r *gin.Engine) {
		r.GET("/fail", func(c *gin.Context) {
			_ = c.Error(errors.New("password=hunter2 leaked detail"))
		})
	}

	// Act
	rec := perform([]gin.HandlerFunc{ErrorHandler(log)}, register, http.MethodGet, "/fail")

	// Assert
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "hunter2") {
		t.Error("回應不應洩漏內部錯誤細節")
	}
	if logs.Len() == 0 {
		t.Error("完整錯誤應寫入日誌")
	}
}

func TestCORS_SetsHeadersAndHandlesPreflight(t *testing.T) {
	// Arrange
	register := func(r *gin.Engine) {
		r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	}
	handlers := []gin.HandlerFunc{CORS()}

	// Act
	recGet := perform(handlers, register, http.MethodGet, "/ping")
	recOptions := perform(handlers, register, http.MethodOptions, "/ping")

	// Assert
	if recGet.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("GET 回應應含 CORS header")
	}
	if recOptions.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want 204", recOptions.Code)
	}
}
