# ci-pipeline Specification

## Purpose
定義 littleduck_website 的持續整合：PR 與推送到 `main` 時，前後端各自跑哪些檢查、何時判定失敗，以及結果如何回報到 GitHub PR。

## Requirements

### Requirement: CI 在 PR 與 main 推送時執行
CI SHALL 在 PR 開啟或更新、以及推送到 `main` 時執行，前端與後端各為一個獨立的 check。CI MUST NOT 推送映像或部署。

#### Scenario: PR 上看得到兩個 check
- **WHEN** 開 PR 到 `main`
- **THEN** `gh pr checks` 列出後端與前端兩個 Drone check

#### Scenario: CI 不部署
- **WHEN** CI 執行完成（不論成敗）
- **THEN** GHCR 與 Hetzner 都沒有因此產生新的映像或部署，CI 建出的映像在結束時被刪除

### Requirement: 後端 CI 門檻
後端 CI SHALL 依序執行格式檢查、`go vet`、單元測試、安全掃描、映像建置與映像弱點掃描；任一步驟失敗 MUST 讓後端 check 失敗並停止後續步驟。

#### Scenario: 測試失敗
- **WHEN** PR 中有一個 Go 測試失敗
- **THEN** 後端 check 失敗，映像建置與掃描不執行

#### Scenario: 映像有高風險弱點
- **WHEN** 建置出的映像含 CRITICAL 或 HIGH 等級且未登記於 `backend/.trivyignore` 的弱點
- **THEN** 後端 check 失敗

#### Scenario: 格式不符
- **WHEN** PR 中有未經 `gofmt` 格式化的 Go 檔
- **THEN** 後端 check 失敗，並在記錄中列出該檔案

### Requirement: 後端測試使用真實 PostgreSQL
後端 CI 的測試步驟 SHALL 能連到一個臨時的 PostgreSQL 16，建置結束後該資料庫 MUST 被移除。

#### Scenario: 需要資料庫的測試
- **WHEN** 測試以 CI 提供的 DB 環境變數連線
- **THEN** 連線成功，建置結束後不殘留資料庫容器

### Requirement: 前端 CI 門檻
前端 CI SHALL 依序執行依賴安裝、lint、型別檢查、設計同步工具測試、正式建置、依賴弱點稽核、映像建置與映像弱點掃描；lint 與型別檢查 MUST 在正式建置之前執行；任一步驟失敗 MUST 讓前端 check 失敗。

#### Scenario: 建置失敗
- **WHEN** PR 讓 `npm run build` 失敗
- **THEN** 前端 check 失敗

#### Scenario: 依賴有高風險弱點
- **WHEN** `npm audit` 回報 high 以上的弱點
- **THEN** 前端 check 失敗

#### Scenario: lint 錯誤
- **WHEN** PR 中有前端檔案違反 ESLint 規則（error 等級）
- **THEN** 前端 check 失敗，記錄中列出違規的檔案與規則，正式建置不執行

#### Scenario: 型別錯誤
- **WHEN** PR 中有前端檔案無法通過型別檢查
- **THEN** 前端 check 失敗，記錄中列出錯誤位置，正式建置不執行

#### Scenario: Dockerfile 壞掉
- **WHEN** PR 讓 `docker build frontend` 失敗
- **THEN** 前端 check 失敗

#### Scenario: 映像有高風險弱點
- **WHEN** 建置出的前端映像含 CRITICAL 或 HIGH 等級且未登記於 `frontend/.trivyignore` 的弱點
- **THEN** 前端 check 失敗
