// backend/internal/seed/seed_data_test.go
// seed 推導單元測試：驗證移植自 lake.js 的文章／留言／設定／贊助與原型一致（不需 DB）

package seed

import (
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"

	"littleduck-backend/internal/model"
)

func TestBuildArticles_OrderAndIDsMatchLakeJS(t *testing.T) {
	// Act
	articles := buildArticles()

	// Assert：a0 隨筆在最前，接 12 篇小說 a1..a12，再接 7 個專案 p1..p7，共 20 篇
	if len(articles) != 20 {
		t.Fatalf("文章總數應為 20（1 隨筆 + 12 小說 + 7 專案），得到 %d", len(articles))
	}
	if articles[0].ID != "a0" || articles[0].Cat != "隨筆" {
		t.Errorf("第一篇應為 a0 隨筆，得到 %+v", articles[0])
	}
	if articles[1].ID != "a1" || articles[1].Title != "紅塵記事" || articles[1].Cat != "小說" {
		t.Errorf("a1 應為小說《紅塵記事》，得到 %+v", articles[1])
	}
	if articles[13].ID != "p1" || articles[13].Cat != "技術" {
		t.Errorf("p1 應為技術專案，得到 %+v", articles[13])
	}
}

func TestBuildArticles_PrivArticlesAreDraft(t *testing.T) {
	// Act
	articles := buildArticles()
	byID := map[string]model.Article{}
	for _, a := range articles {
		byID[a.ID] = a
	}

	// Assert：priv 的「烏雲蔽日」(a2)、「角落的我們」(a11)、「緣分」(a12) 應為 draft 且不上架
	for _, id := range []string{"a2", "a11", "a12"} {
		if byID[id].Status != model.ArticleStatusDraft {
			t.Errorf("%s 應為 draft，得到 %q", id, byID[id].Status)
		}
		if byID[id].Showcase {
			t.Errorf("%s 私藏不應 showcase", id)
		}
	}
	if byID["a1"].Status != model.ArticleStatusPublished {
		t.Errorf("a1 應為 published，得到 %q", byID["a1"].Status)
	}
}

func TestBuildArticles_A1HasGalleryAndBodyStructure(t *testing.T) {
	// Act
	articles := buildArticles()
	var a1 model.Article
	for _, a := range articles {
		if a.ID == "a1" {
			a1 = a
		}
	}

	// Assert
	if len(a1.Gallery) != 3 || a1.Gallery[0] != "/assets/duck-mini.jpg" {
		t.Errorf("a1 gallery 應為示範封面圖，得到 %+v", a1.Gallery)
	}
	if !strings.HasPrefix(a1.Body, "> ") {
		t.Errorf("小說 body 應以引言開頭，得到 %q", a1.Body)
	}
}

func TestBuildSetting_MatchesLakeJSLiterals(t *testing.T) {
	// Act
	s := buildSetting()

	// Assert
	if s.Title != "小鴨ㄉ創作湖畔" || s.Owner != "Steven SHIH" {
		t.Errorf("設定標題／站長錯誤: %+v", s)
	}
	if len(s.Social) != 4 || s.Social[0].Label != "Instagram" {
		t.Errorf("社群連結錯誤: %+v", s.Social)
	}
	if len(s.Sponsor.Tiers) != 3 || s.Sponsor.Tiers[0].Amt != 20 {
		t.Errorf("贊助級距錯誤: %+v", s.Sponsor.Tiers)
	}
	if s.ID != model.SiteSettingID {
		t.Errorf("設定應為單列 id=%d，得到 %d", model.SiteSettingID, s.ID)
	}
}

func TestBuildComments_SeedMatchesLakeJS(t *testing.T) {
	// Act
	comments := buildComments(buildArticles())

	// Assert：5 則、c4 為 pending、c2 連動到《到底》文章、c5 連動 a0
	if len(comments) != 5 {
		t.Fatalf("留言應為 5 則，得到 %d", len(comments))
	}
	byID := map[string]model.Comment{}
	for _, c := range comments {
		byID[c.ID] = c
	}
	if byID["c4"].Status != model.CommentStatusPending {
		t.Errorf("c4 應為 pending，得到 %q", byID["c4"].Status)
	}
	if byID["c5"].ArticleID != "a0" {
		t.Errorf("c5 應連動 a0，得到 %q", byID["c5"].ArticleID)
	}
	// 「到底」文章是 a3（小說列表第 3 筆），c2 應連動該 id
	if byID["c2"].ArticleID == "" {
		t.Error("c2 應連動到《到底》文章 id，不應為空")
	}
}

func TestBuildDonations_MatchesLakeJS(t *testing.T) {
	// Act
	donations := buildDonations()

	// Assert
	if len(donations) != 6 {
		t.Fatalf("贊助紀錄應為 6 筆，得到 %d", len(donations))
	}
	if donations[0].ID != "d1" || donations[0].Amt != 60 || donations[0].Name != "路過的水鳥" {
		t.Errorf("d1 內容錯誤: %+v", donations[0])
	}
	if donations[3].Amt != 200 {
		t.Errorf("d4 金額應為 200，得到 %d", donations[3].Amt)
	}
}

// stubEnsurer 記錄呼叫次數的 FeatureEnsurer
type stubEnsurer struct {
	calls int
	err   error
}

func (s *stubEnsurer) EnsureDefaults() error { s.calls++; return s.err }

func TestSeedFeatures_CallsEnsureDefaults(t *testing.T) {
	// Arrange
	e := &stubEnsurer{}

	// Act
	err := seedFeatures(e, zap.NewNop())

	// Assert
	if err != nil {
		t.Fatalf("seedFeatures 應成功: %v", err)
	}
	if e.calls != 1 {
		t.Errorf("應呼叫 EnsureDefaults 1 次，得到 %d", e.calls)
	}
}

func TestSeedFeatures_PropagatesError(t *testing.T) {
	// Arrange
	e := &stubEnsurer{err: errors.New("db down")}

	// Act
	err := seedFeatures(e, zap.NewNop())

	// Assert
	if err == nil {
		t.Fatal("EnsureDefaults 失敗應回傳錯誤")
	}
}

func TestSeedFeatures_NilEnsurerIsSkipped(t *testing.T) {
	// Act / Assert
	if err := seedFeatures(nil, zap.NewNop()); err != nil {
		t.Fatalf("未提供 ensurer 應略過: %v", err)
	}
}
