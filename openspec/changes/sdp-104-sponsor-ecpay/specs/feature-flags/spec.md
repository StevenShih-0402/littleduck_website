## MODIFIED Requirements

### Requirement: 功能開關登記與初始化
系統 SHALL 以後端程式碼內的登記表定義所有功能開關（key、名稱、說明、預設值）。服務啟動時 MUST 為登記表中資料庫尚未存在的 key 以預設值建立紀錄，且 MUST NOT 覆寫已存在 key 的啟用狀態。資料庫中不在登記表的 key MUST NOT 出現在任何 API 回應。登記表 MUST 包含 `sponsor_payment`（贊助付款（綠界），預設停用）。

#### Scenario: 空資料庫啟動
- **WHEN** 服務以空資料庫啟動
- **THEN** 登記表內每個 key 都以其預設值建立紀錄，`sponsor`、`guestbook`、`article_comments` 皆為啟用，`sponsor_payment` 為停用

#### Scenario: 既有開關值保留
- **WHEN** 管理員已把 `guestbook` 設為停用後重新啟動服務
- **THEN** `guestbook` 仍為停用

#### Scenario: 新增 key 後啟動
- **WHEN** 登記表新增一個預設停用的 key 後，以既有資料庫啟動服務
- **THEN** 該 key 以停用狀態建立，其他 key 的狀態不變

#### Scenario: 既有資料庫補上 sponsor_payment
- **WHEN** 以本變更之前、`sponsor` 為啟用的既有資料庫啟動服務
- **THEN** `sponsor_payment` 以停用建立，`sponsor` 仍為啟用

### Requirement: 前台依開關顯示操作區塊
前台 SHALL 在開關啟用時才對一般瀏覽者渲染對應的操作區塊：`sponsor` 與 `sponsor_payment` **兩者都啟用**時才顯示贊助浮動按鈕與視窗、`guestbook` 控制留言板投書表單、`article_comments` 控制文章頁留言表單。開關停用、不存在，或開關讀取失敗時 MUST 視為停用並對一般瀏覽者隱藏該區塊；頁面其他內容（含導覽列「留言板」連結、首頁「簽到留言」按鈕、已公開的留言列表、付款結果頁）照常顯示。

#### Scenario: 開關啟用
- **WHEN** `guestbook` 為啟用時訪客開啟 `/guest`
- **THEN** 頁面顯示投書表單與已公開留言

#### Scenario: 開關停用
- **WHEN** `guestbook` 為停用時訪客開啟 `/guest`
- **THEN** 頁面不顯示投書表單，已公開留言與導覽列「留言板」連結照常顯示

#### Scenario: 開關讀取失敗
- **WHEN** `GET /api/v1/features` 失敗而其他公開資料載入成功
- **THEN** 所有受控操作區塊都不顯示，其他內容正常渲染、頁面不顯示全頁錯誤

#### Scenario: 贊助付款未啟用
- **WHEN** `sponsor` 啟用、`sponsor_payment` 停用，一般訪客開啟任一前台頁面
- **THEN** 頁面不顯示贊助浮動按鈕

#### Scenario: 贊助兩個開關皆啟用
- **WHEN** `sponsor` 與 `sponsor_payment` 皆啟用，一般訪客開啟任一前台頁面
- **THEN** 頁面顯示贊助浮動按鈕，不帶預覽標示

#### Scenario: 管理員預覽贊助付款
- **WHEN** 管理員在預覽狀態下開啟前台，`sponsor_payment` 停用
- **THEN** 頁面顯示贊助浮動按鈕並帶「停用中・預覽」標示，按「贊助」會實際建立付款並導向綠界
