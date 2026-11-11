# Tasks

## 1. ESLint

- [x] 1.1 [FE] 在 `frontend/` 以 `npm i -D eslint eslint-config-next` 加入依賴，建立 `eslint.config.mjs`（`core-web-vitals`，只用官方預設的 `globalIgnores`），在 `package.json` 加 `"lint": "eslint ."`。驗證：`npm run lint` 能執行，並列出既有違規。
- [x] 1.2 [FE] 修正既有 lint 違規。無法在不改變行為下修正的，以帶理由的單行 `eslint-disable-next-line` 豁免。驗證：`npm run lint` 結束碼 0，`grep -rn "eslint-disable" frontend --include=*.js --include=*.jsx --include=*.mjs --exclude-dir=node_modules --exclude-dir=.next` 的每一則都是單行且附理由。

## 2. 型別檢查

- [x] 2.1 [FE] 加入 `typescript`、`@types/react`、`@types/react-dom`、`@types/node` 為 devDependency。依 design 決策 3 更新 `jsconfig.json`（保留 `baseUrl`、`paths`，不開 `strict`），在 `package.json` 加 `"typecheck": "tsc --noEmit -p jsconfig.json"`。驗證：`npm run typecheck` 能執行，並列出既有錯誤。
- [x] 2.2 [FE] 修正既有型別錯誤。無法修正的，以帶理由的 `// @ts-expect-error` 豁免。驗證：`npm run typecheck` 結束碼 0。
- [x] 2.3 [FE] 確認加入 typescript 後建置不產生多餘檔案。驗證：`npm run build` 成功，且 `git status --porcelain frontend/` 沒有新增 `tsconfig.json`、`next-env.d.ts` 等檔案（有的話依 design 風險處理並記錄）；`npm run test:design` 通過；`npm audit --audit-level=high` 結束碼 0。

## 3. CI

- [x] 3.1 [FE] 在 `.drone.yml` 的 `ci-frontend` 的 `install` 之後、`test-design` 之前，加入 `lint`、`typecheck` 兩步（`node:20-alpine`，`cd frontend && npm run lint`／`npm run typecheck`）。在 `docs/cicd.md`「前端（`ci-frontend`）」表格補上這兩列，並更新 `.drone.yml` 檔頭註解（如有提到階段）。驗證：本機 `drone exec --pipeline ci-frontend`（依 `docs/cicd.md`「在本機跑 CI」）全綠；暫時寫入一個 lint 違規後再跑，在 `lint` 步驟失敗，還原後恢復全綠。

## 4. agent-team 文件與 hook 驗證

- [x] 4.1 [FE] 更新 `CLAUDE.md`：「檢查指令」表前端欄的 lint 改為 `npm run lint`、型別檢查改為 `npm run typecheck`。lint hook 那行改成「前端 hook 會在 `frontend/` 已 `npm ci` 時對編輯的 JS 檔跑 ESLint；後端 hook 只處理 `.py`，`gofmt`／`go vet` 仍要自己跑」。「審查規則三分法」機械性那列補上前端 ESLint（`npm run lint`）。「設計同步」補一行：`/design-pull` 移植後要通過 `npm run lint`、`npm run typecheck`。驗證：`grep -n "專案未設定" CLAUDE.md` 沒有前端 lint／typecheck 的結果。
- [x] 4.2 [FE] 驗證前端 lint hook 會作用：把 `{"tool_input":{"file_path":"<絕對路徑>/frontend/components/ui.jsx"}}` 以 stdin 餵給 `.claude/skills/frontend-processor/hooks/post-edit-lint.sh`。驗證：暫時寫入一個 lint 違規時 exit 2，並輸出 eslint 訊息；還原後 exit 0。

## Workflow follow-up

- `.drone.yml` 屬高風險檔案樣式，PR 合併前要 user 看過 diff。
- 依 `/issue-solver` 歸檔：`ci-pipeline` 與 `frontend-project-structure` 的 delta 同步到主規格。
