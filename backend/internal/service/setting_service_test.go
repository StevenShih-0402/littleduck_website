// backend/internal/service/setting_service_test.go
// setting service 單元測試：讀取、更新與基本驗證（mock repository）

package service

import (
	"net/http"
	"testing"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// mockSettingRepo 測試用 SettingRepository
type mockSettingRepo struct {
	current *model.SiteSetting
	saved   *model.SiteSetting
}

func (m *mockSettingRepo) Get() (*model.SiteSetting, error) {
	if m.current == nil {
		return nil, repository.ErrNotFound
	}
	return m.current, nil
}
func (m *mockSettingRepo) Save(s *model.SiteSetting) error {
	m.saved = s
	m.current = s
	return nil
}
func (m *mockSettingRepo) Exists() (bool, error) { return m.current != nil, nil }

func TestSettingUpdate_ReflectsInSubsequentGet(t *testing.T) {
	// Arrange
	repo := &mockSettingRepo{current: &model.SiteSetting{Title: "舊標題"}}
	svc := NewSettingService(repo)

	// Act
	_, err := svc.Update(model.SiteSetting{Title: "新標題", Tagline: "新標語"})
	got, getErr := svc.Get()

	// Assert
	if err != nil || getErr != nil {
		t.Fatalf("更新與讀取應成功: %v / %v", err, getErr)
	}
	if got.Title != "新標題" || got.Tagline != "新標語" {
		t.Errorf("更新後讀取應反映新值，得到 %+v", got)
	}
}

func TestSettingUpdate_EmptyTitleReturns400(t *testing.T) {
	// Arrange
	svc := NewSettingService(&mockSettingRepo{})

	// Act
	_, err := svc.Update(model.SiteSetting{Title: ""})

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusBadRequest {
		t.Fatalf("空標題應回 400，得到 %v", err)
	}
}

func TestSettingGet_NotFoundReturns404(t *testing.T) {
	// Arrange
	svc := NewSettingService(&mockSettingRepo{})

	// Act
	_, err := svc.Get()

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusNotFound {
		t.Fatalf("未初始化設定應回 404，得到 %v", err)
	}
}

func TestSettingUpdate_SocialURLValidation(t *testing.T) {
	valid := []string{"", "#", "/blog/a1", "https://x", "HTTP://X", "mailto:a@b", " https://x "}
	invalid := []string{"javascript:alert(1)", "JavaScript:x", "data:text/html,x", "//evil.com", "foo/bar", "ftp://x"}
	for _, u := range valid {
		repo := &mockSettingRepo{}
		_, err := NewSettingService(repo).Update(model.SiteSetting{Title: "t", Social: []model.SocialLink{{Label: "a", URL: u}}})
		if err != nil || repo.saved == nil {
			t.Errorf("%q 應合法並寫入，得到 %v", u, err)
		}
	}
	for _, u := range invalid {
		repo := &mockSettingRepo{}
		_, err := NewSettingService(repo).Update(model.SiteSetting{Title: "t", Social: []model.SocialLink{{Label: "a", URL: u}}})
		appErr, ok := apperror.AsAppError(err)
		if !ok || appErr.Status != http.StatusBadRequest || appErr.Message != "社群連結網址格式不正確" {
			t.Errorf("%q 應回 400，得到 %v", u, err)
		}
		if repo.saved != nil {
			t.Errorf("%q 不應寫入", u)
		}
	}
}
