# Spec Delta

## MODIFIED Requirements

### Requirement: Staging 冒煙測試
Staging 部署健康後，CD SHALL 以 Staging 帳密對 Staging 的關鍵讀取端點各發一次請求；任一端點非 2xx MUST 判定失敗。CD 另 SHALL 不帶帳密請求文章列表端點，結果不是 401 MUST 判定失敗。

#### Scenario: 端點異常
- **WHEN** 帶帳密請求 Staging 的 `/health` 或文章列表端點，回應非 2xx
- **THEN** CD 失敗

#### Scenario: 驗證未生效
- **WHEN** 不帶帳密請求 Staging 的文章列表端點，回應不是 401
- **THEN** CD 失敗
