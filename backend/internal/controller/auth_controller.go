// backend/internal/controller/auth_controller.go
// Auth controller：後台登入端點，驗證帳密後回傳 JWT

package controller

import (
	"github.com/gin-gonic/gin"

	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/pkg/response"
	"littleduck-backend/internal/service/auth"
)

// AuthController 認證 controller
type AuthController struct {
	svc auth.Service
}

// NewAuthController 建立 AuthController
func NewAuthController(svc auth.Service) *AuthController {
	return &AuthController{svc: svc}
}

// loginRequest 登入請求 body
type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 後台登入
//
//	@Summary		後台登入
//	@Description	以管理員帳密換發 JWT 存取權杖
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	true	"管理員帳密"
//	@Success		200		{object}	response.Envelope{data=auth.Token}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Router			/api/v1/auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.BadRequest("請提供帳號與密碼"))
		return
	}
	token, err := ctrl.svc.Login(req.Username, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, token)
}
