# backend-delivery Specification

## Purpose
定義 littleduck_website 後端的持續部署：從 `main` 產出映像，經 Hetzner Staging 驗證與人工核准後送上 Production，並能回滾。

## Requirements

### Requirement: main 推送後產出並推送映像
推送到 `main` 時，CD SHALL 直接建置後端映像（不先重跑 CI 檢查），通過弱點掃描後推送到 GHCR，tag 為 `<version>-<git 短 sha>`、`<version>` 與 `latest`；`<version-sha>` tag MUST 是不可變的部署依據。每個 commit 上線用的映像 MUST 只建置一次。

#### Scenario: 推送映像
- **WHEN** 一個 commit 合併進 `main`
- **THEN** GHCR 上出現該 commit 對應的三個 tag，且指向同一個 digest

#### Scenario: 映像有高風險弱點
- **WHEN** 映像含 CRITICAL 或 HIGH 等級且未登記於 `backend/.trivyignore` 的弱點
- **THEN** CD 失敗，映像不推送，Staging 不變

### Requirement: 自動部署到 Staging
映像推送成功後，CD SHALL 把同一個 `<version-sha>` 映像部署到 Hetzner 上的 Staging，並在 120 秒內等到 api 容器健康，否則 MUST 判定失敗。Staging 對外主機名 SHALL 為 `api.staging.stevenshih-littleduck.com`。

#### Scenario: Staging 部署成功
- **WHEN** 映像推送成功
- **THEN** Staging 改跑該映像，`https://api.staging.stevenshih-littleduck.com/health` 回 200

#### Scenario: 健康檢查逾時
- **WHEN** api 容器 120 秒內沒有變成 healthy
- **THEN** CD 失敗，記錄中印出容器狀態，不開放 promote 到 Production

### Requirement: Staging 冒煙測試
Staging 部署健康後，CD SHALL 以 Staging 帳密對 Staging 的關鍵讀取端點各發一次請求；任一端點非 2xx MUST 判定失敗。CD 另 SHALL 不帶帳密請求文章列表端點，結果不是 401 MUST 判定失敗。

#### Scenario: 端點異常
- **WHEN** 帶帳密請求 Staging 的 `/health` 或文章列表端點，回應非 2xx
- **THEN** CD 失敗

#### Scenario: 驗證未生效
- **WHEN** 不帶帳密請求 Staging 的文章列表端點，回應不是 401
- **THEN** CD 失敗

### Requirement: Production 需人工核准
Production 部署 MUST 只在使用者對一個 Staging 已通過的 push 建置觸發 deployment（目標 `production`）時發生；沒有觸發就不部署 Production。部署前 MUST 確認該 commit 在 GitHub 上所有 `ci/woodpecker/push/*` status 都是成功，否則 MUST 在部署前失敗。Production 對外主機名 SHALL 為 `api.stevenshih-littleduck.com`。

#### Scenario: 未核准
- **WHEN** Staging 通過但沒有人觸發 deployment
- **THEN** Production 維持原版本

#### Scenario: 核准後部署
- **WHEN** 使用者在 Woodpecker 對該 push 建置按 Deploy、目標填 `production`
- **THEN** Production 改跑與 Staging 相同的 `<version-sha>` 映像，健康後對 `https://api.stevenshih-littleduck.com` 跑冒煙測試

#### Scenario: push 建置未全綠
- **WHEN** 使用者以 CLI 或 API 對一個 push 建置有失敗 workflow 的 commit 觸發 deployment
- **THEN** CD 在部署前失敗，Production 不變

#### Scenario: 目標不是 production
- **WHEN** 使用者觸發 deployment 但目標不是 `production`
- **THEN** Production 部署的 workflow 不執行

### Requirement: Staging 與 Production 互相隔離
Staging 與 Production SHALL 各自使用獨立的資料庫與 env 檔，敏感值 MUST NOT 跨環境共用。

#### Scenario: 資料隔離
- **WHEN** 在 Staging 寫入資料
- **THEN** Production 資料庫看不到該筆資料

### Requirement: 可回滾到先前版本
運維者 SHALL 能以既有的 `<version-sha>` tag 把 Production 換回先前版本，不需重新建置。

#### Scenario: 回滾
- **WHEN** 依 runbook 指定前一個 `<version-sha>` 重新啟動 Production
- **THEN** Production 改跑該版本且 `/health` 回 200
