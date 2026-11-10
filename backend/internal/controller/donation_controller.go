// backend/internal/controller/donation_controller.go
// Donation controller：後台唯讀贊助紀錄清單

package controller

import (
	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/response"
	"littleduck-backend/internal/service"
)

// DonationController 贊助紀錄 controller
type DonationController struct {
	svc service.DonationService
}

// NewDonationController 建立 DonationController
func NewDonationController(svc service.DonationService) *DonationController {
	return &DonationController{svc: svc}
}

// List 後台讀取贊助紀錄
//
//	@Summary		贊助紀錄列表
//	@Description	取得贊助／付款紀錄清單，需管理員身分（唯讀）
//	@Tags			donations
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]model.Donation}
//	@Failure		401	{object}	response.Envelope
//	@Router			/api/v1/donations [get]
func (ctrl *DonationController) List(c *gin.Context) {
	donations, err := ctrl.svc.List()
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, donations)
}
