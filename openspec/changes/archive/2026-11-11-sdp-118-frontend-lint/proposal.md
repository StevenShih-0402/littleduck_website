# Proposal

## Why

前端（`frontend/`，Next.js 16＋React 19，全部是 `.js`／`.jsx`／`.mjs`）沒有 ESLint，也沒有型別檢查（SDP-118）。結果是 Drone 的 `ci-frontend` 沒有 lint 門檻；agent-team 的前端 lint hook 找不到 ESLint 設定，永遠不會作用；Teshima「檢查」的前端 lint／typecheck 只能記成「專案未設定」。前端 Agent 產出的程式碼目前只靠 `next build` 把關。

## What Changes

- **ESLint**：在 `frontend/` 導入 `eslint` 與 `eslint-config-next`（flat config `eslint.config.mjs`，採 `core-web-vitals`），`package.json` 加 `lint` 腳本，直接呼叫 ESLint CLI（Next 16 已移除 `next lint`）。既有程式碼的錯誤一律修正，或以帶理由的行內註解逐行豁免；不關閉、不降級規則。
- **型別檢查**：採 `jsconfig.json` 開 `checkJs`，搭配 `tsc --noEmit`；不轉 TypeScript（user 決定）。加入 `typescript` 與 React／Node 型別定義為 devDependency，`package.json` 加 `typecheck` 腳本。既有型別錯誤修正，或以帶理由的 `// @ts-expect-error` 豁免。
- **CI**：`.drone.yml` 的 `ci-frontend` 在 `install` 之後、`build` 之前加入 `lint` 與 `typecheck` 兩步；`docs/cicd.md` 的前端門檻表同步更新。
- **agent-team 文件**：`CLAUDE.md`「檢查指令」前端欄的 lint／型別檢查改成實際指令。lint hook 的說明改成前端會作用。「審查規則三分法」的機械性規則補上前端 ESLint。

與 issue 原文的差異：

- issue 寫的 Yuma／Osamu 已在 SDP-88 改名為 Fujimori／Teshima。`.claude/agents/fujimori.md`、`teshima.md` 沒有「專案未設定」字樣，它們依 `CLAUDE.md`「檢查指令」取指令，所以**只改 `CLAUDE.md`，不改 `.claude/agents/**`**。
- 前端 lint hook（`.claude/skills/frontend-processor/hooks/post-edit-lint.sh`）已改為從被編輯檔往上找最近、含 ESLint 設定的目錄。設定放在 `frontend/eslint.config.mjs` 就會生效，**不需設 `LITTLEDUCK_FE_LINT_CMD`，也不改 hook**。

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `ci-pipeline`：「前端 CI 門檻」加入 lint 與型別檢查，兩步都在正式建置之前。
- `frontend-project-structure`：新增「前端提供 lint 與型別檢查腳本」需求，包括零錯誤基準與豁免規則。

## Impact

- **程式碼**：`frontend/package.json`、`package-lock.json`、新增 `frontend/eslint.config.mjs`，以及 `frontend/jsconfig.json`。既有 `frontend/**/*.{js,jsx,mjs}` 可能因修正 lint 與型別錯誤而小幅改動，使用者可見的行為不變。
- **相依**：新增 devDependency `eslint`、`eslint-config-next`、`typescript`、`@types/react`、`@types/react-dom`、`@types/node`。執行期依賴不變。`npm audit --audit-level=high` 仍須通過。
- **CI／CD**：`.drone.yml`（屬高風險檔案樣式，合併前要 user 看過 diff）、`docs/cicd.md`。
- **流程文件**：`CLAUDE.md`。
- **API／後端**：無。沒有 `openapi.yaml` 變動。
