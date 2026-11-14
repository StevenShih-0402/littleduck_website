## MODIFIED Requirements

### Requirement: 自動部署到 Staging
映像推送成功後，CD SHALL 把同一個 `<version-sha>` 映像部署到 Hetzner 上的 Staging，並在 120 秒內等到 api 容器健康，否則 MUST 判定失敗。Staging 對外主機名 SHALL 為 `api.staging.stevenshih-littleduck.com`。

#### Scenario: Staging 部署成功
- **WHEN** 映像推送成功
- **THEN** Staging 改跑該映像，`https://api.staging.stevenshih-littleduck.com/health` 回 200

#### Scenario: 健康檢查逾時
- **WHEN** api 容器 120 秒內沒有變成 healthy
- **THEN** CD 失敗，記錄中印出容器狀態，不開放 promote 到 Production

### Requirement: Production 需人工核准
Production 部署 MUST 只在使用者對一個 Staging 已通過的建置執行 promote（目標 `production`）時發生；沒有 promote 就不部署 Production。Production 對外主機名 SHALL 為 `api.stevenshih-littleduck.com`。

#### Scenario: 未核准
- **WHEN** Staging 通過但沒有人 promote
- **THEN** Production 維持原版本

#### Scenario: 核准後部署
- **WHEN** 使用者 promote 該建置到 `production`
- **THEN** Production 改跑與 Staging 相同的 `<version-sha>` 映像，健康後對 `https://api.stevenshih-littleduck.com` 跑冒煙測試
