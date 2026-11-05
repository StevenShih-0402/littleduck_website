# Design

## Context

動機見 proposal.md。實測到的現況：

- 設計來源是 Migrated Design artifact（`https://claude.ai/artifact/HycuwGKPu29sQriZLbnPtU`），內容放在 `project/` 下，可用 Artifact 工具的 `list_files`、`read_file` 逐檔讀取。使用者匯出的 HTML 與 live 版本的 8 個原始碼檔完全相同。
- artifact 的 `index.html`、`artifact-type/` 屬於編輯器本身，不是設計內容。`project/uploads/`、`project/scraps/` 與 v1 頁面（`direction*.jsx`、`magazine.jsx`、`terminal.jsx` 等）不是現行站點的來源。
- 調查期間版本號從 `1793884311-c09f` 變成 `1793889427-d763`，而納入比對的檔案內容與匯出檔相同。版本號不能當作內容變更的依據。
- 設計是單檔全域腳本（hash router、UMD React），前端是 Next.js App Router。一個設計檔會對應多個前端檔案，無法機械式覆蓋。
- 前後端已在 SDP-107 併入根 repo。Claude Code 平常從根目錄啟動。
- `frontend/` 目前沒有測試框架。

## Goals / Non-Goals

**Goals:**

- 一個指令就能知道設計改了什麼，並把改動帶進程式碼。
- 基準與程式碼一起版控，從任一 commit 都能知道它對應哪版設計。

**Non-Goals:**

- 自動偵測或排程同步。此 session 環境無法監看 artifact 變更。
- 把程式碼的改動回寫到設計。
- 視覺回歸測試或 Storybook。
- 把 tokens 抽成獨立檔案。現行 CSS 變數已與設計一致，等有實際需求再做。

## Decisions

### 1. 用 Artifact 工具拉取，不用 `/design-sync`

使用者已決定。`/design-sync` 需要 design-system 專案、同步對象是預覽元件庫，且依賴將關閉的 standalone 服務。Artifact 讀取直接對應使用者實際編輯的設計，也已實測可用。

### 2. 基準放在 `frontend/design-sync/manifest.json`

基準描述的是前端程式碼對應哪版設計，所以要和程式碼在同一個 commit，放在 `frontend/` 也最容易找到。

```json
{
  "source": { "artifactUrl": "...", "root": "project/" },
  "syncedVersion": "1793889427-d763",
  "syncedAt": "2026-11-05",
  "files": {
    "project/v2/ui.jsx": { "sha256": "...", "targets": ["components/ui.jsx"] }
  }
}
```

不存設計檔的副本：它們可隨時從 artifact 讀取，存副本只會多一份要維護的資料。代價是差異報告只知道哪個檔變了，要看行級差異得靠舊版內容。見 Risks。

### 3. 比對邏輯做成 Node 腳本，流程放在 skill

雜湊與差異分類是確定性工作，放進 `frontend/scripts/design-sync.mjs` 才能測試。移植需要判斷，由 skill 指引 Claude Code 執行。

- `node scripts/design-sync.mjs check <pulled-dir>`：輸出新增、變更、移除清單與對應前端檔案。
- `node scripts/design-sync.mjs update <pulled-dir> --version <v>`：改寫基準。
- 測試用 Node 內建的 `node --test`，不新增依賴。

納入範圍的判定以 `list_files` 結果為準：`project/v2/**`、`project/assets/**`、`project/data.js`、`project/Personal Site v2.html`。

### 4. skill 放根目錄的 `.claude/skills/design-pull/`

Claude Code 從根目錄啟動，Artifact 也在這個 session 可用，和其他流程 skill 放在一起。命名用 `design-pull` 而非 `design-sync`，避免和內建的 `/design-sync` 混淆。

流程：

1. 讀基準，`list_files` 取得目前檔案與版本。
2. `read_file` 把納入範圍的檔案存到 scratchpad。
3. 執行 `check`。沒有差異就回報並結束。
4. 在 frontend 建工作分支，依 `targets` 移植，保留 Next.js 改寫（清單見 specs）。
5. 執行 `npm run build`。通過後才執行 `update`。
6. 回報改動的檔案與版本，不自動 commit、push。

新增或移除設計檔時，由 skill 在移植時決定新的 `targets`，寫入基準。

### 5. 首次基準直接取用現行設計

調查已確認程式碼與樣式對齊，因此不需要先做一次移植。例外是圖片：實作時發現 `public/assets/` 的 `duck-logo.jpg`、`duck-mini.jpg` 從第一個 commit 起就是佔位圖。基準若直接記錄設計的雜湊，這個差異會被隱藏，所以先把圖片換成設計版本。

### 6. 檔案對應

| 設計檔 | 前端檔案 |
|--------|----------|
| `v2/ui.jsx` | `components/ui.jsx` |
| `v2/front-home.jsx` | `components/front/{front-shell,home-page,guest-page,sponsor-fab}.jsx` |
| `v2/front-content.jsx` | `components/front/{blog-page,article-page,works-page}.jsx` |
| `v2/admin-shell.jsx` | `components/admin/{admin-shell,admin-overview,admin-sponsor,admin-settings}.jsx` |
| `v2/admin-manage.jsx` | `components/admin/{admin-posts,admin-editor,admin-works,admin-comments,status-pill}.jsx` |
| `v2/app.jsx` | `store/lake-context.jsx`、`app/` 路由 |
| `v2/store.js` | `store/lake.js` |
| `data.js` | `data/site.js` |
| `Personal Site v2.html` | `app/globals.css`、`app/layout.jsx` |
| `assets/*` | `public/assets/*` |

## Risks / Trade-offs

- [基準不存舊內容，看不到行級差異] → skill 改為比對新設計檔與目前的前端檔案。這也是移植時實際需要的資訊。
- [使用者在編輯器的修改會自動存檔，可能拉到改到一半的設計] → 只在使用者要求時拉取，並回報拉到的版本。
- [移植靠判斷，可能漏改] → 以 `targets` 限定範圍，用 build 驗證。視覺正確性由使用者檢視，build 通過不代表畫面正確。
- [Migrated Design 的檔案結構未來可能改變] → `check` 會把路徑變動列為新增與移除，不會默默略過。
- [後續 sub-issue 改接 API 後，`store.js` 與 `store/lake.js` 會分歧] → 屆時更新該檔的 `targets` 或移出同步範圍。
- [artifact 內容可能含他人寫入的文字] → 一律當資料處理，見 specs。

## Migration Plan

純新增，不影響現有執行行為。要回復時移除 `frontend/design-sync/`、`frontend/scripts/design-sync*` 與 skill 目錄即可。
