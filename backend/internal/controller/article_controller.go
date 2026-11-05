// backend/internal/controller/article_controller.go
// Article controller：HTTP 層，綁定請求、呼叫 service、以統一 envelope 回應

package controller

import (
	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/response"
	"littleduck-backend/internal/service"
)

// ArticleController 文章 controller
type ArticleController struct {
	svc service.ArticleService
}

// NewArticleController 建立 ArticleController
func NewArticleController(svc service.ArticleService) *ArticleController {
	return &ArticleController{svc: svc}
}

// List 取得文章列表
//
//	@Summary		文章列表
//	@Description	取得全部文章（依建立時間新到舊）
//	@Tags			articles
//	@Produce		json
//	@Success		200	{object}	response.Envelope{data=[]model.Article}
//	@Failure		500	{object}	response.Envelope
//	@Router			/api/v1/articles [get]
func (ctrl *ArticleController) List(c *gin.Context) {
	articles, err := ctrl.svc.List()
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, articles)
}
