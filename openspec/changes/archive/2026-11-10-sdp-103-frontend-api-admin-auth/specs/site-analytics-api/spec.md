# Spec Delta

## Purpose

提供站台瀏覽統計的後端 API：以輕量瀏覽追蹤記錄每日 views／visitors，並對後台提供每日統計聚合，取代前端以種子亂數產生的假統計資料。

## ADDED Requirements

### Requirement: 記錄瀏覽事件
系統 SHALL 提供輕量瀏覽追蹤：前台頁面瀏覽可回報一筆瀏覽事件，系統按日累計 views 與去重後的 visitors。追蹤 MUST NOT 要求管理員身分，且不得阻擋或拖慢頁面渲染。

#### Scenario: 回報一次瀏覽
- **WHEN** 前台頁面載入時回報一筆瀏覽事件
- **THEN** 當日的 views 計數增加，對應 visitor 納入當日 visitors 統計

### Requirement: 後台讀取每日統計
系統 SHALL 提供需管理員身分的統計端點，回傳一段期間（如最近 30 天）每日的 views 與 visitors 聚合結果；資料 MUST 來自實際記錄的瀏覽事件，而非固定種子假資料。

#### Scenario: 取得每日統計
- **WHEN** 管理員以有效權杖呼叫統計端點
- **THEN** 回傳 200 與每日 views／visitors 序列，反映實際瀏覽紀錄

#### Scenario: 尚無瀏覽紀錄
- **WHEN** 系統尚未累積任何瀏覽事件時呼叫統計端點
- **THEN** 回傳 200 與各日為零（或空序列）的統計，而非假資料

#### Scenario: 未授權讀取統計
- **WHEN** 未帶有效權杖呼叫統計端點
- **THEN** 回傳 401，不回傳統計資料
