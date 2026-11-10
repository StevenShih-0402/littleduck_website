// backend/internal/service/auth/auth_test.go
// auth service 單元測試：帳密驗證、權杖簽發與解析（含過期），不需 DB

package auth

import (
	"net/http"
	"testing"
	"time"

	"littleduck-backend/internal/config"
	"littleduck-backend/internal/pkg/apperror"
)

// newTestService 建立固定帳密與短效期的 service
func newTestService(ttl time.Duration) *authService {
	return &authService{
		username: "admin",
		password: "s3cret",
		secret:   []byte("unit-test-secret"),
		ttl:      ttl,
		now:      time.Now,
	}
}

func TestLogin_SuccessIssuesToken(t *testing.T) {
	// Arrange
	svc := newTestService(time.Hour)

	// Act
	tok, err := svc.Login("admin", "s3cret")

	// Assert
	if err != nil {
		t.Fatalf("正確帳密應成功，卻回傳錯誤: %v", err)
	}
	if tok.Token == "" {
		t.Error("應回傳非空權杖")
	}
	if !tok.ExpiresAt.After(time.Now()) {
		t.Error("過期時間應晚於現在")
	}
}

func TestLogin_WrongPasswordReturns401(t *testing.T) {
	// Arrange
	svc := newTestService(time.Hour)

	// Act
	_, err := svc.Login("admin", "wrong")

	// Assert
	assertUnauthorized(t, err)
}

func TestLogin_WrongUsernameReturns401(t *testing.T) {
	// Arrange
	svc := newTestService(time.Hour)

	// Act
	_, err := svc.Login("intruder", "s3cret")

	// Assert
	assertUnauthorized(t, err)
}

func TestParse_ValidTokenReturnsSubject(t *testing.T) {
	// Arrange
	svc := newTestService(time.Hour)
	tok, _ := svc.Login("admin", "s3cret")

	// Act
	subject, err := svc.Parse(tok.Token)

	// Assert
	if err != nil {
		t.Fatalf("有效權杖應解析成功，卻回傳錯誤: %v", err)
	}
	if subject != "admin" {
		t.Errorf("subject = %q, want admin", subject)
	}
}

func TestParse_ExpiredTokenReturns401(t *testing.T) {
	// Arrange：以過去時間簽發，使權杖立即過期
	svc := newTestService(time.Minute)
	svc.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	tok, _ := svc.Login("admin", "s3cret")
	svc.now = time.Now // 還原為現在時間驗證

	// Act
	_, err := svc.Parse(tok.Token)

	// Assert
	assertUnauthorized(t, err)
}

func TestParse_TamperedTokenReturns401(t *testing.T) {
	// Arrange
	svc := newTestService(time.Hour)

	// Act
	_, err := svc.Parse("not.a.jwt")

	// Assert
	assertUnauthorized(t, err)
}

func TestNewService_WiresConfig(t *testing.T) {
	// Arrange
	svc := NewService(config.AuthConfig{
		AdminUsername: "u", AdminPassword: "p",
		JWTSecret: "k", JWTTTL: time.Hour,
	})

	// Act
	_, err := svc.Login("u", "p")

	// Assert
	if err != nil {
		t.Fatalf("以 config 建立的 service 應能登入，卻回傳錯誤: %v", err)
	}
}

func assertUnauthorized(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("應回傳錯誤")
	}
	appErr, ok := apperror.AsAppError(err)
	if !ok {
		t.Fatalf("錯誤應為 AppError，得到 %T", err)
	}
	if appErr.Status != http.StatusUnauthorized {
		t.Errorf("Status = %d, want 401", appErr.Status)
	}
}
