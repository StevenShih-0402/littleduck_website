// backend/internal/service/comment_service.go
// Comment service：留言業務邏輯——張貼（預設 pending）、公開讀取、後台審核與回覆

package service

import (
	"errors"
	"fmt"
	"time"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// CommentInput 訪客張貼留言的輸入
type CommentInput struct {
	Name      string `json:"name"`
	Text      string `json:"text"`
	ArticleID string `json:"articleId"`
}

// CommentService 留言業務邏輯介面
type CommentService interface {
	// ListVisible 公開讀取：僅可見留言
	ListVisible(articleID string) ([]model.Comment, error)
	// ListAll 後台讀取：全部留言
	ListAll() ([]model.Comment, error)
	// Create 訪客張貼：預設 pending
	Create(in CommentInput) (*model.Comment, error)
	// SetStatus 後台變更狀態
	SetStatus(id, status string) (*model.Comment, error)
	// Reply 後台回覆
	Reply(id, reply string) (*model.Comment, error)
}

// commentService CommentService 實作
type commentService struct {
	repo repository.CommentRepository
	now  func() time.Time
}

// NewCommentService 建立 CommentService
func NewCommentService(repo repository.CommentRepository) CommentService {
	return &commentService{repo: repo, now: time.Now}
}

// ListVisible 取得可見留言
func (s *commentService) ListVisible(articleID string) ([]model.Comment, error) {
	comments, err := s.repo.FindVisible(articleID)
	if err != nil {
		return nil, apperror.Internal(err, "取得留言失敗")
	}
	return comments, nil
}

// ListAll 取得全部留言
func (s *commentService) ListAll() ([]model.Comment, error) {
	comments, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.Internal(err, "取得留言失敗")
	}
	return comments, nil
}

// Create 建立 pending 留言；缺名稱或內容回 400
func (s *commentService) Create(in CommentInput) (*model.Comment, error) {
	if in.Name == "" || in.Text == "" {
		return nil, apperror.BadRequest("留言名稱與內容不可為空")
	}
	now := s.now()
	c := &model.Comment{
		ID:        fmt.Sprintf("n%d", now.UnixNano()),
		Name:      in.Name,
		Text:      in.Text,
		Date:      now.Format("2006 / 01 / 02"),
		Status:    model.CommentStatusPending,
		ArticleID: in.ArticleID,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, apperror.Internal(err, "建立留言失敗")
	}
	return c, nil
}

// SetStatus 變更留言狀態（需為合法狀態）
func (s *commentService) SetStatus(id, status string) (*model.Comment, error) {
	if !isValidCommentStatus(status) {
		return nil, apperror.BadRequest("留言狀態不合法")
	}
	c, err := s.get(id)
	if err != nil {
		return nil, err
	}
	c.Status = status
	if err := s.repo.Update(c); err != nil {
		return nil, apperror.Internal(err, "更新留言失敗")
	}
	return c, nil
}

// Reply 對留言新增回覆
func (s *commentService) Reply(id, reply string) (*model.Comment, error) {
	c, err := s.get(id)
	if err != nil {
		return nil, err
	}
	c.Reply = reply
	if err := s.repo.Update(c); err != nil {
		return nil, apperror.Internal(err, "更新留言失敗")
	}
	return c, nil
}

// get 內部取單則留言；查無回 404
func (s *commentService) get(id string) (*model.Comment, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.NotFound("留言不存在")
		}
		return nil, apperror.Internal(err, "取得留言失敗")
	}
	return c, nil
}

// isValidCommentStatus 檢查狀態是否為合法值
func isValidCommentStatus(status string) bool {
	switch status {
	case model.CommentStatusPending, model.CommentStatusVisible, model.CommentStatusHidden:
		return true
	default:
		return false
	}
}
