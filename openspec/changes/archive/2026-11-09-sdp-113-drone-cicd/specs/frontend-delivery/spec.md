# Spec Delta

## Purpose

定義 littleduck_website 前端的持續部署：由 Drone 觸發 Vercel Preview 部署作為 Staging，人工核准後才部署到 Production 網域。

## ADDED Requirements

### Requirement: 部署只由 Drone 觸發
前端部署 SHALL 只由 Drone 的 CD pipeline 觸發；Vercel 的 Git 自動部署 MUST 關閉，推送 PR 分支不會產生 Vercel 部署。

#### Scenario: PR 分支推送
- **WHEN** 推送一個非 `main` 的分支
- **THEN** Vercel 沒有新的部署

### Requirement: main 推送後部署 Preview 並冒煙測試
推送到 `main` 時，CD SHALL 建置前端並部署到 Vercel Preview，取得該次 Preview 網址後對首頁發請求；回應非 2xx MUST 判定失敗。

#### Scenario: Preview 正常
- **WHEN** 一個 commit 合併進 `main`
- **THEN** 建置記錄印出該次 Preview 網址，首頁回 200

### Requirement: Production 需人工核准
前端 Production 部署 MUST 只在使用者 promote 一個 Preview 已通過的建置到 `production` 時發生，部署後 SHALL 對正式網域跑冒煙測試。

#### Scenario: 核准後部署
- **WHEN** 使用者 promote 該建置到 `production`
- **THEN** 正式網域改為該 commit 的版本，首頁回 200

#### Scenario: 未核准
- **WHEN** Preview 通過但沒有人 promote
- **THEN** 正式網域維持原版本
