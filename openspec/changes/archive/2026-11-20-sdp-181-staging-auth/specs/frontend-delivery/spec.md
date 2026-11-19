# Spec Delta

## MODIFIED Requirements

### Requirement: 自動部署到前端 Staging 並冒煙測試
映像推送成功後，CD SHALL 把同一個 `<version>-<git 短 sha>` 映像部署到 Hetzner 上的前端 Staging（`staging.stevenshih-littleduck.com`），120 秒內等到容器健康後才更新 Nginx 站台，接著以 Staging 帳密對 Staging 冒煙：首頁、`/blog` 與同源的 `/api/v1/articles` 回 2xx，`/runtime-config.js` 指向前端 Staging 自己 `https://staging.stevenshih-littleduck.com`，且不帶帳密請求首頁回 401。任一項不符 MUST 判定失敗。

#### Scenario: Staging 部署成功
- **WHEN** 映像推送成功
- **THEN** 帶帳密請求 `https://staging.stevenshih-littleduck.com/` 回 200，`/api/v1/articles` 回 2xx，`/runtime-config.js` 的 `apiBaseUrl` 是 `https://staging.stevenshih-littleduck.com`

#### Scenario: 健康檢查逾時
- **WHEN** 前端容器 120 秒內沒有變成 healthy
- **THEN** CD 失敗，記錄中印出容器狀態與記錄，Nginx 站台不更新

#### Scenario: 後端位址設錯
- **WHEN** 主機 `.env` 的 `API_BASE_URL` 不是 `https://staging.stevenshih-littleduck.com`
- **THEN** 冒煙失敗，CD 失敗

#### Scenario: 驗證未生效
- **WHEN** 不帶帳密請求 `https://staging.stevenshih-littleduck.com/` 得到 401 以外的狀態碼
- **THEN** 冒煙失敗，CD 失敗
