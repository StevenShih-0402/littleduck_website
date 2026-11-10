// backend/internal/seed/seed_data.go
// 自前端 frontend/data/site.js（SITE）與 frontend/store/lake.js（seed）忠實移植的種子來源，
// 以 Go 重現同一套確定性推導（文章 id／title／cat／status／blurb／body 與字面 settings／comments／donations）。

package seed

import "littleduck-backend/internal/model"

// litArticle 對齊 SITE.literature.articles 的單筆
type litArticle struct {
	T     string // 標題
	Y     int    // 年份
	Blurb string // 摘要
	Priv  bool   // 是否私藏（draft 且不上架作品展示）
}

// litAward 對齊 SITE.literature.awards 的單筆（t→sub 映射）
type litAward struct {
	T   string
	Sub string
}

// progProject 對齊 SITE.programming.projects 的單筆
type progProject struct {
	T     string
	Stack string
	Blurb string
	Link  string
}

// siteBrand 等字面值對齊 SITE 頂層欄位
const (
	siteBrand   = "小鴨ㄉ創作湖畔"
	siteBrandEn = "Little Duck's Creative Lakeside"
	siteOwner   = "Steven SHIH"
	siteMotto   = "Do what you can, and make it the best you can."
)

// literatureAwards 對齊 SITE.literature.awards
var literatureAwards = []litAward{
	{T: "緣分", Sub: "第 20 屆高雄馭墨三城聯合文學獎・小說組"},
	{T: "角落的我們", Sub: "第 21 屆高雄馭墨三城聯合文學獎・小說組"},
	{T: "湛藍狂想", Sub: "第 36 屆淡江大學五虎崗文學獎・小說組"},
	{T: "追逐", Sub: "第 9 屆淡江大學秋水文章文學創作大賽・最短篇小說組"},
	{T: "逆向行駛", Sub: "第 37 屆淡江大學五虎崗文學獎・小說組"},
	{T: "到底", Sub: "第 38 屆淡江大學五虎崗文學獎・小說組"},
	{T: "烏雲蔽日", Sub: "第 39 屆淡江大學五虎崗文學獎・散文組"},
	{T: "紅塵記事", Sub: "第 39 屆淡江大學五虎崗文學獎・小說組"},
}

// literatureArticles 對齊 SITE.literature.articles（順序不可變）
var literatureArticles = []litArticle{
	{T: "紅塵記事", Y: 2023, Blurb: "新燐高中棒球隊，沒有留下太多成果、沒有得到太多關注，紅土上卻依然留下了他們奔馳的足跡。"},
	{T: "烏雲蔽日", Y: 2023, Blurb: "以商管學院學生視角，回顧對他們至關重要的畢業專題，現實的因素、團隊的紛爭、時間與制度的壓力一步步戳破美好幻想。", Priv: true},
	{T: "到底", Y: 2022, Blurb: "彌勒佛大頭、追求『平』的平哥、白面仔，他們對牌桌上的第四個人感到陌生，三人逐漸發現各自都產生了記憶錯亂……"},
	{T: "逆向行駛", Y: 2021, Blurb: "北漂青年梓晟自己開車回去高雄。在 4 小時、300 公里的路程上，20 年前的回憶讓他想起早該被忘卻的童年傷疤……"},
	{T: "藍衣小男孩", Y: 2021, Blurb: "藍祥樺總是穿著藍色外套，是個無庸置疑的實作天才，但他特立獨行的古怪行徑，卻使他被班上同學當成笑話看待……"},
	{T: "夢醒", Y: 2020, Blurb: "夢境中，他拼命在泥沼中搶下那顆果實；現實中，他拼命在腦海中維繫那份夢想，然而等他回過神來，卻發現自己早已身陷泥沼。"},
	{T: "追逐", Y: 2020, Blurb: "曾經是體育選手的女孩，長年在深山叢林的棕熊，毫無相似之處的兩個生命，如今卻相遇了。"},
	{T: "湛藍狂想", Y: 2020, Blurb: "升上大學後，璇慈感覺自己不再像以前一樣自在，本應變化多端的藍，卻逐漸安定在鬱悶和煩惱……"},
	{T: "肥宅與魔法師", Y: 2020, Blurb: "一百五十公斤的肥宅、單身三十年的魔法師、連標題和封面都遺忘的邊緣人，人生失敗三人組——直到情人節的網聚邀請。"},
	{T: "夜明", Y: 2019, Blurb: "本來只存在於《JOJO 的奇妙冒險》的「替身」，居然出現在號稱「北高一中」的新莊高中；身為唯一沒有替身的阿愷，該如何在這所獨特的學校存活？"},
	{T: "角落的我們", Y: 2019, Blurb: "聒噪的王冠軒、遊戲宅莊子洋、完美主義者黃鈺婷、絕緣人方宇辰…曾經以為標籤就能反映自己的洞察，但戶外報告後才第一次認識到這些人……", Priv: true},
	{T: "緣分", Y: 2018, Blurb: "邊緣人阿明原以為高中生活會在孤寂中度過，一場意料之外的班遊，卻讓他注意到那些自己未曾留意過的緣分……", Priv: true},
}

