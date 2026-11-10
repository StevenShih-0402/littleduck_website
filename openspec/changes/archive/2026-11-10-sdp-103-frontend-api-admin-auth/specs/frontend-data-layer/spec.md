# Spec Delta

## Purpose

定義前端資料層改接後端 API 的行為：透過 API client 以非同步方式讀寫後端資料、處理載入中與錯誤狀態，並在後台登入後保存與附帶存取權杖。

## ADDED Requirements

### Requirement: 前端資料經由後端 API 存取
前端 SHALL 透過集中的 API client 以 HTTP 呼叫後端取得與變更文章、留言、設定、統計與贊助紀錄，MUST NOT 再以瀏覽器 localStorage 作為這些資料的來源或儲存。主題偏好仍可持久化於 localStorage。

#### Scenario: 前台內容來自後端
- **WHEN** 訪客開啟前台頁面
- **THEN** 文章、留言與設定由後端 API 取得並渲染，而非讀取 localStorage 的 `lakeside_v2_store`

#### Scenario: 後台變更寫回後端
- **WHEN** 管理員在後台新增或修改文章、留言狀態或設定
- **THEN** 變更經 API 寫入後端，重新整理或換裝置後仍一致

### Requirement: 處理非同步載入與錯誤狀態
前端 SHALL 在資料尚未取得時呈現載入狀態，並在 API 失敗時顯示可理解的錯誤而非整頁崩潰或顯示假資料。

#### Scenario: 載入中
- **WHEN** 頁面正在等待 API 回應
- **THEN** 顯示載入指示，資料到達後才渲染內容

#### Scenario: API 失敗
- **WHEN** API 呼叫失敗（網路錯誤或非 2xx）
- **THEN** 顯示錯誤提示，不以假資料填補、頁面不崩潰

### Requirement: 後台登入保存與附帶權杖
前端後台 SHALL 以後端 `auth` API 登入並保存回傳的存取權杖，對受保護端點的請求 MUST 附帶該權杖；移除「任意密碼即可登入」的原型行為。權杖缺失或失效時 MUST 導回登入。

#### Scenario: 登入成功後進入後台
- **WHEN** 管理員輸入正確帳密並登入成功
- **THEN** 前端保存權杖並進入後台，後續管理請求帶該權杖

#### Scenario: 帳密錯誤
- **WHEN** 管理員輸入錯誤帳密
- **THEN** 顯示登入失敗，且不進入後台

#### Scenario: 權杖失效
- **WHEN** 已登入狀態下權杖過期或被拒（401）
- **THEN** 前端清除權杖並導回登入畫面
