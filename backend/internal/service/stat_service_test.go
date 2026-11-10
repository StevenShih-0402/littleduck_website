// backend/internal/service/stat_service_test.go
// stat service 單元測試：visitor_key 去識別化與去重、每日聚合與空序列（mock repository，無 DB）

package service

import (
	"net/http"
	"testing"
	"time"

	"littleduck-backend/internal/model"
	"littleduck-backend/internal/pkg/apperror"
	"littleduck-backend/internal/repository"
)

// mockStatRepo 測試用 StatRepository，模擬以 day + visitor_key 聚合
type mockStatRepo struct {
	views     []model.PageView
	aggErr    error
	recordErr error
	sinceArg  string
}

func (m *mockStatRepo) RecordView(v *model.PageView) error {
	if m.recordErr != nil {
		return m.recordErr
	}
	m.views = append(m.views, *v)
	return nil
}

// AggregateSince 以記憶體內 views 重現 GROUP BY day、DISTINCT visitor_key 的語意
func (m *mockStatRepo) AggregateSince(since string) ([]repository.DailyStat, error) {
	m.sinceArg = since
	if m.aggErr != nil {
		return nil, m.aggErr
	}
	type agg struct {
		views    int
		visitors map[string]struct{}
	}
	byDay := map[string]*agg{}
	var order []string
	for _, v := range m.views {
		if v.Day < since {
			continue
		}
		a, ok := byDay[v.Day]
		if !ok {
			a = &agg{visitors: map[string]struct{}{}}
			byDay[v.Day] = a
			order = append(order, v.Day)
		}
		a.views++
		a.visitors[v.VisitorKey] = struct{}{}
	}
	out := make([]repository.DailyStat, 0, len(order))
	for _, d := range order {
		out = append(out, repository.DailyStat{Day: d, Views: byDay[d].views, Visitors: len(byDay[d].visitors)})
	}
	return out, nil
}

// fixedNow 回傳固定時間，使測試不受當前日期影響
func fixedNow() time.Time { return time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC) }

func TestStatRecordAndAggregate_DedupsVisitorPerDay(t *testing.T) {
	// Arrange：同一天同一訪客回報兩次、另一訪客一次
	repo := &mockStatRepo{}
	svc := &statService{repo: repo, now: fixedNow}

	// Act
	_ = svc.RecordView("1.1.1.1", "UA-A", "/")
	_ = svc.RecordView("1.1.1.1", "UA-A", "/blog") // 同訪客（同 ip+ua）
	_ = svc.RecordView("2.2.2.2", "UA-B", "/")
	daily, err := svc.Daily()

	// Assert
	if err != nil {
		t.Fatalf("聚合應成功: %v", err)
	}
	if len(daily) != 1 {
		t.Fatalf("應聚合為單日，得到 %+v", daily)
	}
	if daily[0].Views != 3 {
		t.Errorf("views 應為 3，得到 %d", daily[0].Views)
	}
	if daily[0].Visitors != 2 {
		t.Errorf("visitors 去重後應為 2，得到 %d", daily[0].Visitors)
	}
}

func TestStatRecordView_DoesNotStoreRawIP(t *testing.T) {
	// Arrange
	repo := &mockStatRepo{}
	svc := &statService{repo: repo, now: fixedNow}

	// Act
	_ = svc.RecordView("203.0.113.9", "UA", "/")

	// Assert
	v := repo.views[0]
	if v.VisitorKey == "" || len(v.VisitorKey) != 16 {
		t.Errorf("visitor_key 應為 16 字元雜湊，得到 %q", v.VisitorKey)
	}
	if v.VisitorKey == "203.0.113.9" {
		t.Error("不得以原始 IP 當 visitor_key")
	}
}

func TestStatDaily_EmptyReturnsEmptySeriesNotError(t *testing.T) {
	// Arrange：無任何事件
	repo := &mockStatRepo{}
	svc := &statService{repo: repo, now: fixedNow}

	// Act
	daily, err := svc.Daily()

	// Assert
	if err != nil {
		t.Fatalf("無紀錄時應回成功: %v", err)
	}
	if daily == nil {
		t.Error("應回傳非 nil 空序列")
	}
	if len(daily) != 0 {
		t.Errorf("無紀錄時序列應為空，得到 %+v", daily)
	}
}

func TestStatDaily_WrapsRepositoryError(t *testing.T) {
	// Arrange
	repo := &mockStatRepo{aggErr: apperror.Internal(nil, "boom")}
	svc := &statService{repo: repo, now: fixedNow}

	// Act
	_, err := svc.Daily()

	// Assert
	appErr, ok := apperror.AsAppError(err)
	if !ok || appErr.Status != http.StatusInternalServerError {
		t.Fatalf("repository 失敗應包裝為 500，得到 %v", err)
	}
}
