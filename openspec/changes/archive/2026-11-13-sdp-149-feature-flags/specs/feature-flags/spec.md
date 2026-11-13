# Spec Delta

## Purpose
定義網站的功能開關：每個前台可操作的功能都對應一個開關，由後台管理員切換；開關未啟用時，一般瀏覽者看不到也無法使用該功能的操作區塊。

## ADDED Requirements

### Requirement: 功能開關登記與初始化
系統 SHALL 以後端程式碼內的登記表定義所有功能開關（key、名稱、說明、預設值）。服務啟動時 MUST 為登記表中資料庫尚未存在的 key 以預設值建立紀錄，且 MUST NOT 覆寫已存在 key 的啟用狀態。資料庫中不在登記表的 key MUST NOT 出現在任何 API 回應。

#### Scenario: 空資料庫啟動
- **WHEN** 服務以空資料庫啟動
- **THEN** 登記表內每個 key 都以其預設值建立紀錄，`sponsor`、`guestbook`、`article_comments` 皆為啟用

#### Scenario: 既有開關值保留
- **WHEN** 管理員已把 `guestbook` 設為停用後重新啟動服務
- **THEN** `guestbook` 仍為停用

#### Scenario: 新增 key 後啟動
- **WHEN** 登記表新增一個預設停用的 key 後，以既有資料庫啟動服務
- **THEN** 該 key 以停用狀態建立，其他 key 的狀態不變

### Requirement: 公開讀取功能開關
系統 SHALL 提供不需驗證的 `GET /api/v1/features`，以統一 envelope 回傳登記表中每個 key 是否啟用的對照表（key → boolean）。

#### Scenario: 讀取開關對照表
- **WHEN** 未帶權杖呼叫 `GET /api/v1/features`
- **THEN** 回傳 200，`data` 為包含登記表所有 key 的物件，值為 `true` 或 `false`

### Requirement: 後台列出與切換功能開關
系統 SHALL 提供需管理員 JWT 的 `GET /api/v1/admin/features`（回傳每個開關的 key、名稱、說明、啟用狀態、更新時間）與 `PUT /api/v1/features/{key}`（以 `{"enabled": boolean}` 切換單一開關）。切換後 MUST 立即反映於公開讀取結果。

#### Scenario: 管理員停用開關
- **WHEN** 管理員以有效權杖對 `PUT /api/v1/features/guestbook` 送出 `{"enabled": false}`
- **THEN** 回傳 200 與更新後的開關，後續 `GET /api/v1/features` 的 `guestbook` 為 `false`

#### Scenario: 未授權切換
- **WHEN** 未帶有效權杖呼叫 `PUT /api/v1/features/guestbook`
- **THEN** 回傳 401，開關狀態不變

#### Scenario: 不存在的 key
- **WHEN** 管理員對不在登記表的 key 呼叫 `PUT /api/v1/features/{key}`
- **THEN** 回傳 404 與統一錯誤格式，不建立任何紀錄

#### Scenario: 缺少 enabled 欄位
- **WHEN** 管理員送出的 body 缺少 `enabled` 或其值不是布林
- **THEN** 回傳 400 與統一錯誤格式，開關狀態不變

### Requirement: 前台依開關顯示操作區塊
前台 SHALL 在開關啟用時才對一般瀏覽者渲染對應的操作區塊：`sponsor` 控制贊助浮動按鈕與視窗、`guestbook` 控制留言板投書表單、`article_comments` 控制文章頁留言表單。開關停用、不存在，或開關讀取失敗時 MUST 視為停用並對一般瀏覽者隱藏該區塊；頁面其他內容（含導覽列「留言板」連結、首頁「簽到留言」按鈕、已公開的留言列表）照常顯示。

#### Scenario: 開關啟用
- **WHEN** `guestbook` 為啟用時訪客開啟 `/guest`
- **THEN** 頁面顯示投書表單與已公開留言

#### Scenario: 開關停用
- **WHEN** `guestbook` 為停用時訪客開啟 `/guest`
- **THEN** 頁面不顯示投書表單，已公開留言與導覽列「留言板」連結照常顯示

#### Scenario: 開關讀取失敗
- **WHEN** `GET /api/v1/features` 失敗而其他公開資料載入成功
- **THEN** 三個受控操作區塊都不顯示，其他內容正常渲染、頁面不顯示全頁錯誤

### Requirement: 管理員預覽停用中的功能
前台 SHALL 在「管理員預覽」狀態下，仍渲染停用中的操作區塊並可實際操作，且該區塊 MUST 帶有可見的「停用中・預覽」標示。管理員預覽狀態的成立條件：同一瀏覽器分頁持有管理員權杖，且已以該權杖成功取得後台功能開關清單；權杖缺失、失效（收到 401）或清單讀取失敗時 MUST NOT 進入預覽狀態。已啟用的區塊在預覽狀態下照常顯示，不加標示。

#### Scenario: 管理員看到停用中的區塊
- **WHEN** 管理員已在後台登入，`guestbook` 為停用，於同一分頁開啟 `/guest`
- **THEN** 頁面顯示投書表單並帶「停用中・預覽」標示

#### Scenario: 管理員試用停用中的留言表單
- **WHEN** 管理員在預覽狀態下於停用中的投書表單送出留言
- **THEN** 留言以 pending 建立，畫面顯示與一般送出相同的成功提示

#### Scenario: 一般瀏覽者不受預覽影響
- **WHEN** 未登入後台的訪客開啟 `/guest`，`guestbook` 為停用
- **THEN** 頁面不顯示投書表單，也不顯示任何預覽標示

#### Scenario: 權杖失效退出預覽
- **WHEN** 管理員權杖過期，前台的後台資料請求收到 401
- **THEN** 前台退出預覽狀態，停用中的區塊被隱藏

#### Scenario: 啟用中的區塊不加標示
- **WHEN** 管理員在預覽狀態下開啟 `/guest`，`guestbook` 為啟用
- **THEN** 投書表單照常顯示，不帶預覽標示

### Requirement: 後台功能開關管理介面
後台 SHALL 提供「功能開關」分頁，列出每個開關的名稱、說明、key 與目前狀態，並可逐一切換；切換成功後前台在同一分頁的狀態立即更新，失敗時顯示錯誤且狀態維持原值。

#### Scenario: 後台切換後前台反映
- **WHEN** 管理員在「功能開關」分頁停用 `sponsor` 後點「看前台」
- **THEN** 前台不顯示贊助浮動按鈕

#### Scenario: 切換失敗
- **WHEN** 切換請求回傳非 2xx
- **THEN** 分頁顯示錯誤訊息，該開關顯示的狀態維持切換前的值
