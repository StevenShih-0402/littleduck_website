# Spec Delta

## MODIFIED Requirements

### Requirement: 保留原型功能行為
重構後 SHALL 保留原型的使用者可見行為：雙主題切換（day 雜誌／night 終端機，選擇持久化於 localStorage）、前台與後台全部功能。文章、留言、設定、統計與贊助紀錄等業務資料 SHALL 以後端 API 作為來源與儲存，不再以 localStorage 作為這些資料的持久層（主題偏好仍持久化於 localStorage）。

#### Scenario: 主題切換與持久化
- **WHEN** 使用者切換主題後重新整理頁面
- **THEN** 主題維持上次選擇，且全站 theme tokens（背景、文字、邊框）正確套用

#### Scenario: 後台編輯反映到前台
- **WHEN** 使用者在 `/admin` 修改文章或站台設定
- **THEN** 變更經後端 API 寫入，前台頁面讀取到更新後的內容，且換裝置或清除瀏覽器資料後仍一致
