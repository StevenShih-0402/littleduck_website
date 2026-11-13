// backend/internal/pkg/apperror/apperror_test.go
// apperror 套件單元測試：驗證 AppError 的錯誤介面、包裝與常用建構子

package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_ImplementsErrorInterface(t *testing.T) {
	// Arrange
	appErr := New(CodeNotFound, http.StatusNotFound, "文章不存在")

	// Act
	msg := appErr.Error()

	// Assert
	if msg != "文章不存在" {
		t.Errorf("Error() = %q, want %q", msg, "文章不存在")
	}
}

func TestAppError_CarriesCodeAndStatus(t *testing.T) {
	// Arrange & Act
	appErr := New(CodeNotFound, http.StatusNotFound, "文章不存在")

	// Assert
	if appErr.Code != CodeNotFound {
		t.Errorf("Code = %q, want %q", appErr.Code, CodeNotFound)
	}
	if appErr.Status != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", appErr.Status, http.StatusNotFound)
	}
}

func TestWrap_PreservesUnderlyingError(t *testing.T) {
	// Arrange
	cause := errors.New("connection refused")

	// Act
	appErr := Wrap(cause, CodeInternal, http.StatusInternalServerError, "系統錯誤")

	// Assert
	if !errors.Is(appErr, cause) {
		t.Error("errors.Is 應能追溯被包裝的底層錯誤")
	}
}

func TestNotFound_BuildsNotFoundError(t *testing.T) {
	// Arrange & Act
	appErr := NotFound("文章不存在")

	// Assert
	if appErr.Status != http.StatusNotFound || appErr.Code != CodeNotFound {
		t.Errorf("NotFound() = %+v, want status 404 / code %q", appErr, CodeNotFound)
	}
}

func TestBadRequest_BuildsBadRequestError(t *testing.T) {
	// Arrange & Act
	appErr := BadRequest("參數格式錯誤")

	// Assert
	if appErr.Status != http.StatusBadRequest || appErr.Code != CodeBadRequest {
		t.Errorf("BadRequest() = %+v, want status 400 / code %q", appErr, CodeBadRequest)
	}
}

func TestNotImplemented_BuildsNotImplementedError(t *testing.T) {
	// Arrange & Act
	appErr := NotImplemented("尚未實作")

	// Assert
	if appErr.Status != http.StatusNotImplemented || appErr.Code != CodeNotImplemented {
		t.Errorf("NotImplemented() = %+v, want status 501 / code %q", appErr, CodeNotImplemented)
	}
}

func TestAsAppError_ExtractsFromWrappedChain(t *testing.T) {
	// Arrange
	appErr := NotFound("文章不存在")
	wrapped := errors.Join(errors.New("outer"), appErr)

	// Act
	got, ok := AsAppError(wrapped)

	// Assert
	if !ok {
		t.Fatal("AsAppError 應能從錯誤鏈中取出 AppError")
	}
	if got.Code != CodeNotFound {
		t.Errorf("Code = %q, want %q", got.Code, CodeNotFound)
	}
}

func TestAsAppError_ReturnsFalseForPlainError(t *testing.T) {
	// Arrange
	plain := errors.New("plain error")

	// Act
	_, ok := AsAppError(plain)

	// Assert
	if ok {
		t.Error("一般 error 不應被辨識為 AppError")
	}
}

func TestFeatureDisabled_Returns403WithCode(t *testing.T) {
	// Act
	err := FeatureDisabled("功能已停用")

	// Assert
	if err.Status != http.StatusForbidden {
		t.Errorf("Status = %d, want 403", err.Status)
	}
	if err.Code != "FEATURE_DISABLED" {
		t.Errorf("Code = %q, want FEATURE_DISABLED", err.Code)
	}
	if err.Message != "功能已停用" {
		t.Errorf("Message = %q", err.Message)
	}
}
