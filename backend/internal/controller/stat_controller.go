// backend/internal/controller/stat_controller.go
// Stat controller：公開瀏覽回報（非阻塞語意）、後台每日統計讀取

package controller

import (
	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/response"
	"littleduck-backend/internal/service"
)

// StatController 統計 controller
type StatController struct {
	svc service.StatService
}

// NewStatController 建立 StatController
func NewStatController(svc service.StatService) *StatController {
	return &StatController{svc: svc}
}

// pageviewRequest 瀏覽回報 body（path 選填）
type pageviewRequest struct {
	Path string `json:"path"`
}

// RecordPageview 公開回報一次瀏覽
//
//	@Summary		回報瀏覽事件
//	@Description	前台頁面載入時回報一次瀏覽；非阻塞語意，成功回 204
//	@Tags			stats
//	@Accept			json
//	@Produce		json
//	@Param			body	body	pageviewRequest	false	"瀏覽路徑"
//	@Success		204
//	@Router			/api/v1/stats/pageview [post]
func (ctrl *StatController) RecordPageview(c *gin.Context) {
	var req pageviewRequest
	_ = c.ShouldBindJSON(&req) // body 選填，忽略綁定錯誤

	// 記錄失敗不阻擋前台渲染，吞掉錯誤只回成功
	_ = ctrl.svc.RecordView(c.ClientIP(), c.Request.UserAgent(), req.Path)
	response.NoContentOK(c)
}

// Daily 後台讀取每日統計
//
//	@Summary		每日統計
//	@Description	取得最近 30 天每日 views／visitors 聚合，需管理員身分
//	@Tags			stats
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]repository.DailyStat}
//	@Failure		401	{object}	response.Envelope
//	@Router			/api/v1/stats [get]
func (ctrl *StatController) Daily(c *gin.Context) {
	rows, err := ctrl.svc.Daily()
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, rows)
}
