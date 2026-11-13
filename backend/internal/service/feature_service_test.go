// backend/internal/service/feature_service_test.go
// feature service 單元測試：公開對照表、登記表過濾與預設回退、Update 驗證、EnsureDefaults 不覆寫（mock repository）

package service

import (
	"errors"
	"net/http"
	"testing"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// mockFeatureRepo 測試用 FeatureRepository
type mockFeatureRepo struct {
	rows    map[string]*model.FeatureFlag
	saved   []model.FeatureFlag
	created []model.FeatureFlag
	err     error
}

func newMockFeatureRepo(rows ...model.FeatureFlag) *mockFeatureRepo {
	m := &mockFeatureRepo{rows: map[string]*model.FeatureFlag{}}
	for i := range rows {
		r := rows[i]
		m.rows[r.Key] = &r
	}
	return m
}

func (m *mockFeatureRepo) FindAll() ([]model.FeatureFlag, error) {
	if m.err != nil {
		return nil, m.err
	}
	var out []model.FeatureFlag
	for _, r := range m.rows {
		out = append(out, *r)
	}
	return out, nil
}
func (m *mockFeatureRepo) Find(key string) (*model.FeatureFlag, error) {
	if m.err != nil {
		return nil, m.err
	}
	if r, ok := m.rows[key]; ok {
		c := *r
		return &c, nil
	}
	return nil, repository.ErrNotFound
}
func (m *mockFeatureRepo) Save(f *model.FeatureFlag) error {
	if m.err != nil {
		return m.err
	}
	m.saved = append(m.saved, *f)
	c := *f
	m.rows[f.Key] = &c
	return nil
}
func (m *mockFeatureRepo) CreateIfMissing(f *model.FeatureFlag) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.rows[f.Key]; ok {
		return nil
	}
	m.created = append(m.created, *f)
	c := *f
	m.rows[f.Key] = &c
	return nil
}

func boolPtr(b bool) *bool { return &b }

func assertFeatureAppError(t *testing.T, err error, status int) {
	t.Helper()
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != status {
		t.Fatalf("應為 %d AppError，得到 %v", status, err)
	}
}

func TestFeaturePublicMap_UsesDBValuesAndDefaults(t *testing.T) {
	// Arrange：guestbook 在 DB 為 false；其他 key 缺漏
	svc := NewFeatureService(newMockFeatureRepo(model.FeatureFlag{Key: FeatureGuestbook, Enabled: false}))

	// Act
	m, err := svc.PublicMap()

	// Assert
	if err != nil {
		t.Fatalf("PublicMap 應成功: %v", err)
	}
	if len(m) != 3 {
		t.Fatalf("應有 3 個 key，得到 %v", m)
	}
	if m[FeatureGuestbook] {
		t.Error("guestbook 應為 false（取 DB 值）")
	}
	if !m[FeatureSponsor] || !m[FeatureArticleComments] {
		t.Error("DB 缺 key 應回登記表預設值 true")
	}
}

func TestFeaturePublicMap_FiltersKeysNotInRegistry(t *testing.T) {
	// Arrange
	svc := NewFeatureService(newMockFeatureRepo(model.FeatureFlag{Key: "legacy_flag", Enabled: true}))

	// Act
	m, _ := svc.PublicMap()

	// Assert
	if _, ok := m["legacy_flag"]; ok {
		t.Error("登記表沒有的 key 不應出現在公開對照表")
	}
}

func TestFeaturePublicMap_RepoErrorReturns500(t *testing.T) {
	// Arrange
	repo := newMockFeatureRepo()
	repo.err = errors.New("db down")
	svc := NewFeatureService(repo)

	// Act
	_, err := svc.PublicMap()

	// Assert
	assertFeatureAppError(t, err, http.StatusInternalServerError)
}

func TestFeatureList_FollowsRegistryOrderWithMetadata(t *testing.T) {
	// Arrange
	svc := NewFeatureService(newMockFeatureRepo(model.FeatureFlag{Key: FeatureSponsor, Enabled: false}))

	// Act
	list, err := svc.List()

	// Assert
	if err != nil {
		t.Fatalf("List 應成功: %v", err)
	}
	wantKeys := []string{FeatureSponsor, FeatureGuestbook, FeatureArticleComments}
	if len(list) != len(wantKeys) {
		t.Fatalf("長度應為 %d，得到 %d", len(wantKeys), len(list))
	}
	for i, k := range wantKeys {
		if list[i].Key != k {
			t.Errorf("第 %d 項應為 %s，得到 %s", i, k, list[i].Key)
		}
		if list[i].Name == "" || list[i].Description == "" {
			t.Errorf("%s 應帶名稱與說明", k)
		}
	}
	if list[0].Enabled {
		t.Error("sponsor 應為 false")
	}
}

