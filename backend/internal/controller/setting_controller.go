// backend/internal/controller/setting_controller.go
// Setting controller：公開讀取、後台更新網站設定

package controller

import (
	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/pkg/response"
	"littleduck-backend/internal/service"
)

// SettingController 網站設定 controller
type SettingController struct {
	svc service.SettingService
}

// NewSettingController 建立 SettingController
func NewSettingController(svc service.SettingService) *SettingController {
	return &SettingController{svc: svc}
}

// Get 公開讀取網站設定
//
//	@Summary		讀取網站設定
//	@Description	取得前台渲染所需的站台設定（標題、標語、社群連結、公告、贊助級距等）
//	@Tags			settings
//	@Produce		json
//	@Success		200	{object}	response.Envelope{data=model.SiteSetting}
//	@Failure		404	{object}	response.Envelope
//	@Router			/api/v1/settings [get]
func (ctrl *SettingController) Get(c *gin.Context) {
	setting, err := ctrl.svc.Get()
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, setting)
}

// Update 後台更新網站設定
//
//	@Summary		更新網站設定
//	@Description	更新站台設定，需管理員身分；更新後立即反映於公開讀取
//	@Tags			settings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		model.SiteSetting	true	"網站設定"
//	@Success		200		{object}	response.Envelope{data=model.SiteSetting}
//	@Failure		400		{object}	response.Envelope	"標題為空，或社群連結網址格式不正確"
//	@Failure		401		{object}	response.Envelope
//	@Router			/api/v1/settings [put]
func (ctrl *SettingController) Update(c *gin.Context) {
	var in model.SiteSetting
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperror.BadRequest("設定內容格式錯誤"))
		return
	}
	setting, err := ctrl.svc.Update(in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, setting)
}
