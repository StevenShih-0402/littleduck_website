// backend/internal/router/router.go
// Gin engine 組裝：掛載 middleware、註冊路由（health、/api/v1）

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

	articleCtrl := controller.NewArticleController(
		service.NewArticleService(repository.NewArticleRepository(db)),
	)

	paymentCtrl := controller.NewPaymentController(payment.NewECPayService(cfg.ECPay))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/articles", articleCtrl.List)
		v1.POST("/payments", paymentCtrl.Create)
		v1.POST("/payments/ecpay/callback", paymentCtrl.ECPayCallback)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
