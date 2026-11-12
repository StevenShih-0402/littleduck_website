## ADDED Requirements

### Requirement: 社群連結網址格式驗證
設定更新端點 SHALL 驗證每個社群連結的網址（`social[].url`），只接受：空字串、`#`（佔位）、站內路徑（以單一 `/` 開頭）、`http://`、`https://`、`mailto:` 開頭的網址（scheme 不分大小寫）。任一網址不符時 MUST 回傳 400 與統一錯誤格式，且設定維持原狀。

#### Scenario: 合法的社群連結網址
- **WHEN** 管理員以有效權杖更新設定，社群連結網址為 `https://github.com/example` 與 `#`
- **THEN** 回傳 200，且後續公開讀取回傳更新後的社群連結

#### Scenario: 不安全的社群連結網址
- **WHEN** 管理員以有效權杖更新設定，其中一個社群連結網址為 `javascript:alert(1)`
- **THEN** 回傳 400 與統一錯誤格式，後續公開讀取的設定與更新前相同
