// backend/internal/controller/comment_controller.go
// Comment controller：公開張貼與讀取、後台審核與回覆

package controller

import (
	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/middleware"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/pkg/response"
	"littleduck-backend/internal/service"
)

// CommentController 留言 controller
type CommentController struct {
	svc service.CommentService
}

// NewCommentController 建立 CommentController
func NewCommentController(svc service.CommentService) *CommentController {
	return &CommentController{svc: svc}
}

// List 公開讀取可見留言
//
//	@Summary		公開留言列表
//	@Description	取得可見（visible）留言，可用 articleId 過濾；pending／hidden 不列入
//	@Tags			comments
//	@Produce		json
//	@Param			articleId	query		string	false	"文章 id；留空為全部可見留言"
//	@Success		200			{object}	response.Envelope{data=[]model.Comment}
//	@Router			/api/v1/comments [get]
func (ctrl *CommentController) List(c *gin.Context) {
	comments, err := ctrl.svc.ListVisible(c.Query("articleId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, comments)
}

// Create 訪客張貼留言（預設 pending）
//
//	@Summary		張貼留言
//	@Description	訪客張貼留言，建立為待審核（pending）狀態。對應功能開關（留言板 guestbook／文章留言 article_comments）停用時回 403 FEATURE_DISABLED。
//	@Description	可選帶管理員 Authorization: Bearer；權杖有效時略過功能開關（管理員預覽），無效或過期的權杖視同訪客、不回 401。
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					false	"選用：Bearer <管理員 JWT>"
//	@Param			body			body		service.CommentInput	true	"留言內容"
//	@Success		201				{object}	response.Envelope{data=model.Comment}
//	@Failure		400				{object}	response.Envelope
//	@Failure		403				{object}	response.Envelope	"功能已停用（FEATURE_DISABLED）"
//	@Failure		500				{object}	response.Envelope
//	@Router			/api/v1/comments [post]
func (ctrl *CommentController) Create(c *gin.Context) {
	var in service.CommentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperror.BadRequest("留言內容格式錯誤"))
		return
	}
	_, asAdmin := c.Get(middleware.ContextKeyAdmin)
	comment, err := ctrl.svc.Create(in, asAdmin)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, comment)
}

// AdminList 後台讀取全部留言
//
//	@Summary		後台留言列表
//	@Description	取得全部留言（含 pending／visible／hidden），需管理員身分
//	@Tags			comments
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]model.Comment}
//	@Failure		401	{object}	response.Envelope
//	@Router			/api/v1/admin/comments [get]
func (ctrl *CommentController) AdminList(c *gin.Context) {
	comments, err := ctrl.svc.ListAll()
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, comments)
}

// statusRequest 變更狀態請求
type statusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateStatus 後台變更留言狀態
//
//	@Summary		變更留言狀態
//	@Description	將留言狀態改為 pending／visible／hidden，需管理員身分
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"留言 id"
//	@Param			body	body		statusRequest	true	"目標狀態"
//	@Success		200		{object}	response.Envelope{data=model.Comment}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/api/v1/comments/{id}/status [put]
func (ctrl *CommentController) UpdateStatus(c *gin.Context) {
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.BadRequest("請提供目標狀態"))
		return
	}
	comment, err := ctrl.svc.SetStatus(c.Param("id"), req.Status)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, comment)
}

// replyRequest 回覆請求
type replyRequest struct {
	Reply string `json:"reply"`
}

// Reply 後台回覆留言
//
//	@Summary		回覆留言
//	@Description	對留言新增管理員回覆，需管理員身分
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"留言 id"
//	@Param			body	body		replyRequest	true	"回覆內容"
//	@Success		200		{object}	response.Envelope{data=model.Comment}
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/api/v1/comments/{id}/reply [put]
func (ctrl *CommentController) Reply(c *gin.Context) {
	var req replyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.BadRequest("回覆內容格式錯誤"))
		return
	}
	comment, err := ctrl.svc.Reply(c.Param("id"), req.Reply)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, comment)
}
