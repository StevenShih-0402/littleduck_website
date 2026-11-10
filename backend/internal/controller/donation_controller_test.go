// backend/internal/controller/donation_controller_test.go
// donation controller 單元測試：唯讀清單回 envelope（mock service；未授權由 middleware 層負責）

package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"littleduck-backend/internal/middleware"
	"littleduck-backend/internal/model"
	"littleduck-backend/internal/service"
)

// mockDonationService 測試用 DonationService
type mockDonationService struct {
	donations []model.Donation
}

func (m *mockDonationService) List() ([]model.Donation, error) { return m.donations, nil }

func newDonationRouter(svc service.DonationService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(zap.NewNop()))
	r.GET("/api/v1/donations", NewDonationController(svc).List)
	return r
}

func TestDonationList_ReturnsEnvelope(t *testing.T) {
	// Arrange
	svc := &mockDonationService{donations: []model.Donation{{ID: "d1", Amt: 60, Name: "水鳥"}}}
	r := newDonationRouter(svc)

	// Act
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/donations", nil))

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Data []model.Donation `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Data) != 1 || body.Data[0].ID != "d1" {
		t.Errorf("清單內容錯誤: %+v", body.Data)
	}
}