// programmingProjects 對齊 SITE.programming.projects（順序不可變）
var programmingProjects = []progProject{
	{T: "簡易社群媒體系統", Stack: "Vue.js + Spring Boot", Blurb: "獨立開發；註冊登入、瀏覽貼文、留言互動、新增/編輯/刪除，分頁與無限滾動。", Link: "GitHub 連結"},
	{T: "商業大樓停車位預約系統", Stack: "Vue + Spring Boot + Trello", Blurb: "公司內部協作開發，簡化每週臨時車位申請流程；GitHub 程式審核與版控。", Link: "GitHub 連結 ×2"},
	{T: "電影購票系統 (Spring Boot 練習)", Stack: "Spring Boot", Blurb: "公司培訓金融專案，RESTful API、列表查詢、非同步查詢、Unit Test 與 Swagger 文件。", Link: "GitHub 連結"},
	{T: "《ANON TOKYO》LineBot 梗圖機器人", Stack: "Python + Line API + ChatGPT", Blurb: "MyGO!!!!! 動畫梗圖機器人，3000+ 張截圖，編號 / 關鍵字查詢，支援樂團原創歌曲搜尋。", Link: "下載 / 說明 / 製作分享"},
	{T: "巴哈姆特動畫瘋作品受眾分析", Stack: "Python + Power BI", Blurb: "爬取 1475 部動畫名稱、時間、集數與觀看數，三頁視覺化分析：時間、屬性、排行榜。"},
	{T: "2015–2023 中華職棒選秀會分析", Stack: "Power BI", Blurb: "5 頁分析：時間軸、年份/學校/守位、二次參選對比。"},
	{T: "112年1月高雄市交通事故儀表板", Stack: "Power BI", Blurb: "大四商業智慧實務期末作品，從政府開放資料設計維度資料表並視覺化。"},
}

// deterministicRnd 對齊 lake.js 的 makeRnd：線性同餘法產生 [0,1) 偽亂數
type deterministicRnd struct{ x int64 }

func newRnd(seed int64) *deterministicRnd { return &deterministicRnd{x: seed} }

func (r *deterministicRnd) next() float64 {
	r.x = (r.x*1103515245 + 12345) % 2147483648
	return float64(r.x) / 2147483648
}

// mathRound 對齊 JS Math.round（0.5 向正無窮進位）
func mathRound(f float64) int {
	return int(f + 0.5)
}