func TestFeatureIsEnabled(t *testing.T) {
	tests := []struct {
		name string
		rows []model.FeatureFlag
		key  string
		want bool
	}{
		{"DB 為 true", []model.FeatureFlag{{Key: FeatureGuestbook, Enabled: true}}, FeatureGuestbook, true},
		{"DB 為 false", []model.FeatureFlag{{Key: FeatureGuestbook, Enabled: false}}, FeatureGuestbook, false},
		{"DB 缺 key 回預設值", nil, FeatureArticleComments, true},
		{"未登記 key 視為停用", []model.FeatureFlag{{Key: "nope", Enabled: true}}, "nope", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			svc := NewFeatureService(newMockFeatureRepo(tt.rows...))

			// Act
			got, err := svc.IsEnabled(tt.key)

			// Assert
			if err != nil {
				t.Fatalf("IsEnabled 應成功: %v", err)
			}
			if got != tt.want {
				t.Errorf("IsEnabled(%s) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestFeatureIsEnabled_RepoErrorPropagates(t *testing.T) {
	// Arrange
	repo := newMockFeatureRepo()
	repo.err = errors.New("db down")
	svc := NewFeatureService(repo)

	// Act
	got, err := svc.IsEnabled(FeatureGuestbook)

	// Assert
	assertFeatureAppError(t, err, http.StatusInternalServerError)
	if got {
		t.Error("讀取失敗不得回傳 true")
	}
}

func TestFeatureUpdate_UnknownKeyReturns404(t *testing.T) {
	// Arrange
	svc := NewFeatureService(newMockFeatureRepo())

	// Act
	_, err := svc.Update("nope", boolPtr(true))

	// Assert
	assertFeatureAppError(t, err, http.StatusNotFound)
}

func TestFeatureUpdate_MissingEnabledReturns400(t *testing.T) {
	// Arrange
	repo := newMockFeatureRepo()
	svc := NewFeatureService(repo)

	// Act
	_, err := svc.Update(FeatureGuestbook, nil)

	// Assert
	assertFeatureAppError(t, err, http.StatusBadRequest)
	if len(repo.saved) != 0 {
		t.Error("缺欄位時不應寫入")
	}
}

func TestFeatureUpdate_UnknownKeyBeatsMissingEnabled(t *testing.T) {
	// Arrange
	svc := NewFeatureService(newMockFeatureRepo())

	// Act：未知 key 且缺欄位
	_, err := svc.Update("nope", nil)

	// Assert：先判斷 key，回 404
	assertFeatureAppError(t, err, http.StatusNotFound)
}

func TestFeatureUpdate_FalseIsPersisted(t *testing.T) {
	// Arrange：DB 缺列時也要能寫入 false
	repo := newMockFeatureRepo()
	svc := NewFeatureService(repo)

	// Act
	v, err := svc.Update(FeatureGuestbook, boolPtr(false))

	// Assert
	if err != nil {
		t.Fatalf("Update 應成功: %v", err)
	}
	if v.Key != FeatureGuestbook || v.Enabled || v.Name == "" {
		t.Errorf("回傳內容不正確: %+v", v)
	}
	if len(repo.saved) != 1 || repo.saved[0].Enabled {
		t.Errorf("應寫入 enabled=false: %+v", repo.saved)
	}
}

func TestFeatureUpdate_SaveErrorReturns500(t *testing.T) {
	// Arrange
	repo := newMockFeatureRepo()
	repo.err = errors.New("db down")
	svc := NewFeatureService(repo)

	// Act
	_, err := svc.Update(FeatureGuestbook, boolPtr(true))

	// Assert
	assertFeatureAppError(t, err, http.StatusInternalServerError)
}

func TestFeatureEnsureDefaults_CreatesMissingOnly(t *testing.T) {
	// Arrange：guestbook 已被管理員停用
	repo := newMockFeatureRepo(model.FeatureFlag{Key: FeatureGuestbook, Enabled: false})
	svc := NewFeatureService(repo)

	// Act
	err := svc.EnsureDefaults()

	// Assert
	if err != nil {
		t.Fatalf("EnsureDefaults 應成功: %v", err)
	}
	if len(repo.created) != 2 {
		t.Fatalf("應只補 2 個缺漏 key，得到 %v", repo.created)
	}
	if repo.rows[FeatureGuestbook].Enabled {
		t.Error("既有 guestbook=false 不得被覆寫")
	}
	if !repo.rows[FeatureSponsor].Enabled {
		t.Error("補齊的 key 應為預設值 true")
	}
}

func TestFeatureEnsureDefaults_RepoErrorReturns500(t *testing.T) {
	// Arrange
	repo := newMockFeatureRepo()
	repo.err = errors.New("db down")
	svc := NewFeatureService(repo)

	// Act
	err := svc.EnsureDefaults()

	// Assert
	assertFeatureAppError(t, err, http.StatusInternalServerError)
}
