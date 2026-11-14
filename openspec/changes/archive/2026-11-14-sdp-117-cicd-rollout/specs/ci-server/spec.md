## MODIFIED Requirements

### Requirement: Drone server 以容器部署並對外提供 HTTPS
Drone server 與 Docker runner SHALL 以 docker compose 部署在 Hetzner 主機上，並 MUST 只透過 HTTPS 子網域 `drone.stevenshih-littleduck.com` 對外；server 容器 MUST NOT 直接對外開埠。

#### Scenario: 透過子網域存取
- **WHEN** 使用者以瀏覽器開啟 `https://drone.stevenshih-littleduck.com`
- **THEN** 看到 Drone 登入頁，且憑證有效

#### Scenario: 未經代理的連線被擋下
- **WHEN** 從外部直接連主機上 Drone server 容器的埠
- **THEN** 連線失敗