// buildArticles 重現 lake.js 的 articles 陣列（a0 在前，a1..a12，p1..p7；a1 補 gallery）
func buildArticles() []model.Article {
	awardMap := map[string]string{}
	for _, a := range literatureAwards {
		awardMap[a.T] = a.Sub
	}

	rndV := newRnd(7)

	articles := make([]model.Article, 0, 1+len(literatureArticles)+len(programmingProjects))

	// 小說 → 文章
	lit := make([]model.Article, 0, len(literatureArticles))
	for i, a := range literatureArticles {
		status := model.ArticleStatusPublished
		if a.Priv {
			status = model.ArticleStatusDraft
		}
		lit = append(lit, model.Article{
			ID:         "a" + itoa(i+1),
			Title:      a.T,
			Cat:        "小說",
			Date:       itoa(a.Y) + " / " + padZero((i*3)%12+1) + " / " + padZero((i*7)%26+2),
			Y:          a.Y,
			Status:     status,
			Views:      120 + mathRound(rndV.next()*1400),
			Blurb:      a.Blurb,
			Award:      awardMap[a.T],
			Banner:     "",
			BannerFade: false,
			Showcase:   !a.Priv,
			WorkCat:    "小說作品",
			Gallery:    []string{},
			Body:       "> " + a.Blurb + "\n\n（全文整理中，之後由後台補上。）",
		})
	}

	// 程式專案 → 文章
	proj := make([]model.Article, 0, len(programmingProjects))
	for i, p := range programmingProjects {
		y := 2024 - i/4
		award := p.Stack
		body := "> " + p.Blurb + "\n\n**技術棧**：" + orDash(p.Stack)
		if p.Link != "" {
			body += "\n\n**連結**：" + p.Link
		}
		body += "\n\n（專案介紹整理中，之後由後台補上。）"
		proj = append(proj, model.Article{
			ID:         "p" + itoa(i+1),
			Title:      p.T,
			Cat:        "技術",
			Date:       itoa(y) + " / " + padZero((i*2)%12+1) + " / " + padZero((i*5)%26+3),
			Y:          y,
			Status:     model.ArticleStatusPublished,
			Views:      60 + mathRound(rndV.next()*700),
			Blurb:      p.Blurb,
			Award:      award,
			Banner:     "",
			BannerFade: false,
			Showcase:   true,
			WorkCat:    "程式專案",
			Gallery:    []string{},
			Body:       body,
		})
	}

	// a0 隨筆置於最前
	essay := model.Article{
		ID:         "a0",
		Title:      "湖畔改版記事：從 Google Sites 到自己的小屋",
		Cat:        "隨筆",
		Date:       "2026 / 06 / 21",
		Y:          2026,
		Status:     model.ArticleStatusPublished,
		Views:      213,
		Blurb:      "記錄這次網站 2.0 改版的想法：為什麼想要一個結合部落格與作品展示的個人小屋。",
		Award:      "",
		Banner:     "/assets/duck-mini.jpg",
		BannerFade: true,
		Showcase:   false,
		WorkCat:    "",
		Gallery:    []string{},
		Body:       "## 為什麼改版\n\n舊網站是 Google Sites 拼起來的，展示可以，但**寫東西**很痛苦。\n\n這次想要的是：\n\n- 一個能隨手發文的部落格\n- 一個能好好陳列專案與小說的展示區\n- 一個自己的後台，管文章、看數據\n\n![舊站截圖（待補）]()\n\n> 鴨子在湖面上看似悠哉漂泊，實則在底下奮力划水。\n\n之後的隨筆、技術筆記都會發在這裡。",
	}

	articles = append(articles, essay)
	articles = append(articles, lit...)
	articles = append(articles, proj...)

	// a1 補示範封面圖
	for i := range articles {
		if articles[i].ID == "a1" {
			articles[i].Gallery = []string{"/assets/duck-mini.jpg", "", ""}
		}
	}

	return articles
}

