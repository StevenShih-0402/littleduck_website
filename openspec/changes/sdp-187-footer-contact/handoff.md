<!-- handoff.md 範本：複製到 openspec/changes/<change-id>/handoff.md。「狀態」區塊就地改值、永遠反映最新進度；下方各角色紀錄只追加，不改前面的內容。 -->
# Handoff：sdp-187-footer-contact

## 狀態

<!-- 就地改值，不追加。中斷後重新派工時，開發 Agent 以這裡判斷從哪裡接續；不寫 commit SHA（以 git log --grep "wip(<ISSUE-ID>): <任務編號>" 對應）。
  階段由誰寫入：proposal（architect 建立；主對話提交提案時更新「最後完成」）→ backend（後端 Agent 開工）
  → frontend（後端 Agent 交棒，或沒有 [BE] 任務時前端 Agent 開工）→ integrating（最後一個開發 Agent 交接）→ reviewing（integrator 檢查通過；主對話做 PR 前審查）
  → pr-checking（integrator 收尾重切並開 PR 後；主對話做 PR 中檢查）
  → awaiting-approval（只有高風險：主對話等待 user 回覆「合併」）→ archived（integrator 歸檔；之後由 integrator 合併）。
  integrator 檢查、PR 前審查、CI 或 PR 中檢查退回時改回被退回那一端的 backend／frontend（退回 architect 為 proposal）。
  合併隊列失敗不改階段，停下交給 user。 -->

- **階段**：pr-checking <!-- proposal | backend | frontend | integrating | reviewing | pr-checking | awaiting-approval | archived -->
- **最後完成**：PR 中檢查
- **下一步**：風險審批（中，自動往下）
- **暫停原因**：無
- **更新**：主對話 — 2026-11-25 23:39

---

## architect（Hikaru）— 2026-11-25 23:29

- **做了什麼**：讀 SDP-187，確認前台頁尾（`frontend/components/front/front-shell.jsx`）目前只有版權字樣與後台可編輯的社群連結，沒有固定聯絡資訊。建立 proposal／design／specs（新 capability `site-contact-info`）／tasks（4 項 [FE]、0 項 [BE]）／openapi.yaml（`paths: {}`）。決策：Email 放 `SITE.contactEmail` 常數、頁尾 mailto 連結且以小寫顯示、不加功能開關、manifest 記兩筆偏離（design.md D1–D4）。
- **未完成**：無。
- **風險**：頁尾由 client component 在 settings 載入後才渲染，執行驗證只能以打包產物 grep 觀察；兩主題與小寫顯示需靠審查 diff 與前端 Agent 手動確認（專案未設定截圖工具）。
- **API 契約變動**：無
- **退回**：無

---

## frontend（Fujimori）— 2026-11-25

- **做了什麼**：`SITE.contactEmail`（data/site.js）；front-shell.jsx 頁尾新增「Contact: <mailto 連結>」，Email 套 `normal-case tracking-normal`、`hover-rule`，無 FeatureGate、不開新分頁；manifest 新增兩筆 SDP-187 偏離。
- **檢查結果**：`npm run test:design` 38 pass／0 fail；`npm run lint` 0 error（9 個既有 `<img>` warning）；`npm run typecheck` 通過；`npm run build` 通過；`npm audit --audit-level=high` 通過（僅 7 個 moderate，既有）；`grep -rlF 'steven0210423@gmail.com' .next/static` = 1（>0）；含 `mailto:` 且含 Email 的 chunk = 1；`normal-case` 已進 CSS。
- **手動確認**：環境無瀏覽器，也未啟動後端，未能實際開頁面看 MAG／TERM。以程式碼確認：頁尾為兩主題共用單一 JSX、顏色沿用 `v-dim`、Email 有 `normal-case`。建議 integrator／審查補視覺確認。
- **未完成**：無。仍靠 mock：無。
- **API 契約變動**：無

---

## integrator（Teshima）— 檢查通過

- **前端（frontend/）**：`npm ci` 成功；`npm run test:design` 38 pass／0 fail；`npm run lint` 0 error（9 個既有 `<img>` warning）；`npm run typecheck` 通過；`npm run build` 通過；`npm audit --audit-level=high` exit 0。
- **規格**：`openspec validate sdp-187-footer-contact --strict` 有效；`openspec validate --specs` 25 passed／0 failed。
- **後端**：無 [BE] 任務，不跑後端檢查。無 `openspec/spec-sync.yaml`，規格同步不適用；無規格免同步聲明。
- **核對**：tasks 1.1–1.4 全勾；worktree 乾淨、與 origin 一致；diff 只動 frontend 與變更目錄；openapi `paths: {}` 無契約變動。
- **變異測試**：不適用（design.md 目標為「無」，CLAUDE.md 亦未設定）。
- **執行驗證（專案通用）**：後端 docker 啟動、前端 next start（3100）就緒；`/health` = `{"status":"ok","database":"up"}`；`/api/v1/articles` = 200；`/blog/a1` = 200。
- **執行驗證（本次專屬）**：打包產物含 Email 的檔案數 = 1（>0）；同時含 `mailto:` 與 Email 的 chunk = 1（>0）；`/` = 200；CSS 含 `normal-case`。
- **截圖觀察點**：無設定；環境無瀏覽器，未能實際看 MAG／TERM 頁尾。
- **停止**：docker compose down -v、next 已停，3100／18080 無殘留。
- **風險**：兩主題畫面與小寫顯示僅能由程式碼審視佐證（頁尾為兩主題共用 JSX、Email 有 `normal-case tracking-normal`、色彩沿用 `v-dim`），請 PR 前審查確認。

---

## integrator（Teshima）— 收尾

- **重切前 SHA（OLD）**：`cc2d914f1f23092f92013d1aa0ec942c322a2380`
- **重切起點（FROM）**：merge-base origin/main（`git merge-base HEAD origin/main`）
- **重切後 commit**：`09c763f` feat(web): 前台頁尾新增聯絡 Email（SDP-187）（前端程式、manifest 與變更目錄同一批；本 change 無後端）
- **tree 比對**：`git diff OLD HEAD` 為空
- **PR**：https://github.com/StevenShih-0402/littleduck_website/pull/23
- **附進 PR 的 review.md 備註**：1 條（FE，規格情境無自動化測試）
- **風險**：無新增
- **收尾 HEAD**：`09c763fa49749a1b65ae8ecf30e88284f11e8c2f`
