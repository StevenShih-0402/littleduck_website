## MODIFIED Requirements

### Requirement: Production 需人工核准
前端 Production 部署 MUST 只在使用者 promote 一個整體通過的建置到 `production` 時發生；Production MUST 部署與 Staging 相同的 `<version>-<git 短 sha>` 映像而不重新建置，健康後 SHALL 對正式網域 `stevenshih-littleduck.com` 跑冒煙測試。

#### Scenario: 核准後部署
- **WHEN** 使用者 promote 該建置到 `production`
- **THEN** `https://stevenshih-littleduck.com` 改跑與 Staging 相同 tag 的映像，首頁回 200，且 `/runtime-config.js` 的 `apiBaseUrl` 是 `https://api.stevenshih-littleduck.com`

#### Scenario: 未核准
- **WHEN** Staging 通過但沒有人 promote
- **THEN** `https://stevenshih-littleduck.com` 維持原版本

#### Scenario: 建置未全綠
- **WHEN** 使用者 promote 一個 push 建置未全部通過的 commit
- **THEN** CD 在部署前失敗，Production 不變

### Requirement: 自動部署到前端 Staging 並冒煙測試
映像推送成功後，CD SHALL 把同一個 `<version>-<git 短 sha>` 映像部署到 Hetzner 上的前端 Staging（`staging.stevenshih-littleduck.com`），120 秒內等到容器健康後才更新 Nginx 站台，接著對 Staging 冒煙：首頁與 `/blog` 回 2xx，且 `/runtime-config.js` 指向後端 Staging `https://api.staging.stevenshih-littleduck.com`。任一項不符 MUST 判定失敗。

#### Scenario: Staging 部署成功
- **WHEN** 映像推送成功
- **THEN** `https://staging.stevenshih-littleduck.com/` 回 200，`/runtime-config.js` 的 `apiBaseUrl` 是 `https://api.staging.stevenshih-littleduck.com`

#### Scenario: 健康檢查逾時
- **WHEN** 前端容器 120 秒內沒有變成 healthy
- **THEN** CD 失敗，記錄中印出容器狀態與記錄，Nginx 站台不更新

#### Scenario: 後端位址設錯
- **WHEN** 主機 `.env` 的 `API_BASE_URL` 不是 `https://api.staging.stevenshih-littleduck.com`
- **THEN** 冒煙失敗，CD 失敗

### Requirement: 前端 Staging 與 Production 互相隔離
前端 Staging 與 Production SHALL 各自使用獨立的主機目錄、compose project、容器名與 `.env`，Staging（`staging.stevenshih-littleduck.com`）MUST 只指向後端 Staging `api.staging.stevenshih-littleduck.com`，Production（`stevenshih-littleduck.com`）MUST 只指向後端正式 `api.stevenshih-littleduck.com`。

#### Scenario: 各自指向對應後端
- **WHEN** 分別取 `https://staging.stevenshih-littleduck.com/runtime-config.js` 與 `https://stevenshih-littleduck.com/runtime-config.js`
- **THEN** 前者回傳 `https://api.staging.stevenshih-littleduck.com`，後者回傳 `https://api.stevenshih-littleduck.com`
