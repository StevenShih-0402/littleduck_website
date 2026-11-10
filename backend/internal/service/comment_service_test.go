// backend/internal/service/comment_service_test.go
// comment service 單元測試：張貼預設 pending、空內容 400、狀態轉換與公開篩選（mock repository）

package service

import (
	"net/http"
	"testing"

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

func TestCommentCreate_DefaultsToPending(t *testing.T) {
	// Arrange
	repo := &mockCommentRepo{}
	svc := NewCommentService(repo)

	// Act
	c, err := svc.Create(CommentInput{Name: "訪客", Text: "你好"})

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
	svc := NewCommentService(&mockCommentRepo{})

	// Act
	_, err := svc.Create(CommentInput{Name: "訪客", Text: ""})

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
	svc := NewCommentService(repo)

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
	svc := NewCommentService(repo)

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
	svc := NewCommentService(repo)

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
	svc := NewCommentService(repo)

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
	svc := NewCommentService(repo)

	// Act
	_, err := svc.SetStatus("missing", model.CommentStatusVisible)

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusNotFound {
		t.Fatalf("不存在留言應回 404，得到 %v", err)
	}
}
