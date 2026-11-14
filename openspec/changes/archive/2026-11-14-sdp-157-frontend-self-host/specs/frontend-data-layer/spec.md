## ADDED Requirements

### Requirement: 後端位址於執行期決定
前端呼叫後端時的 base URL SHALL 在執行期決定，優先序為：前端伺服器環境變數 `API_BASE_URL`（經 `/runtime-config.js` 交給瀏覽器）→ 建置期 `NEXT_PUBLIC_API_BASE_URL` → `http://localhost:8080`。只有 `http:`／`https:` 的絕對 URL 視為有效值，無效值 MUST 視同未設定；`/runtime-config.js` 的輸出 MUST 跳脫 `<`、`>`、`&` 與行分隔字元，且 MUST NOT 被快取。

#### Scenario: 執行期有設定
- **WHEN** 前端伺服器以 `API_BASE_URL=https://api.example.com/` 啟動
- **THEN** `/runtime-config.js` 回傳 `apiBaseUrl` 為 `https://api.example.com`（去掉結尾斜線），瀏覽器呼叫 `https://api.example.com/api/v1/...`

#### Scenario: 執行期未設定
- **WHEN** 前端伺服器沒有 `API_BASE_URL`
- **THEN** `/runtime-config.js` 回傳 `apiBaseUrl` 為 `null`，瀏覽器改用建置期值，沒有建置期值時用 `http://localhost:8080`

#### Scenario: 無效值
- **WHEN** `API_BASE_URL` 是 `javascript:alert(1)` 或相對路徑
- **THEN** 視同未設定，`apiBaseUrl` 為 `null`

#### Scenario: 不被快取
- **WHEN** 取 `/runtime-config.js`
- **THEN** 回應標頭含 `Cache-Control: no-store`，`Content-Type` 為 JavaScript
