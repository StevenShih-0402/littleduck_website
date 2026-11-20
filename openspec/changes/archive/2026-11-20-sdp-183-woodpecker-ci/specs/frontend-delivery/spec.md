# Spec Delta

## RENAMED Requirements

- FROM: `### Requirement: 部署只由 Drone 觸發`
- TO: `### Requirement: 部署只由 CI server 的 CD 觸發`

## MODIFIED Requirements

### Requirement: 部署只由 CI server 的 CD 觸發
前端部署 SHALL 只由 Woodpecker 的 CD workflow 觸發，部署目標為 Hetzner 主機上的自架容器；repo MUST NOT 含有任何第三方託管平台（如 Vercel）的部署步驟或設定，推送 PR 分支 MUST NOT 產生任何部署。

#### Scenario: PR 分支推送
- **WHEN** 推送一個非 `main` 的分支或開 PR
- **THEN** GHCR 沒有新的前端映像，Hetzner 上的前端容器沒有變動

#### Scenario: repo 不含 Vercel 部署設定
- **WHEN** 在 `.woodpecker/`、`deploy/`、`docs/`、`frontend/`、根目錄 `README.md`、`.gitignore`、`CLAUDE.md` 搜尋 `vercel`（不分大小寫）
- **THEN** 沒有任何結果

### Requirement: Production 需人工核准
前端 Production 部署 MUST 只在使用者對一個整體通過的 push 建置觸發 deployment（目標 `production`）時發生；部署前 MUST 確認該 commit 所有 `ci/woodpecker/push/*` status 都是成功。Production MUST 部署與 Staging 相同的 `<version>-<git 短 sha>` 映像而不重新建置，健康後 SHALL 對正式網域 `stevenshih-littleduck.com` 跑冒煙測試。

#### Scenario: 核准後部署
- **WHEN** 使用者在 Woodpecker 對該 push 建置按 Deploy、目標填 `production`
- **THEN** `https://stevenshih-littleduck.com` 改跑與 Staging 相同 tag 的映像，首頁回 200，且 `/runtime-config.js` 的 `apiBaseUrl` 是 `https://api.stevenshih-littleduck.com`

#### Scenario: 未核准
- **WHEN** Staging 通過但沒有人觸發 deployment
- **THEN** `https://stevenshih-littleduck.com` 維持原版本

#### Scenario: 建置未全綠
- **WHEN** 使用者對一個 push 建置未全部通過的 commit 觸發 deployment
- **THEN** CD 在部署前失敗，Production 不變

### Requirement: main 推送後產出並推送前端映像
推送到 `main` 時，CD SHALL 直接建置前端映像（不先重跑 CI 檢查）、通過弱點掃描後推送到 GHCR，tag 為 `<version>-<git 短 sha>`、`<version>` 與 `latest`，`<version>` 取自 `frontend/package.json`；`<version>-<git 短 sha>` MUST 是不可變的部署依據。

#### Scenario: 推送映像
- **WHEN** 一個 commit 合併進 `main`
- **THEN** GHCR 的 `littleduck-web` 出現該 commit 對應的三個 tag，且指向同一個 digest

#### Scenario: 映像有高風險弱點
- **WHEN** 映像含 CRITICAL 或 HIGH 等級且未登記於 `frontend/.trivyignore` 的弱點
- **THEN** CD 失敗，映像不推送
