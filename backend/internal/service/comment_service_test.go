// backend/internal/service/comment_service_test.go
// comment service 單元測試：張貼預設 pending、空內容 400、狀態轉換與公開篩選（mock repository）

package service

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// mockCommentRepo 測試用 CommentRepository
type mockCommentRepo struct {
	visible []model.Comment
	all     []model.Comment
	byID    map[string]*model.Comment
	created *model.Comment
	updated *model.Comment
}

func (m *mockCommentRepo) FindVisible(articleID string) ([]model.Comment, error) {
	if articleID == "" {
		return m.visible, nil
	}
	var out []model.Comment
	for _, c := range m.visible {
		if c.ArticleID == articleID {
			out = append(out, c)
		}
	}
	return out, nil
}
func (m *mockCommentRepo) FindAll() ([]model.Comment, error) { return m.all, nil }
func (m *mockCommentRepo) FindByID(id string) (*model.Comment, error) {
	if c, ok := m.byID[id]; ok {
		return c, nil
	}
	return nil, repository.ErrNotFound
}
func (m *mockCommentRepo) Create(c *model.Comment) error { m.created = c; return nil }
func (m *mockCommentRepo) Update(c *model.Comment) error { m.updated = c; return nil }
func (m *mockCommentRepo) Count() (int64, error)         { return int64(len(m.all)), nil }

// allFeaturesOn 所有開關皆啟用的 FeatureChecker
type allFeaturesOn struct{}

func (allFeaturesOn) IsEnabled(string) (bool, error) { return true, nil }

// stubFeatures 可指定各 key 狀態並記錄被查詢的 key
type stubFeatures struct {
	states  map[string]bool
	err     error
	queried []string
}

func (s *stubFeatures) IsEnabled(key string) (bool, error) {
	s.queried = append(s.queried, key)
	return s.states[key], s.err
}

func TestCommentCreate_DefaultsToPending(t *testing.T) {
	// Arrange
	repo := &mockCommentRepo{}
	svc := NewCommentService(repo, allFeaturesOn{})

	// Act
	c, err := svc.Create(CommentInput{Name: "訪客", Text: "你好"}, false)

	// Assert
	if err != nil {
		t.Fatalf("張貼應成功: %v", err)
	}
	if c.Status != model.CommentStatusPending {
		t.Errorf("新留言應為 pending，得到 %q", c.Status)
	}
	if repo.created == nil || repo.created.ID == "" {
		t.Error("應寫入 repository 並指派 id")
	}
}

func TestCommentCreate_EmptyTextReturns400(t *testing.T) {
	// Arrange
	svc := NewCommentService(&mockCommentRepo{}, allFeaturesOn{})

	// Act
	_, err := svc.Create(CommentInput{Name: "訪客", Text: ""}, false)

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusBadRequest {
		t.Fatalf("空內容應回 400，得到 %v", err)
	}
}

func TestCommentListVisible_OnlyVisible(t *testing.T) {
	// Arrange：repository 已負責過濾，service 直接轉交
	repo := &mockCommentRepo{visible: []model.Comment{
		{ID: "c1", Status: model.CommentStatusVisible},
	}}
	svc := NewCommentService(repo, allFeaturesOn{})

	// Act
	got, err := svc.ListVisible("")

	// Assert
	if err != nil {
		t.Fatalf("公開讀取應成功: %v", err)
	}
	if len(got) != 1 || got[0].Status != model.CommentStatusVisible {
		t.Errorf("公開結果應僅含 visible，得到 %+v", got)
	}
}

func TestCommentSetStatus_ApprovePending(t *testing.T) {
	// Arrange
	existing := &model.Comment{ID: "c4", Status: model.CommentStatusPending}
	repo := &mockCommentRepo{byID: map[string]*model.Comment{"c4": existing}}
	svc := NewCommentService(repo, allFeaturesOn{})

	// Act
	c, err := svc.SetStatus("c4", model.CommentStatusVisible)

	// Assert
	if err != nil {
		t.Fatalf("狀態變更應成功: %v", err)
	}
	if c.Status != model.CommentStatusVisible {
		t.Errorf("狀態應改為 visible，得到 %q", c.Status)
	}
}

func TestCommentSetStatus_InvalidStatusReturns400(t *testing.T) {
	// Arrange
	repo := &mockCommentRepo{byID: map[string]*model.Comment{"c1": {ID: "c1"}}}
	svc := NewCommentService(repo, allFeaturesOn{})

	// Act
	_, err := svc.SetStatus("c1", "garbage")

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusBadRequest {
		t.Fatalf("非法狀態應回 400，得到 %v", err)
	}
}

func TestCommentReply_SetsReply(t *testing.T) {
	// Arrange
	existing := &model.Comment{ID: "c1"}
	repo := &mockCommentRepo{byID: map[string]*model.Comment{"c1": existing}}
	svc := NewCommentService(repo, allFeaturesOn{})

	// Act
	c, err := svc.Reply("c1", "謝謝留言")

	// Assert
	if err != nil {
		t.Fatalf("回覆應成功: %v", err)
	}
	if c.Reply != "謝謝留言" {
		t.Errorf("回覆內容錯誤: %q", c.Reply)
	}
}

func TestCommentSetStatus_NotFoundReturns404(t *testing.T) {
	// Arrange
	repo := &mockCommentRepo{byID: map[string]*model.Comment{}}
	svc := NewCommentService(repo, allFeaturesOn{})

	// Act
	_, err := svc.SetStatus("missing", model.CommentStatusVisible)

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusNotFound {
		t.Fatalf("不存在留言應回 404，得到 %v", err)
	}
}

