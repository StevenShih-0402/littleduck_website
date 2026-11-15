<!-- handoff.md 範本：複製到 openspec/changes/<change-id>/handoff.md。「狀態」區塊就地改值、永遠反映最新進度；下方各角色紀錄只追加，不改前面的內容。 -->
# Handoff：sdp-168-install-sites-cert-check

## 狀態

<!-- 就地改值，不追加。中斷後重新派工時，開發 Agent 以這裡判斷從哪裡接續；不寫 commit SHA（以 git log --grep "wip(<ISSUE-ID>): <任務編號>" 對應）。
  階段由誰寫入：proposal（architect 建立；主對話提交提案時更新「最後完成」）→ backend（後端 Agent 開工）
  → frontend（後端 Agent 交棒，或沒有 [BE] 任務時前端 Agent 開工）→ integrating（最後一個開發 Agent 交接）→ reviewing（integrator 檢查通過；主對話做 PR 前審查）
  → pr-checking（integrator 收尾重切並開 PR 後；主對話做 PR 中檢查）
  → awaiting-approval（只有高風險：主對話等待 user 回覆「合併」）→ archived（integrator 歸檔；之後由 integrator 合併）。
  integrator 檢查、PR 前審查、CI 或 PR 中檢查退回時改回被退回那一端的 backend／frontend（退回 architect 為 proposal）。
  合併隊列失敗不改階段，停下交給 user。 -->

- **階段**：reviewing <!-- proposal | backend | frontend | integrating | reviewing | pr-checking | awaiting-approval | archived -->
- **最後完成**：PR 前審查
- **下一步**：Teshima 收尾
- **暫停原因**：無
- **更新**：主對話 — 2026-11-15

---

## architect（Hikaru）— 2026-11-15 18:00

- **做了什麼**：建立 proposal／specs（新 capability `edge-sites`）／design（資料流圖、CD 互動時序圖、執行驗證觀察點 6 項、變異測試目標）／tasks（5 項 `[BE]`、0 項 `[FE]`）／openapi.yaml（`paths: {}`）。
- **未完成**：無（提案階段）。
- **風險**：
  - 命中高風險檔案樣式 `deploy/**`，整體判定**高風險**，PR 要等 user 回覆「合併」。
  - `[BE]` 任務改的是 `deploy/`、`docs/`，依賴主 repo `CLAUDE.md` 暫時放寬的「後端子目錄：backend/、deploy/、docs/」；**合併後要改回 `backend/`**。worktree 內的 `CLAUDE.md` 仍是 `backend/`（寫入限制 hook 讀的是主 repo 的設定）。
  - 本次沒有 `backend/` Go 程式碼變更；後端檢查指令（`go test`／`gofmt`／`go vet`）照跑即可，不會受影響。執行驗證的觀察點需要本機 Docker（假 Nginx 容器與 shellcheck 映像）。
  - 非破壞性；無 schema 變更。
- **API 契約變動**：無。
- **退回**：無。
- **合併後 user 實機驗收**（不屬 PR 內任務）：
  1. 以 root 執行 `chmod 700 /etc/letsencrypt/live /etc/letsencrypt/archive`。
  2. 以 `deploy` 執行 `install-sites.sh`（手動或觸發一次 CD）：憑證存在時能安裝並 reload。
  3. 以引用不存在憑證的 conf 執行：顯示原錯誤訊息 `…：憑證 … 不存在，請先依 runbook 核發` 並中止，`/opt/edge-sites` 不變。

## backend（Yasumi）— 2026-11-15

- **做了什麼**：
  - `install-sites.sh` 憑證檢查改為 `docker exec "$nginx_container" test -f "$cert"`，回傳 1 沿用原錯誤訊息，其他非 0 回報「無法在 Nginx 容器內檢查」；開頭註解補掛載前置。
  - 新增 `deploy/edge/test-install-sites.sh`（三案例，假 Nginx 容器、EDGE_SITES_DIR 暫存目錄、trap 清理）。
  - `docs/cicd.md`、`docs/deployment-runbook.md` 註明容器內檢查、權限維持 700、記 SDP-117 的 chmod 755 暫時繞過。
