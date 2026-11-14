# ci-server Specification

## Purpose
自架的 Drone CI server 與 runner：負責接收 GitHub webhook、執行 littleduck_website 的流水線，並把結果回報到 GitHub。

## Requirements

### Requirement: Drone server 以容器部署並對外提供 HTTPS
Drone server 與 Docker runner SHALL 以 docker compose 部署在 Hetzner 主機上，並 MUST 只透過 HTTPS 子網域對外；server 容器 MUST NOT 直接對外開埠。

#### Scenario: 透過子網域存取
- **WHEN** 使用者以瀏覽器開啟 `https://drone.<網域>`
- **THEN** 看到 Drone 登入頁，且憑證有效

#### Scenario: 未經代理的連線被擋下
- **WHEN** 從外部直接連主機上 Drone server 容器的埠
- **THEN** 連線失敗

### Requirement: 只允許指定的 GitHub 帳號登入與啟用 repo
Drone server MUST 以 GitHub OAuth 登入，且只允許設定中列出的 GitHub 帳號登入與啟用 repo。

#### Scenario: 非授權帳號登入
- **WHEN** 不在允許清單內的 GitHub 帳號嘗試登入
- **THEN** Drone 拒絕登入

### Requirement: 推送與 PR 事件觸發建置
已啟用的 repo 收到推送或 PR 事件時，Drone SHALL 依該 repo 的 `.drone.yml` 建立建置，並把每個 pipeline 的結果以 GitHub commit status 回報。

#### Scenario: PR 觸發
- **WHEN** 已啟用的 repo 開啟或更新 PR
- **THEN** Drone 建立建置，PR 上出現 Drone 的 check，狀態隨建置結束更新為成功或失敗

### Requirement: 祕密不寫進 repo
部署用祕密（SSH 私鑰、GHCR token、各環境 env）MUST 存在 Drone 的 repo secret 或主機上的 env 檔，MUST NOT 出現在 repo 或建置記錄的明文中。

#### Scenario: 建置記錄遮蔽祕密
- **WHEN** 流水線步驟引用 secret 並把它印出
- **THEN** Drone 的建置記錄顯示遮蔽後的值

#### Scenario: PR 建置拿不到部署祕密
- **WHEN** 由 PR 觸發的 CI pipeline 執行
- **THEN** 該 pipeline 取不到部署用的 secret
