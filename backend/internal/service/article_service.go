// backend/internal/service/article_service.go
// Article service：文章業務邏輯介面與實作，依賴 repository 介面

package service

import (
	"errors"
	"fmt"
	"time"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// ArticleInput 建立／修改文章的輸入（不含唯讀欄位 views／時間戳）
type ArticleInput struct {
	Title      string   `json:"title"`
	Cat        string   `json:"cat"`
	Date       string   `json:"date"`
	Y          int      `json:"y"`
	Status     string   `json:"status"`
	Blurb      string   `json:"blurb"`
	Award      string   `json:"award"`
	Banner     string   `json:"banner"`
	BannerFade bool     `json:"bannerFade"`
	Showcase   bool     `json:"showcase"`
	WorkCat    string   `json:"workCat"`
	Gallery    []string `json:"gallery"`
	Body       string   `json:"body"`
}

// ArticleService 文章業務邏輯介面（controller 依賴此介面，測試以 mock 替換）
type ArticleService interface {
	// ListPublished 公開列表：僅已發布
	ListPublished() ([]model.Article, error)
	// ListAll 後台列表：含草稿
	ListAll() ([]model.Article, error)
	// Get 單篇讀取；查無回傳 404
	Get(id string) (*model.Article, error)
	Create(in ArticleInput) (*model.Article, error)
	Update(id string, in ArticleInput) (*model.Article, error)
	Delete(id string) error
}

// articleService ArticleService 實作
type articleService struct {
	repo repository.ArticleRepository
	now  func() time.Time
}

// NewArticleService 建立 ArticleService
func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{repo: repo, now: time.Now}
}

// ListPublished 取得已發布文章列表
func (s *articleService) ListPublished() ([]model.Article, error) {
	articles, err := s.repo.FindAll(true)
	if err != nil {
		return nil, apperror.Internal(err, "取得文章列表失敗")
	}
	return articles, nil
}

// ListAll 取得全部文章（後台用，含草稿）
func (s *articleService) ListAll() ([]model.Article, error) {
	articles, err := s.repo.FindAll(false)
	if err != nil {
		return nil, apperror.Internal(err, "取得文章列表失敗")
	}
	return articles, nil
}

// Get 以 id 取得單篇
func (s *articleService) Get(id string) (*model.Article, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.NotFound("文章不存在")
		}
		return nil, apperror.Internal(err, "取得文章失敗")
	}
	return a, nil
}

// Create 建立文章；status 空值預設 draft
func (s *articleService) Create(in ArticleInput) (*model.Article, error) {
	if in.Title == "" {
		return nil, apperror.BadRequest("標題不可為空")
	}
	status := normalizeStatus(in.Status)
	a := applyInput(&model.Article{ID: s.genID()}, in)
	a.Status = status
	if err := s.repo.Create(a); err != nil {
		return nil, apperror.Internal(err, "建立文章失敗")
	}
	return a, nil
}

// Update 修改既有文章（含發布狀態切換）
func (s *articleService) Update(id string, in ArticleInput) (*model.Article, error) {
	existing, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if in.Title == "" {
		return nil, apperror.BadRequest("標題不可為空")
	}
	applyInput(existing, in)
	existing.Status = normalizeStatus(in.Status)
	if err := s.repo.Update(existing); err != nil {
		return nil, apperror.Internal(err, "更新文章失敗")
	}
	return existing, nil
}

// Delete 刪除文章
func (s *articleService) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.NotFound("文章不存在")
		}
		return apperror.Internal(err, "刪除文章失敗")
	}
	return nil
}

// genID 以奈秒時間戳產生唯一字串 id（前綴 n 以與種子 id a/p 區隔）
func (s *articleService) genID() string {
	return fmt.Sprintf("n%d", s.now().UnixNano())
}

// applyInput 將可編輯欄位套用到 Article（不動 ID／Views／時間戳）
func applyInput(a *model.Article, in ArticleInput) *model.Article {
	a.Title = in.Title
	a.Cat = in.Cat
	a.Date = in.Date
	a.Y = in.Y
	a.Blurb = in.Blurb
	a.Award = in.Award
	a.Banner = in.Banner
	a.BannerFade = in.BannerFade
	a.Showcase = in.Showcase
	a.WorkCat = in.WorkCat
	a.Gallery = in.Gallery
	a.Body = in.Body
	return a
}

// normalizeStatus 非 published 一律視為 draft
func normalizeStatus(status string) string {
	if status == model.ArticleStatusPublished {
		return model.ArticleStatusPublished
	}
	return model.ArticleStatusDraft
}
