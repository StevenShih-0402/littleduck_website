# Tasks

> 指令都在 repo（或 worktree）根目錄執行，除非註明在 `frontend/`。開發機沒有 Node 24，凡是要求「在 Node 24 上」的驗證，一律走容器（design D5）。
> 任務都標 `[FE]`：`.drone.yml` 只改 `ci-frontend` 的映像 tag，屬於前端 CI；若用多 Agent 流程派工，前端 Agent 的寫入範圍要暫時放寬到 `.drone.yml`。

## 1. 前端版本宣告與相依（frontend/）

- [x] 1.1 [FE] `frontend/package.json`：`@types/node` 改為 `^24`（取當下最新的 24.x），新增 `"engines": { "node": ">=24" }`。接著在 Node 24 容器內重新產生 lockfile：`docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -v "$PWD/frontend:/app" -w /app node:24-alpine npm install --package-lock-only`。驗證：`git diff --stat frontend/package-lock.json` 只動到 `@types/node`、`undici-types` 相關項目與根套件的 `engines`；`grep -c '"@types/node": "\^24' frontend/package.json` 為 `1`。
- [x] 1.2 [FE] 新增 `frontend/.nvmrc`，內容只有一行 `24`（design D3）。驗證：`cat frontend/.nvmrc` 輸出 `24`。
- [x] 1.3 [FE] 在 Node 24 上跑前端全部檢查：`docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -e NEXT_TELEMETRY_DISABLED=1 -v "$PWD/frontend:/app" -w /app node:24-alpine sh -c 'npm ci && npm run lint && npm run typecheck && npm run test:design && npm run build && npm audit --audit-level=high'`。`@types/node` 24 若帶出型別錯誤，就地修正，不加豁免註解（design D4）。驗證：結束碼 0。完成後在本機 Node 20 的 `frontend/` 再跑一次 `npm ci`，確認只有 `EBADENGINE` 警告、沒有失敗。

## 2. 前端映像（frontend/Dockerfile）

- [x] 2.1 [FE] `frontend/Dockerfile` 的 `deps`、`build`、`runner` 三個 `FROM` 都改為 `node:24-alpine`（design D1）。驗證：`grep -c '^FROM node:24-alpine' frontend/Dockerfile` 為 `3`，`grep -c 'node:20' frontend/Dockerfile` 為 `0`；然後跑 design.md「執行驗證觀察點」中的映像相關指令：`node -v` 以 `v24.` 開頭、`User` 為 `node`、`smoke-frontend.sh` 對 3101 的結果為 `exit=0`。驗完停止容器並刪除 `ld-web-check` 映像。
- [x] 2.2 [FE] （實作時新增，design D6）`frontend/Dockerfile` 的 `runner` stage 刪除映像內建的 npm、npx、corepack、yarn，原因是 Node 24 映像的 npm 自帶 undici 等套件，會被 trivy 報出 HIGH CVE。驗證：重建 `ld-web-check` 後跑 design.md「執行驗證觀察點」的映像相關指令，`command -v npm …` 只印 `n=127`、trivy 結果為 `exit=0`、`smoke-frontend.sh` 結果為 `exit=0`。驗完停止容器並刪除映像。

## 3. 前端 CI（.drone.yml）

- [x] 3.1 [FE] `.drone.yml` 的 `ci-frontend`：`install`、`lint`、`typecheck`、`test-design`、`build`、`audit` 六個步驟的 `image` 改為 `node:24-alpine`，其他 pipeline 不動。驗證：`grep -c 'image: node:24-alpine' .drone.yml` 為 `6`，`grep -c 'node:20' .drone.yml` 為 `0`；`docker run --rm -v "$PWD:/w" -w /w drone/cli:1.8.0 lint --trusted .drone.yml` 通過；依 `docs/cicd.md`「在本機跑 CI」以 `--pipeline ci-frontend` 執行，全部步驟成功（含映像建置與 trivy）。跑完依該文件把 `frontend/node_modules`、`frontend/.next` 的擁有者改回來。

## Workflow follow-up

- Linear 分析列出其他 repo 的 Node 版本現況（`batting_record/functions` 的 `engines.node: "20"`、`luke` 的 `node:22-alpine`），建議各自另開 issue；本變更不處理。
- 本變更命中高風險檔案樣式（`.drone.yml`、`**/Dockerfile`），合併前等 user 審批。
- 合併後確認 Drone 的 `cd-frontend-staging` 冒煙通過，再 promote 到 Production。
