// backend/internal/controller/feature_controller.go
// Feature controller：公開讀取開關對照表、後台清單與切換

package controller

import (
	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/pkg/response"
	"littleduck-backend/internal/service"
)

// FeatureController 功能開關 controller
type FeatureController struct {
	svc service.FeatureService
}

// NewFeatureController 建立 FeatureController
func NewFeatureController(svc service.FeatureService) *FeatureController {
	return &FeatureController{svc: svc}
}

// Public 公開讀取開關對照表
//
//	@Summary		讀取功能開關
//	@Description	取得 {key: 是否啟用} 對照表，只含布林值，不含名稱與說明
//	@Tags			features
//	@Produce		json
//	@Success		200	{object}	response.Envelope{data=map[string]bool}
//	@Router			/api/v1/features [get]
func (ctrl *FeatureController) Public(c *gin.Context) {
	flags, err := ctrl.svc.PublicMap()
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, flags)
}

// AdminList 後台讀取開關清單
//
//	@Summary		後台功能開關清單
//	@Description	依登記表順序回傳全部開關（含名稱、說明、狀態、更新時間），需管理員身分
//	@Tags			features
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]service.FeatureView}
//	@Failure		401	{object}	response.Envelope
//	@Router			/api/v1/admin/features [get]
func (ctrl *FeatureController) AdminList(c *gin.Context) {
	list, err := ctrl.svc.List()
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, list)
}

// featureUpdateRequest 切換請求；enabled 用 *bool 以區分缺欄位與 false
type featureUpdateRequest struct {
	Enabled *bool `json:"enabled"`
}

// Update 後台切換開關
//
//	@Summary		切換功能開關
//	@Description	啟用或停用指定功能，需管理員身分；立即反映於公開讀取
//	@Tags			features
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			key		path		string					true	"開關 key"
//	@Param			body	body		featureUpdateRequest	true	"目標狀態"
//	@Success		200		{object}	response.Envelope{data=service.FeatureView}
//	@Failure		400		{object}	response.Envelope	"缺少 enabled 或型別錯誤"
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope	"未知 key"
//	@Router			/api/v1/features/{key} [put]
func (ctrl *FeatureController) Update(c *gin.Context) {
	var req featureUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.BadRequest("請提供 enabled 布林值"))
		return
	}
	view, err := ctrl.svc.Update(c.Param("key"), req.Enabled)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, view)
}
