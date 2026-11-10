// backend/internal/controller/article_controller.go
// Article controller：HTTP 層，綁定請求、呼叫 service、以統一 envelope 回應

package controller

import (
	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/apperror"
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

// List 公開文章列表（僅已發布）
//
//	@Summary		公開文章列表
//	@Description	取得已發布文章（依建立時間新到舊），草稿不列入
//	@Tags			articles
//	@Produce		json
//	@Success		200	{object}	response.Envelope{data=[]model.Article}
//	@Failure		500	{object}	response.Envelope
//	@Router			/api/v1/articles [get]
func (ctrl *ArticleController) List(c *gin.Context) {
	articles, err := ctrl.svc.ListPublished()
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, articles)
}

// Get 公開讀取單篇文章
//
//	@Summary		讀取單篇文章
//	@Description	以文章 id 取得單篇完整內容
//	@Tags			articles
//	@Produce		json
//	@Param			id	path		string	true	"文章 id"
//	@Success		200	{object}	response.Envelope{data=model.Article}
//	@Failure		404	{object}	response.Envelope
//	@Router			/api/v1/articles/{id} [get]
func (ctrl *ArticleController) Get(c *gin.Context) {
	a, err := ctrl.svc.Get(c.Param("id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, a)
}

// AdminList 後台文章列表（含草稿）
//
//	@Summary		後台文章列表
//	@Description	取得全部文章（含草稿），需管理員身分
//	@Tags			articles
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]model.Article}
//	@Failure		401	{object}	response.Envelope
//	@Router			/api/v1/admin/articles [get]
func (ctrl *ArticleController) AdminList(c *gin.Context) {
	articles, err := ctrl.svc.ListAll()
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, articles)
}

// Create 建立文章
//
//	@Summary		建立文章
//	@Description	建立文章（含發布狀態），需管理員身分
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		service.ArticleInput	true	"文章內容"
//	@Success		201		{object}	response.Envelope{data=model.Article}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Router			/api/v1/articles [post]
func (ctrl *ArticleController) Create(c *gin.Context) {
	var in service.ArticleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperror.BadRequest("文章內容格式錯誤"))
		return
	}
	a, err := ctrl.svc.Create(in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, a)
}

// Update 修改文章
//
//	@Summary		修改文章
//	@Description	修改文章內容與發布狀態，需管理員身分
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"文章 id"
//	@Param			body	body		service.ArticleInput	true	"文章內容"
//	@Success		200		{object}	response.Envelope{data=model.Article}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/api/v1/articles/{id} [put]
func (ctrl *ArticleController) Update(c *gin.Context) {
	var in service.ArticleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperror.BadRequest("文章內容格式錯誤"))
		return
	}
	a, err := ctrl.svc.Update(c.Param("id"), in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, a)
}

// Delete 刪除文章
//
//	@Summary		刪除文章
//	@Description	刪除文章，需管理員身分
//	@Tags			articles
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"文章 id"
//	@Success		200	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/api/v1/articles/{id} [delete]
func (ctrl *ArticleController) Delete(c *gin.Context) {
	if err := ctrl.svc.Delete(c.Param("id")); err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, nil)
}
