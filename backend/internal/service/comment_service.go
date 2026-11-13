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
	// Create 張貼留言：預設 pending；asAdmin 為 true（管理員預覽）時略過功能開關檢查
	Create(in CommentInput, asAdmin bool) (*model.Comment, error)
	// SetStatus 後台變更狀態
	SetStatus(id, status string) (*model.Comment, error)
	// Reply 後台回覆
	Reply(id, reply string) (*model.Comment, error)
}

// taipeiZone 留言日期使用的固定 UTC+8 時區（台灣無日光節約，不依賴 zoneinfo）
var taipeiZone = time.FixedZone("Asia/Taipei", 8*60*60)

// commentService CommentService 實作
type commentService struct {
	repo     repository.CommentRepository
	features FeatureChecker
	now      func() time.Time
}

// NewCommentService 建立 CommentService
func NewCommentService(repo repository.CommentRepository, features FeatureChecker) CommentService {
	return &commentService{repo: repo, features: features, now: time.Now}
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

// Create 建立 pending 留言；缺名稱或內容回 400。
// 非管理員時依 articleId 檢查功能開關（空＝guestbook，其餘＝article_comments）：
// 停用回 403 FEATURE_DISABLED，讀取失敗回 500（fail-closed）。
func (s *commentService) Create(in CommentInput, asAdmin bool) (*model.Comment, error) {
	if in.Name == "" || in.Text == "" {
		return nil, apperror.BadRequest("留言名稱與內容不可為空")
	}
	if !asAdmin {
		if err := s.ensureEnabled(in.ArticleID); err != nil {
			return nil, err
		}
	}
	now := s.now()
	c := &model.Comment{
		ID:        fmt.Sprintf("n%d", now.UnixNano()),
		Name:      in.Name,
		Text:      in.Text,
		Date:      now.In(taipeiZone).Format("2006 / 01 / 02"),
		Status:    model.CommentStatusPending,
		ArticleID: in.ArticleID,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, apperror.Internal(err, "建立留言失敗")
	}
	return c, nil
}

// ensureEnabled 檢查張貼對應的功能開關
func (s *commentService) ensureEnabled(articleID string) error {
	key, label := FeatureArticleComments, "文章留言"
	if articleID == "" {
		key, label = FeatureGuestbook, "留言板投書"
	}
	enabled, err := s.features.IsEnabled(key)
	if err != nil {
		if _, ok := apperror.AsAppError(err); ok {
			return err
		}
		return apperror.Internal(err, "取得功能開關失敗")
	}
	if !enabled {
		return apperror.FeatureDisabled(label + "功能目前已停用")
	}
	return nil
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
