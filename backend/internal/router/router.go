// backend/internal/router/router.go
// Gin engine 組裝：掛載 middleware、註冊路由（health、公開 /api/v1、受保護 /api/v1）

package router

import (
	"errors"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "littleduck-backend/docs"
	"littleduck-backend/internal/config"
	"littleduck-backend/internal/controller"
	"littleduck-backend/internal/middleware"
	"littleduck-backend/internal/repository"
	"littleduck-backend/internal/service"
	"littleduck-backend/internal/service/auth"
	"littleduck-backend/internal/service/payment"
)

// gormPinger 將 *gorm.DB 適配為 controller.Pinger
type gormPinger struct {
	db *gorm.DB
}

// Ping 檢查底層資料庫連線
func (p gormPinger) Ping() error {
	if p.db == nil {
		return errors.New("資料庫未初始化")
	}
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// New 建立掛好 middleware 與路由的 Gin engine
func New(cfg *config.Config, log *zap.Logger, db *gorm.DB) *gin.Engine {
	gin.SetMode(cfg.Server.GinMode)

	r := gin.New()
	r.Use(
		middleware.Recovery(log),
		middleware.RequestLogger(log),
		middleware.ErrorHandler(log),
		middleware.CORS(),
	)

	healthCtrl := controller.NewHealthController(gormPinger{db: db})
	r.GET("/health", healthCtrl.Check)

	// 各層組裝
	authSvc := auth.NewService(cfg.Auth)
	authCtrl := controller.NewAuthController(authSvc)
	articleCtrl := controller.NewArticleController(
		service.NewArticleService(repository.NewArticleRepository(db)),
	)
	commentCtrl := controller.NewCommentController(
		service.NewCommentService(repository.NewCommentRepository(db)),
	)
	settingCtrl := controller.NewSettingController(
		service.NewSettingService(repository.NewSettingRepository(db)),
	)
	statCtrl := controller.NewStatController(
		service.NewStatService(repository.NewStatRepository(db)),
	)
	donationCtrl := controller.NewDonationController(
		service.NewDonationService(repository.NewDonationRepository(db)),
	)
	paymentCtrl := controller.NewPaymentController(payment.NewECPayService(cfg.ECPay))

	// 公開路由群組（免驗證）
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", authCtrl.Login)

		v1.GET("/articles", articleCtrl.List)
		v1.GET("/articles/:id", articleCtrl.Get)

		v1.GET("/comments", commentCtrl.List)
		v1.POST("/comments", commentCtrl.Create)

		v1.GET("/settings", settingCtrl.Get)

		v1.POST("/stats/pageview", statCtrl.RecordPageview)

		v1.POST("/payments", paymentCtrl.Create)
		v1.POST("/payments/ecpay/callback", paymentCtrl.ECPayCallback)
	}

	// 受保護路由群組（需管理員 JWT）
	admin := r.Group("/api/v1")
	admin.Use(middleware.Auth(authSvc))
	{
		admin.GET("/admin/articles", articleCtrl.AdminList)
		admin.POST("/articles", articleCtrl.Create)
		admin.PUT("/articles/:id", articleCtrl.Update)
		admin.DELETE("/articles/:id", articleCtrl.Delete)

		admin.GET("/admin/comments", commentCtrl.AdminList)
		admin.PUT("/comments/:id/status", commentCtrl.UpdateStatus)
		admin.PUT("/comments/:id/reply", commentCtrl.Reply)

		admin.PUT("/settings", settingCtrl.Update)

		admin.GET("/stats", statCtrl.Daily)

		admin.GET("/donations", donationCtrl.List)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
