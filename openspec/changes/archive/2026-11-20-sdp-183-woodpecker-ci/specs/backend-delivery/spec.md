# Spec Delta

## MODIFIED Requirements

### Requirement: main 推送後產出並推送映像
推送到 `main` 時，CD SHALL 直接建置後端映像（不先重跑 CI 檢查），通過弱點掃描後推送到 GHCR，tag 為 `<version>-<git 短 sha>`、`<version>` 與 `latest`；`<version-sha>` tag MUST 是不可變的部署依據。每個 commit 上線用的映像 MUST 只建置一次。

#### Scenario: 推送映像
- **WHEN** 一個 commit 合併進 `main`
- **THEN** GHCR 上出現該 commit 對應的三個 tag，且指向同一個 digest

#### Scenario: 映像有高風險弱點
- **WHEN** 映像含 CRITICAL 或 HIGH 等級且未登記於 `backend/.trivyignore` 的弱點
- **THEN** CD 失敗，映像不推送，Staging 不變

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
