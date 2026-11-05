// backend/internal/pkg/response/response_test.go
// response 套件單元測試：驗證成功與錯誤回應的統一 JSON envelope 格式

package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/apperror"
)

// envelope 測試用的回應結構
type envelope struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// newTestContext 建立測試用 gin.Context 與 recorder
func newTestContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	return c, rec
}

func TestOK_WrapsDataInEnvelope(t *testing.T) {
	// Arrange
	c, rec := newTestContext(t)
	data := map[string]string{"title": "hello"}

	// Act
	OK(c, data)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("回應不是合法 JSON: %v", err)
	}
	if got.Code != CodeSuccess {
		t.Errorf("code = %q, want %q", got.Code, CodeSuccess)
	}
	var payload map[string]string
	if err := json.Unmarshal(got.Data, &payload); err != nil || payload["title"] != "hello" {
		t.Errorf("data 內容錯誤: %s", got.Data)
	}
}

func TestError_UsesAppErrorStatusAndCode(t *testing.T) {
	// Arrange
	c, rec := newTestContext(t)
	appErr := apperror.NotFound("文章不存在")

	// Act
	Error(c, appErr)

	// Assert
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var got envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("回應不是合法 JSON: %v", err)
	}
	if got.Code != apperror.CodeNotFound || got.Message != "文章不存在" {
		t.Errorf("envelope = %+v, want code %q message %q", got, apperror.CodeNotFound, "文章不存在")
	}
}
