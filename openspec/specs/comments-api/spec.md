# comments-api Specification

## Purpose
提供留言的後端 API：訪客可張貼留言、前台可讀取公開留言、後台可審核留言狀態與回覆，取代前端僅存在 localStorage 的留言資料。

## Requirements

### Requirement: 訪客張貼留言
系統 SHALL 提供公開的留言張貼端點，接受留言者名稱、內容與所屬對象（如文章 id 或留言板）；新留言 MUST 以待審核（pending）狀態建立，不立即公開。

#### Scenario: 成功張貼留言
- **WHEN** 訪客以合法名稱與內容呼叫留言張貼端點
- **THEN** 回傳 201（或 200）並建立一筆狀態為 pending 的留言

#### Scenario: 留言內容缺漏
- **WHEN** 留言缺少必要欄位（如空白內容）
- **THEN** 回傳 400 與統一錯誤格式，不建立留言

### Requirement: 前台讀取公開留言
系統 SHALL 提供公開讀取端點，僅回傳狀態為可見（visible）的留言，可依對象（文章 id／留言板）過濾；pending 與 hidden 留言 MUST NOT 出現在公開結果。

#### Scenario: 只回傳可見留言
- **WHEN** 未帶權杖呼叫公開留言讀取端點
- **THEN** 回傳 200 且結果僅含 visible 留言

### Requirement: 後台審核與回覆留言
系統 SHALL 提供需管理員身分的端點：列出全部留言（含 pending／visible／hidden）、變更留言狀態、以及對留言新增回覆。

#### Scenario: 核准留言為可見
- **WHEN** 管理員以有效權杖將一則 pending 留言狀態改為 visible
- **THEN** 該留言出現在前台公開留言結果

#### Scenario: 隱藏留言
- **WHEN** 管理員將一則留言狀態改為 hidden
- **THEN** 該留言不再出現在前台公開留言結果

#### Scenario: 回覆留言
- **WHEN** 管理員對一則留言新增回覆內容
- **THEN** 該留言帶有回覆並可於前台隨留言顯示

#### Scenario: 未授權的審核操作
- **WHEN** 未帶有效權杖呼叫留言狀態變更或回覆端點
- **THEN** 回傳 401，且留言不變

### Requirement: 留言顯示日期以台灣時間計算
新留言的顯示日期（`date`，格式 `YYYY / MM / DD`）SHALL 以台灣時間（UTC+8）的當天日期產生，MUST NOT 受伺服器所在時區影響。

#### Scenario: 台灣凌晨張貼留言
- **WHEN** 訪客在台灣時間 2026-11-12 03:00（UTC 2026-11-11 19:00）張貼留言
- **THEN** 回傳的留言 `date` 為 `2026 / 11 / 12`

#### Scenario: 台灣白天張貼留言
- **WHEN** 訪客在台灣時間 2026-11-12 15:00 張貼留言
- **THEN** 回傳的留言 `date` 為 `2026 / 11 / 12`
