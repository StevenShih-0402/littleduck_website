# Design

## Context

- `frontend/` 有 34 個 `.js`／`.jsx`／`.mjs` 檔，約 2.7k 行。`jsconfig.json` 只設了 `baseUrl` 與 `@/*` paths。devDependency 只有 Tailwind 相關套件，沒有 `eslint` 也沒有 `typescript`。
- Next 16 已移除 `next lint`。官方做法見 `frontend/node_modules/next/dist/docs/01-app/03-api-reference/05-config/03-eslint.md`：`eslint` 加 `eslint-config-next`，flat config 用 `defineConfig([...nextVitals, globalIgnores([...])])`，再以 `eslint .` 執行。
- 前端 lint hook（`.claude/skills/frontend-processor/hooks/post-edit-lint.sh`）會從被編輯檔往上找最近、含 `eslint.config.*` 的目錄，而且該目錄的 `node_modules/.bin/eslint` 要存在才會跑。所以設定檔放在 `frontend/` 就會生效。worktree 需要先在 `frontend/` 執行 `npm ci`。
- `frontend/design-sync/manifest.json` 記錄的是**設計來源檔**的雜湊，前端檔只是 `targets`。因此為了 lint 修改前端檔，不影響 `design:check` 的基準。

## Goals / Non-Goals

**Goals:**

- `npm run lint`、`npm run typecheck` 在主幹上結束碼為 0，CI 擋得住新的違規。
- 前端 Agent 編輯 JS 檔時，lint hook 實際執行。

**Non-Goals:**

- 不轉成 TypeScript，不開 `strict`（之後如果要收緊，另開 issue）。
- 不導入 Prettier 或其他格式化工具，不加 `--fix` 到 CI。
- 不改 lint hook 本身與 `.claude/agents/**`。

## Decisions

1. **ESLint 設定採 `eslint-config-next/core-web-vitals`**：官方推薦，涵蓋 Next、React、React Hooks 規則，而且把影響 Core Web Vitals 的規則升為 error。另一個選擇是只用 base config，但規則較鬆，而 issue 要求「不要靠降低規則讓它過」，所以選 vitals。`globalIgnores` 只保留官方預設（`.next/**`、`out/**`、`build/**`、`next-env.d.ts`）。
2. **`lint` 腳本為 `eslint .`**：涵蓋 `app/`、`components/`、`store/`、`lib/`、`data/`、`scripts/` 與根目錄設定檔。不加 `--max-warnings 0`：warning 不擋 CI，以免把 vitals 的 warn 規則變相升為 error。這是刻意的選擇，日後要收緊再調整。
3. **型別檢查採 checkJs 加 `tsc --noEmit -p jsconfig.json`**（user 已決定）：`.jsx` 不改名，design-sync 的 `targets` 對應不受影響。`jsconfig.json` 補上 `allowJs`、`checkJs`、`noEmit`、`jsx: "preserve"`、`module: "esnext"`、`moduleResolution: "bundler"`、`target: "ES2022"`、`lib: ["dom", "dom.iterable", "esnext"]`、`skipLibCheck: true`，並以 `include` 限定原始碼目錄，保留既有的 `baseUrl` 與 `paths`。不開 `strict`。另一個選擇是轉成 TS，但影響面大，而且 design-pull 的移植對象是 `.jsx`，所以不採用。
4. **加入 `typescript`、`@types/react`、`@types/react-dom`、`@types/node`**（`typescript` 同時是 `eslint-config-next` 內 `typescript-eslint` 的 peer 依賴，沒裝 ESLint 會直接無法載入）：沒有這些型別定義，`tsc` 會把 JSX 與 React import 報成錯誤或推成 `any`，型別檢查就沒有意義。版本對齊目前的 React 19 與 Node 20（CI image）。
5. **既有違規的處理順序**：先修正；只有修正會改變使用者可見行為時，才用行內註解豁免並寫明理由。豁免的規則依 spec「既有違規只能逐行豁免」。
   - 實作結果（2026-11-11）：導入時有 16 個 error。`react-hooks/static-components` 8 個，把 `Kpi` 移到模組層修正。`react-hooks/set-state-in-effect` 中「props 同步到表單 state」的 3 個，改用 render 期間比對前值的寫法修正。「掛載後才讀 localStorage」的 5 個（字級、後台外觀、首頁檢視、主題），經 user 決定以單行 `eslint-disable-next-line … -- 理由` 豁免：這是避免 SSR hydration 不一致的標準寫法，改用 `useSyncExternalStore` 會讓首頁不再保存依主題推導的預設檢視。9 個 `@next/next/no-img-element` warning 維持現狀（不擋 CI）。
   - 型別錯誤 5 個，都不靠 `@ts-expect-error`：`<img onError>` 改用 `e.currentTarget`，選填 props 以 JSDoc 標 `?`。
6. **CI 步驟**：`.drone.yml` 的 `ci-frontend` 在 `install` 之後依序加入 `lint`（`npm run lint`）與 `typecheck`（`npm run typecheck`），兩步都用 `node:20-alpine`，排在 `test-design` 與 `build` 之前，讓靜態錯誤最早失敗。

## Risks / Trade-offs

- [`eslint-config-next` 內建的 react-hooks 新規則（如 `set-state-in-effect`）可能在既有程式碼報出大量錯誤] → 逐一修正。修正會動到行為時改用行內豁免，並在 PR 說明列出豁免清單。
- [安裝 `typescript` 後，`next build` 可能自動產生 `tsconfig.json` 或 `next-env.d.ts`] → 實作時確認 build 後 `git status` 沒有多出檔案。有的話，在 `jsconfig.json` 調整或於 `.gitignore` 處理，並在 handoff 註明。
- [新增 devDependency 帶進 high 以上弱點，讓 `npm audit` 失敗] → 選用最新的穩定版本。仍有弱點時停下回報，不降低門檻（依 `docs/cicd.md`）。
- [之後 `/design-pull` 移植設計時，又帶入 lint 錯誤] → 前端從此有 lint 門檻，design-pull 完成前要跑 `npm run lint`、`npm run typecheck`。本變更在 `CLAUDE.md`「設計同步」補一行提醒。
- [`.drone.yml` 屬高風險檔案樣式] → 合併前要 user 看過 diff 並回覆「合併」。

## Migration Plan

純開發工具與 CI 變更，不影響部署產物。回滾方式是 revert 本 PR：移除 CI 步驟與 devDependency 即可。

## 執行驗證觀察點

- 在 `frontend/` 執行 `npm run lint` → 結束碼 0
- 在 `frontend/` 執行 `npm run typecheck` → 結束碼 0
- 在 `frontend/` 執行 `npm run build` 後，`git status --porcelain frontend/` 沒有新增或修改的檔案
- 把 `{"tool_input":{"file_path":"<worktree>/frontend/components/ui.jsx"}}` 餵給 `.claude/skills/frontend-processor/hooks/post-edit-lint.sh` → 實際執行 eslint（暫時寫入一個違規時 exit 2，還原後 exit 0）
- 暫時寫入一個 lint 違規，執行本機 `drone exec --pipeline ci-frontend`（指令見 `docs/cicd.md`「在本機跑 CI」）→ 在 `lint` 步驟失敗

## 變異測試目標

無（本變更是工具與 CI 設定，沒有業務邏輯模組）。
