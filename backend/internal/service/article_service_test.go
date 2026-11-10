// backend/internal/service/article_service_test.go
// article service 單元測試：以 mock repository 驗證發布篩選、單篇 404、CRUD 與錯誤包裝

package service

import (
	"errors"
	"net/http"
	"testing"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// mockArticleRepo 測試用 ArticleRepository
type mockArticleRepo struct {
	all         []model.Article
	published   []model.Article
	byID        map[string]*model.Article
	findAllErr  error
	findByIDErr error
	created     *model.Article
	updated     *model.Article
	deletedID   string
	deleteErr   error
}

func (m *mockArticleRepo) FindAll(onlyPublished bool) ([]model.Article, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	if onlyPublished {
		return m.published, nil
	}
	return m.all, nil
}

func (m *mockArticleRepo) FindByID(id string) (*model.Article, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if a, ok := m.byID[id]; ok {
		return a, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockArticleRepo) Create(a *model.Article) error { m.created = a; return nil }
func (m *mockArticleRepo) Update(a *model.Article) error { m.updated = a; return nil }
func (m *mockArticleRepo) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedID = id
	return nil
}
func (m *mockArticleRepo) Count() (int64, error) { return int64(len(m.all)), nil }

func TestArticleListPublished_OnlyReturnsPublished(t *testing.T) {
	// Arrange
	repo := &mockArticleRepo{published: []model.Article{{ID: "a1", Status: model.ArticleStatusPublished}}}
	svc := NewArticleService(repo)

	// Act
	got, err := svc.ListPublished()

	// Assert
	if err != nil {
		t.Fatalf("ListPublished() 應成功: %v", err)
	}
	if len(got) != 1 || got[0].ID != "a1" {
		t.Errorf("回傳內容錯誤: %+v", got)
	}
}

func TestArticleListAll_IncludesDrafts(t *testing.T) {
	// Arrange
	repo := &mockArticleRepo{all: []model.Article{
		{ID: "a1", Status: model.ArticleStatusPublished},
		{ID: "a2", Status: model.ArticleStatusDraft},
	}}
	svc := NewArticleService(repo)

	// Act
	got, err := svc.ListAll()

	// Assert
	if err != nil {
		t.Fatalf("ListAll() 應成功: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("後台列表應含草稿，得到 %+v", got)
	}
}

func TestArticleGet_NotFoundReturns404(t *testing.T) {
	// Arrange
	repo := &mockArticleRepo{byID: map[string]*model.Article{}}
	svc := NewArticleService(repo)

	// Act
	_, err := svc.Get("missing")

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusNotFound {
		t.Fatalf("不存在文章應回 404，得到 %v", err)
	}
}

func TestArticleCreate_DefaultsToDraftWhenStatusInvalid(t *testing.T) {
	// Arrange
	repo := &mockArticleRepo{}
	svc := NewArticleService(repo)

	// Act
	a, err := svc.Create(ArticleInput{Title: "新文章", Status: "garbage"})

	// Assert
	if err != nil {
		t.Fatalf("建立應成功: %v", err)
	}
	if a.Status != model.ArticleStatusDraft {
		t.Errorf("非法狀態應轉為 draft，得到 %q", a.Status)
	}
	if repo.created == nil || repo.created.ID == "" {
		t.Error("應指派非空 id 並寫入 repository")
	}
}

func TestArticleCreate_EmptyTitleReturns400(t *testing.T) {
	// Arrange
	svc := NewArticleService(&mockArticleRepo{})

	// Act
	_, err := svc.Create(ArticleInput{Title: ""})

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusBadRequest {
		t.Fatalf("空標題應回 400，得到 %v", err)
	}
}

func TestArticleUpdate_PublishesDraft(t *testing.T) {
	// Arrange
	existing := &model.Article{ID: "a1", Title: "舊標題", Status: model.ArticleStatusDraft}
	repo := &mockArticleRepo{byID: map[string]*model.Article{"a1": existing}}
	svc := NewArticleService(repo)

	// Act
	a, err := svc.Update("a1", ArticleInput{Title: "新標題", Status: model.ArticleStatusPublished})

	// Assert
	if err != nil {
		t.Fatalf("更新應成功: %v", err)
	}
	if a.Status != model.ArticleStatusPublished || a.Title != "新標題" {
		t.Errorf("更新結果錯誤: %+v", a)
	}
}

func TestArticleDelete_NotFoundReturns404(t *testing.T) {
	// Arrange
	repo := &mockArticleRepo{deleteErr: repository.ErrNotFound}
	svc := NewArticleService(repo)

	// Act
	err := svc.Delete("missing")

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusNotFound {
		t.Fatalf("刪除不存在文章應回 404，得到 %v", err)
	}
}

func TestArticleListPublished_WrapsRepositoryError(t *testing.T) {
	// Arrange
	repo := &mockArticleRepo{findAllErr: errors.New("db down")}
	svc := NewArticleService(repo)

	// Act
	_, err := svc.ListPublished()

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusInternalServerError {
		t.Fatalf("repository 失敗應包裝為 500 AppError，得到 %v", err)
	}
}
