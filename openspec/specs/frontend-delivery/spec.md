# frontend-delivery Specification

## Purpose
定義 littleduck_website 前端的持續部署：由 Drone 建置映像並部署到 Hetzner Staging，人工 promote 同一映像到 Production。

## Requirements

### Requirement: 部署只由 Drone 觸發
前端部署 SHALL 只由 Drone 的 CD pipeline 觸發，部署目標為 Hetzner 主機上的自架容器；repo MUST NOT 含有任何第三方託管平台（如 Vercel）的部署步驟或設定，推送 PR 分支 MUST NOT 產生任何部署。

#### Scenario: PR 分支推送
- **WHEN** 推送一個非 `main` 的分支或開 PR
- **THEN** GHCR 沒有新的前端映像，Hetzner 上的前端容器沒有變動

#### Scenario: repo 不含 Vercel 部署設定
- **WHEN** 在 `.drone.yml`、`deploy/`、`docs/`、`frontend/`、根目錄 `README.md`、`.gitignore`、`CLAUDE.md` 搜尋 `vercel`（不分大小寫）
- **THEN** 沒有任何結果

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

### Requirement: 前端可打包為自架映像
前端 SHALL 提供 Dockerfile，產出以 Next.js standalone 輸出執行的映像；映像 MUST 以非 root 使用者執行、監聽 3000 埠，且 MUST NOT 在建置時寫死後端位址。

#### Scenario: 本機執行映像
- **WHEN** `docker build` 後以 `API_BASE_URL=<位址>` 執行容器
- **THEN** 首頁回 200，`/runtime-config.js` 回傳的 `apiBaseUrl` 等於該位址

#### Scenario: 同一映像換位址
- **WHEN** 以同一個映像、不同的 `API_BASE_URL` 再啟動一個容器
- **THEN** 新容器的 `/runtime-config.js` 回傳新的位址，不需重新建置

### Requirement: main 推送後產出並推送前端映像
推送到 `main` 且前端 CI 通過時，CD SHALL 建置前端映像、通過弱點掃描後推送到 GHCR，tag 為 `<version>-<git 短 sha>`、`<version>` 與 `latest`，`<version>` 取自 `frontend/package.json`；`<version>-<git 短 sha>` MUST 是不可變的部署依據。

#### Scenario: 推送映像
- **WHEN** 一個 commit 合併進 `main`
- **THEN** GHCR 的 `littleduck-web` 出現該 commit 對應的三個 tag，且指向同一個 digest

#### Scenario: 映像有高風險弱點
- **WHEN** 映像含 CRITICAL 或 HIGH 等級且未登記於 `frontend/.trivyignore` 的弱點
- **THEN** CD 失敗，映像不推送

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

### Requirement: 前端可回滾到先前版本
運維者 SHALL 能以既有的 `<version>-<git 短 sha>` tag 把前端 Production 換回先前版本，不需重新建置。

#### Scenario: 回滾
- **WHEN** 依 runbook 指定前一個 `<version>-<git 短 sha>` 重新啟動前端 Production
- **THEN** 前端容器改跑該版本，正式網域首頁回 200