- **驗證**：`sh -n` 無輸出；`grep -c '\[ -f "\$cert" \]'` 為 0；測試腳本三行 PASS 與 `ALL PASS`（rc 0）；暫改回主機 `[ -f ]` 時「憑證存在」「無法檢查」案例 FAIL（暫改已還原未提交）；shellcheck v0.10.0 `-S warning` 無輸出；沒有 `go` 變更，未跑 go test。
- **偏差（請 integrator 留意）**：測試腳本的「既有容器」偵測用 `docker ps -q`（只看執行中），不含 `-a`。本機有 5 個已停止的 mygo-v2 容器（postgres、redis 等，非 nginx），用 `-a` 會讓測試永遠中止。因此 design 觀察點 6（`docker ps -a ... | wc -l` 為 0）在這台機器不會是 0；改以「測試前後數量相同（5）」確認沒有殘留，測試後仍為 5。
- **未完成**：無。無 `[FE]` 任務。
- **風險**：命中 `deploy/**`，高風險；非破壞性、無 schema 變更；主 repo CLAUDE.md 的後端子目錄放寬須於合併後改回。
- **API 契約變動**：無（openapi.yaml `paths: {}` 不需改）。

## integrator（Teshima）— 檢查通過

- **後端（backend/）**：`go test ./...` 全 ok；`gofmt -l .` 無輸出；`go vet ./...` 通過。本次無 Go 變更。
- **前端（frontend/）**：`npm ci` 成功；`npm run test:design` 32 pass／0 fail；`npm run lint` 0 error（9 既有 warning）；`npm run typecheck` 通過；`npm run build` 成功；`npm audit --audit-level=high` exit 0（僅 8 moderate）。本次無前端變更。
- **規格**：`openspec validate sdp-168-install-sites-cert-check --strict` valid；`openspec validate --specs` 23 passed。
- **migration**：無 model 變更，不適用。
- **變異測試**：略過（CLAUDE.md 未設定變異測試）。目標為 install-sites.sh 憑證檢查分流；Yasumi 已手動變異（改回主機 `[ -f ]` 時案例 FAIL）。
- **執行驗證（專案通用）**：本次不影響 backend／frontend 服務（只改 deploy/、docs/），未啟動；略過。
- **執行驗證（本次專屬 design.md 6 點）**：
  1. `sh -n` 兩支 → ok，通過
  2. shellcheck v0.10.0 `-S warning` → 無輸出 rc=0，通過
  3. `grep -c '\[ -f "\$cert" \]'` → 0，通過
  4. `sh deploy/edge/test-install-sites.sh` → 三行 PASS＋ALL PASS，rc=0，通過
  5. 輸出含「請先依 runbook 核發」次數 → 1，通過
  6. 殘留容器：本機有 5 個既有已停止 mygo-v2 容器（非 nginx），計數非 0；測試前 5、後 5，無殘留。判斷：偏差合理（測試腳本用 `docker ps -q` 只看執行中，若用 -a 會永遠中止），視為通過；風險見下。
- **tasks.md**：全勾選。worktree 乾淨、與 origin 一致；無越界 commit（無 frontend／backend 變動）；proposal.md 只有 `wip(SDP-168): 提案` 一筆。無 spec-sync.yaml，不適用。
- **規格免同步聲明**：無。
- **風險**：命中 deploy/**，整體高風險；design.md 觀察點 6 在有殘留已停止容器的機器上字面不為 0（可建議 Hikaru／後續把觀察點改為前後數量相同）；後端子目錄放寬須於合併後改回 backend/；不涵蓋實機 700 權限驗收（user 合併後手動）。
- **退回**：無。
