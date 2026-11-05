// backend/internal/controller/health_controller.go
// 健康檢查端點：回報服務狀態與資料庫連線（ping）狀態

package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Pinger 資料庫連線檢查介面（以 interface 注入方便測試替換）
type Pinger interface {
	Ping() error
}

// HealthController 健康檢查 controller
type HealthController struct {
	db Pinger
}

// NewHealthController 建立 HealthController
func NewHealthController(db Pinger) *HealthController {
	return &HealthController{db: db}
}

// healthStatus 健康檢查回應格式
type healthStatus struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// Check 健康檢查
//
//	@Summary		健康檢查
//	@Description	回報服務與資料庫連線狀態；資料庫異常時回傳 503
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	controller.healthStatus
//	@Failure		503	{object}	controller.healthStatus
//	@Router			/health [get]
func (h *HealthController) Check(c *gin.Context) {
	status := healthStatus{Status: "ok", Database: "up"}
	code := http.StatusOK

	if err := h.db.Ping(); err != nil {
		status.Status = "degraded"
		status.Database = "down"
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, status)
}
