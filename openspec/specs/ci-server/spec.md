# ci-server Specification

## Purpose
自架的 CI server（Woodpecker server 與 agent）：負責接收 GitHub webhook、執行 littleduck_website 的流水線，並把結果回報到 GitHub。

## Requirements

### Requirement: CI server 以容器部署並對外提供 HTTPS
CI server（Woodpecker server）與 agent SHALL 以 docker compose 部署在 Hetzner 主機上，並 MUST 只透過 HTTPS 子網域 `ci.stevenshih-littleduck.com` 對外；server 與 agent 容器 MUST NOT 直接對外開埠，agent 只在內部網路以 gRPC 連 server。

#### Scenario: 透過子網域存取
- **WHEN** 使用者以瀏覽器開啟 `https://ci.stevenshih-littleduck.com`
- **THEN** 看到 Woodpecker 登入頁，且憑證有效

#### Scenario: 未經代理的連線被擋下
- **WHEN** 從外部直接連主機上 Woodpecker server 的 HTTP 埠（8000）或 gRPC 埠（9000）
- **THEN** 連線失敗

### Requirement: 只允許指定的 GitHub 帳號登入與啟用 repo
CI server MUST 以 GitHub OAuth 登入，且只允許設定中列出的 GitHub 帳號登入與啟用 repo；不在清單內的帳號 MUST NOT 能註冊。

#### Scenario: 非授權帳號登入
- **WHEN** 不在允許清單內的 GitHub 帳號嘗試登入
- **THEN** Woodpecker 拒絕登入

### Requirement: 推送與 PR 事件觸發建置
已啟用的 repo 收到推送或 PR 事件時，CI server SHALL 依該 repo 的 `.woodpecker/` 設定建立建置，並把每個 workflow 的結果以 GitHub commit status 回報，context 格式為 `ci/woodpecker/<event>/<workflow>`。

#### Scenario: PR 觸發
- **WHEN** 已啟用的 repo 開啟或更新 PR
- **THEN** Woodpecker 建立建置，PR 上出現 `ci/woodpecker/pr/ci-backend` 與 `ci/woodpecker/pr/ci-frontend` 兩個 check，狀態隨建置結束更新為成功或失敗

### Requirement: 祕密不寫進 repo
部署用祕密（SSH 私鑰、GHCR token、各環境 env）MUST 存在 CI server 的 repo secret 或主機上的 env 檔，MUST NOT 出現在 repo 或建置記錄的明文中；部署用 secret MUST 只開放給 `push` 與 `deployment` 事件。

#### Scenario: 建置記錄遮蔽祕密
- **WHEN** 流水線步驟引用 secret 並把它印出
- **THEN** 建置記錄顯示遮蔽後的值

#### Scenario: PR 建置拿不到部署祕密
- **WHEN** 由 PR 觸發的 CI workflow 執行
- **THEN** 該 workflow 取不到部署用的 secret

### Requirement: fork 來的 PR 需人工核准才執行
由 fork 開到本 repo 的 PR，CI server MUST 在 repo 擁有者核准後才執行建置；核准前 MUST NOT 在掛有主機 docker.sock 的 agent 上執行任何步驟。

#### Scenario: fork PR 待核准
- **WHEN** 其他帳號從 fork 開 PR 到 `main`
- **THEN** Woodpecker 上該建置停在等待核准，沒有任何步驟容器被建立

### Requirement: repo 只取得必要的 Trusted 權限
本 repo 在 CI server 的 Trusted 設定 SHALL 只開啟 volumes（流水線需要掛載主機 docker.sock 與快取目錄）；network 與 security（privileged）MUST 維持關閉。

#### Scenario: 檢查 repo 設定
- **WHEN** 管理員在 Woodpecker 檢視本 repo 的 Trusted 設定
- **THEN** 只有 volumes 為開啟

#### Scenario: 要求 privileged 的步驟
- **WHEN** 某個 workflow 的步驟設定 `privileged: true`
- **THEN** 建置被拒絕執行

### Requirement: 舊 CI server 停用後不再回報
切換完成後，Drone MUST NOT 再收到本 repo 的 webhook，也 MUST NOT 再對本 repo 的 commit 回報 status。

#### Scenario: 切換後推送
- **WHEN** 切換完成後有 commit 推送到 `main`
- **THEN** 該 commit 只有 `ci/woodpecker/...` 的 status，沒有 `continuous-integration/drone/...`，且 GitHub repo 的 webhook 清單沒有指向 Drone 的項目
