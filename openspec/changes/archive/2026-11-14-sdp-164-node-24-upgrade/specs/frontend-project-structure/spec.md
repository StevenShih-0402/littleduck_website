# Spec Delta

## ADDED Requirements

### Requirement: 前端統一使用 Node 24 LTS
前端的 Node.js 執行環境 SHALL 統一為 Node 24。這包括前端 CI 每個用到 Node 的步驟、前端映像的建置與執行階段，以及開發環境的版本宣告（`package.json` 的 `engines.node` 與 `frontend/.nvmrc`）。repo MUST NOT 在前端 CI、前端映像或版本宣告中指定已停止支援的 Node 大版本。`@types/node` 的大版本 MUST 與執行環境的 Node 大版本一致。

#### Scenario: CI 使用 Node 24
- **WHEN** 檢視 `.drone.yml` 的 `ci-frontend` pipeline 中所有以 Node 映像執行的步驟
- **THEN** 每一步都使用 Node 24 映像，沒有任何一步使用 Node 20 或更舊的映像

#### Scenario: 前端容器以 Node 24 執行
- **WHEN** 以 `frontend/Dockerfile` 建置映像後執行 `docker run --rm <image> node -v`
- **THEN** 輸出的版本以 `v24.` 開頭

#### Scenario: 版本宣告一致
- **WHEN** 讀取 `frontend/package.json` 的 `engines.node`、`devDependencies["@types/node"]` 與 `frontend/.nvmrc`
- **THEN** `engines.node` 為 `>=24`、`@types/node` 的大版本為 24、`.nvmrc` 內容為 `24`

#### Scenario: 升級後前端檢查全數通過
- **WHEN** 在 Node 24 環境的 `frontend/` 依序執行 `npm ci`、`npm run lint`、`npm run typecheck`、`npm run test:design`、`npm run build`、`npm audit --audit-level=high`
- **THEN** 每個指令的結束碼都是 0
