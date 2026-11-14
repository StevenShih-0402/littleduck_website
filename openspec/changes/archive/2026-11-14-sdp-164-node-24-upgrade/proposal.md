# Proposal

## Why

前端 CI 與前端映像目前都用 `node:20-alpine`。Node 20 已於 2026-04-30 停止支援，不再有安全修補（SDP-157 PR 審查備註 #2）。Node 24 是目前的 Active LTS，這次把 littleduck_website 中所有用到 Node.js 的地方統一升到 Node 24（SDP-164）。

## What Changes

- `.drone.yml` 的 `ci-frontend` pipeline：`install`、`lint`、`typecheck`、`test-design`、`build`、`audit` 六個步驟的映像由 `node:20-alpine` 改為 `node:24-alpine`。
- `frontend/Dockerfile`：`deps`、`build`、`runner` 三個 stage 的基底映像由 `node:20-alpine` 改為 `node:24-alpine`，所以 Staging／Production 的前端容器也跑 Node 24。
- `frontend/package.json`：`@types/node` 由 `^20` 升到 `^24`（同步更新 `package-lock.json`）；新增 `engines.node: ">=24"`，宣告支援的 Node 版本。
- 新增 `frontend/.nvmrc`（內容 `24`），讓本機開發用 nvm／fnm 時切到同一個大版本。
- 範圍只限本 repo。其他 repo 也有 Node 版本設定（`batting_record` 的 Firebase Functions `engines.node: "20"`、`luke` 的 `node:22-alpine`），本次不改，只寫進 Linear 分析，建議各自另開 issue。

不含破壞性 API 變更；後端（Go）不受影響。

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `frontend-project-structure`：新增一條需求，規定前端的 Node.js 執行環境版本：CI、映像、`engines` 與 `.nvmrc` 一律用 Node 24 LTS，且不得使用已停止支援的 Node 版本。

## Impact

- **檔案**：`.drone.yml`、`frontend/Dockerfile`、`frontend/package.json`、`frontend/package-lock.json`、新增 `frontend/.nvmrc`。
- **風險分級**：`.drone.yml` 與 `**/Dockerfile` 屬於 CLAUDE.md「高風險檔案樣式」中的 CI／CD 與部署類，整個變更判定為**高風險**，合併前要等 user 審批。
- **部署**：合併後 CD 會用 Node 24 重建前端映像，再部署到 Staging；Production 要等人工 promote。`runner` 的健康檢查用到 `wget`，`node:24-alpine` 的 busybox 一樣提供。
- **相依**：Next.js 16.4.5 需要 `node >=20.9.0`，Node 24 符合；其餘套件的 `engines` 都沒有上限。
- **本機開發**：目前開發機的 Node 是 20.20.0。在 Node 20 上執行 `npm ci` 只會出現 `EBADENGINE` 警告，不會失敗（專案沒有開 `engine-strict`）。建議開發機改用 Node 24，但本 issue 不處理。