func TestCommentCreate_DateUsesTaipeiTime(t *testing.T) {
	cases := []struct {
		name string
		now  time.Time
		want string
	}{
		{"UTC 前一日 19:00 跨日", time.Date(2026, 11, 11, 19, 0, 0, 0, time.UTC), "2026 / 11 / 12"},
		{"UTC 當日 07:00", time.Date(2026, 11, 12, 7, 0, 0, 0, time.UTC), "2026 / 11 / 12"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			svc := &commentService{repo: &mockCommentRepo{}, features: allFeaturesOn{}, now: func() time.Time { return tc.now }}

			// Act
			c, err := svc.Create(CommentInput{Name: "訪客", Text: "你好"}, false)

			// Assert
			if err != nil {
				t.Fatalf("張貼應成功: %v", err)
			}
			if c.Date != tc.want {
				t.Errorf("date 應為 %q，得到 %q", tc.want, c.Date)
			}
		})
	}
}

func TestCommentCreate_GuestbookDisabledReturns403(t *testing.T) {
	// Arrange
	repo := &mockCommentRepo{}
	feat := &stubFeatures{states: map[string]bool{FeatureGuestbook: false, FeatureArticleComments: true}}
	svc := NewCommentService(repo, feat)

	// Act：articleId 為空 → 留言板
	_, err := svc.Create(CommentInput{Name: "訪客", Text: "你好"}, false)

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusForbidden || appErr.Code != apperror.CodeFeatureDisabled {
		t.Fatalf("應回 403 FEATURE_DISABLED，得到 %v", err)
	}
	if repo.created != nil {
		t.Error("被拒絕時不應寫入")
	}
	if len(feat.queried) != 1 || feat.queried[0] != FeatureGuestbook {
		t.Errorf("應只查 guestbook，得到 %v", feat.queried)
	}
}

func TestCommentCreate_ArticleCommentsDisabledReturns403(t *testing.T) {
	// Arrange：guestbook 開、article_comments 關
	repo := &mockCommentRepo{}
	feat := &stubFeatures{states: map[string]bool{FeatureGuestbook: true, FeatureArticleComments: false}}
	svc := NewCommentService(repo, feat)

	// Act
	_, err := svc.Create(CommentInput{Name: "訪客", Text: "你好", ArticleID: "a1"}, false)

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Code != apperror.CodeFeatureDisabled {
		t.Fatalf("應回 FEATURE_DISABLED，得到 %v", err)
	}
	if len(feat.queried) != 1 || feat.queried[0] != FeatureArticleComments {
		t.Errorf("有 articleId 應查 article_comments，得到 %v", feat.queried)
	}
}

func TestCommentCreate_EnabledCreatesPending(t *testing.T) {
	// Arrange
	repo := &mockCommentRepo{}
	feat := &stubFeatures{states: map[string]bool{FeatureGuestbook: true, FeatureArticleComments: true}}
	svc := NewCommentService(repo, feat)

	// Act
	c1, err1 := svc.Create(CommentInput{Name: "訪客", Text: "留言板"}, false)
	c2, err2 := svc.Create(CommentInput{Name: "訪客", Text: "文章", ArticleID: "a1"}, false)

	// Assert
	if err1 != nil || err2 != nil {
		t.Fatalf("啟用時應成功: %v %v", err1, err2)
	}
	if c1.Status != model.CommentStatusPending || c2.Status != model.CommentStatusPending {
		t.Error("應為 pending")
	}
	if len(feat.queried) != 2 || feat.queried[0] != FeatureGuestbook || feat.queried[1] != FeatureArticleComments {
		t.Errorf("查詢的 key 不正確: %v", feat.queried)
	}
}

func TestCommentCreate_FeatureReadErrorReturns500(t *testing.T) {
	// Arrange
	repo := &mockCommentRepo{}
	feat := &stubFeatures{err: errors.New("db down")}
	svc := NewCommentService(repo, feat)

	// Act
	_, err := svc.Create(CommentInput{Name: "訪客", Text: "你好"}, false)

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusInternalServerError {
		t.Fatalf("讀取失敗應回 500，得到 %v", err)
	}
	if repo.created != nil {
		t.Error("讀取失敗不得放行寫入")
	}
}

func TestCommentCreate_AdminSkipsFeatureCheck(t *testing.T) {
	// Arrange：兩個開關都關，且讀取會失敗；管理員仍應成功且不查開關
	repo := &mockCommentRepo{}
	feat := &stubFeatures{err: errors.New("db down")}
	svc := NewCommentService(repo, feat)

	// Act
	c, err := svc.Create(CommentInput{Name: "管理員", Text: "預覽"}, true)

	// Assert
	if err != nil {
		t.Fatalf("管理員應略過開關: %v", err)
	}
	if c.Status != model.CommentStatusPending {
		t.Errorf("管理員預覽留言仍應為 pending，得到 %q", c.Status)
	}
	if len(feat.queried) != 0 {
		t.Errorf("管理員不應查開關，得到 %v", feat.queried)
	}
}

func TestCommentCreate_ValidationBeforeFeatureCheck(t *testing.T) {
	// Arrange
	feat := &stubFeatures{states: map[string]bool{}}
	svc := NewCommentService(&mockCommentRepo{}, feat)

	// Act
	_, err := svc.Create(CommentInput{Name: "", Text: ""}, false)

	// Assert：驗證先於開關，回 400
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusBadRequest {
		t.Fatalf("應回 400，得到 %v", err)
	}
}