// buildSetting 重現 lake.js seed.settings 字面值
func buildSetting() model.SiteSetting {
	return model.SiteSetting{
		ID:           model.SiteSettingID,
		Title:        siteBrand,
		TitleEn:      siteBrandEn,
		Owner:        siteOwner,
		Aka:          "站長小鴨",
		Tagline:      "Programmer × Novelist",
		Subtitle:     "程式與小說，是成本最低的『創作』。",
		Motto:        siteMotto,
		Announcement: "小屋 2.0 開張 — 留言板已開放，歡迎簽到。",
		Social: []model.SocialLink{
			{Label: "Instagram", URL: "#"},
			{Label: "GitHub", URL: "#"},
			{Label: "巴哈姆特", URL: "#"},
			{Label: "Medium", URL: "#"},
		},
		CatsBlog:  []string{"小說", "隨筆", "技術"},
		CatsWorks: []string{"程式專案", "小說作品"},
		Sponsor: model.Sponsor{
			Title:    "請小鴨喝杯蜂蜜檸檬",
			Subtitle: "人要吃飯才能創作、網站要錢才能運作。",
			Tiers: []model.SponsorTier{
				{Amt: 20, Label: "Claude 訂閱費"},
				{Amt: 60, Label: "一杯飲料"},
				{Amt: 120, Label: "一份便當"},
			},
		},
	}
}

// buildDonations 重現 lake.js seed.donations 字面值
func buildDonations() []model.Donation {
	return []model.Donation{
		{ID: "d1", Date: "2026 / 06 / 05", Amt: 60, Name: "路過的水鳥", Note: "新家開張賀！"},
		{ID: "d2", Date: "2026 / 06 / 12", Amt: 120, Name: "夜貓讀者", Note: "敲碗《到底》續篇"},
		{ID: "d3", Date: "2026 / 06 / 18", Amt: 20, Name: "匿名訪客", Note: ""},
		{ID: "d4", Date: "2026 / 06 / 24", Amt: 200, Name: "MyGO 同好", Note: "ANON TOKYO 太好用"},
		{ID: "d5", Date: "2026 / 06 / 29", Amt: 60, Name: "潛水讀者", Note: ""},
		{ID: "d6", Date: "2026 / 07 / 01", Amt: 20, Name: "匿名訪客", Note: "加油"},
	}
}

// buildComments 重現 lake.js seed.comments 字面值；c2 的 articleId 依「到底」文章連動
func buildComments(articles []model.Article) []model.Comment {
	daoDiID := ""
	for _, a := range articles {
		if contains(a.Title, "到底") {
			daoDiID = a.ID
			break
		}
	}
	return []model.Comment{
		{ID: "c1", Name: "路過的水鳥", Date: "2026 / 06 / 28", Text: "從巴哈小屋追過來的，新家很讚，簽到！", Status: model.CommentStatusVisible, Reply: "歡迎入住，記得常來。", ArticleID: ""},
		{ID: "c2", Name: "夜貓讀者", Date: "2026 / 06 / 25", Text: "《到底》的結尾看了三遍，後勁很強。想問有沒有出版的打算？", Status: model.CommentStatusVisible, Reply: "謝謝！那篇結尾改了七次。出版還很遙遠，先在湖邊慢慢寫。", ArticleID: daoDiID},
		{ID: "c3", Name: "MyGO 同好", Date: "2026 / 06 / 19", Text: "ANON TOKYO 每天都在用，關鍵字查梗圖太方便了，感謝大大。", Status: model.CommentStatusVisible, Reply: "", ArticleID: ""},
		{ID: "c4", Name: "匿名訪客", Date: "2026 / 06 / 30", Text: "請問接不接網站案子？", Status: model.CommentStatusPending, Reply: "", ArticleID: ""},
		{ID: "c5", Name: "潛水讀者", Date: "2026 / 07 / 01", Text: "改版記事看完了，新家很有味道，期待技術筆記！", Status: model.CommentStatusVisible, Reply: "會努力更新的！", ArticleID: "a0"},
	}
}

// orDash 空字串回傳 "—"（對齊 lake.js p.stack || "—"）
func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
