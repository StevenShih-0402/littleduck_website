# Tasks

> 本 change 只動流程設定與文件，不動 `backend/`、`frontend/` 執行期程式碼，因此沒有 `[BE]`／`[FE]` 任務；適合用 `/issue-processor` 由主對話實作，不經 `/ship-issue` 派工。
> 驗證用的搜尋指令（下稱「殘留檢查」）：`grep -rniE 'hyuse|chika|yuma|osamu|shiori' --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target --exclude-dir=wt --exclude-dir=archive .`

## 1. littleduck_website：Agent 定義與寫入限制 hook

- [x] 1.1 以 `git mv` 把 `.claude/agents/{hyuse,chika,yuma,osamu,shiori}.md` 改名為 `{hikaru,yasumi,fujimori,teshima,akafuku}.md`，`git status` 顯示五筆 rename
- [x] 1.2 改五個定義檔的 frontmatter：`name` 改新名、`color` 依名冊（hikaru purple、yasumi pink、fujimori green、teshima cyan、akafuku yellow），`description` 與內文的舊名改新名，hook 指令參數改新名；以 `grep -n '^name\|^color\|guard-write-scope' .claude/agents/*.md` 確認
- [x] 1.3 改 `.claude/hooks/guard-write-scope.sh`：註解、`case` 角色清單、未知角色訊息、`yasumi|fujimori` 分支與 `other=Fujimori/Yasumi`；以 hook JSON 手動測試四個情境並確認結果：`yasumi` 寫 `frontend/x` 被拒且訊息提到 Fujimori、`hikaru` 寫 `openspec/changes/x/proposal.md` 放行、`teshima` 寫 `backend/x` 被拒、`chika` 以未知角色被拒

## 2. littleduck_website：指令、skills 與文件

- [x] 2.1 改 `.claude/commands/ship-issue.md` 的 `subagent_type` 與說明文字為新名；檔內殘留檢查無結果
- [x] 2.2 改已安裝 skills：`issue-architect`、`backend-processor`（含 `hooks/post-edit-lint.sh` 註解）、`frontend-processor`（含 hooks 註解）、`issue-integrator`、`custom-review`；`.claude/skills/` 殘留檢查無結果
- [x] 2.3 改 `CLAUDE.md`（Agent 表格、單一 repo 設定、交接與審查段落），並在「Agent 與指令」段補一行：重跑 littleduck-skills 的 `install.sh` 會把 skills 覆寫回舊名，需重新套用 SDP-88 的改名；`CLAUDE.md` 殘留檢查無結果
- [x] 2.4 改 `docs/multi-agent-workflow.md`、`openspec/review-template.md`、`spec-sync/README.md`、`.gitignore` 註解；全 repo 殘留檢查只剩 `openspec/changes/sdp-88-rename-agents/`（本提案的對照表）

## 3. code_city：分支與改名（跨 repo，`/home/steven/side_projects/code_city`）

- [x] 3.1 在 code_city 執行 `git fetch origin && git worktree add -b improvement/SDP-88-rename-agents wt/sdp-88-rename-agents origin/main`（若 code_city 未忽略 `wt/`，改在 code_city 根目錄 `git switch -c` 開分支），之後的 code_city 任務都在該處進行；`git branch --show-current` 顯示新分支（已於 `code_city/wt/sdp-88-rename-agents` 建立）
- [x] 3.2 以 `git mv` 把 `.claude/agents/{hyuse,yuma,osamu,shiori}.md` 改名為 `{gintoki,kagura,shinpachi,sadaharu}.md`，改 frontmatter：`name` 改新名、`color`（gintoki cyan、kagura red、shinpachi blue、sadaharu pink）、`description`、內文與 hook 參數；`grep -n '^name\|^color' .claude/agents/*.md` 確認
- [x] 3.3 改 code_city 的 `.claude/hooks/guard-write-scope.sh`（`gintoki|shinpachi`），以 hook JSON 測試 `gintoki` 寫 `openspec/x` 放行、`shinpachi` 寫 `src/x` 被拒、`hyuse` 以未知角色被拒
- [x] 3.4 改 code_city 的 `.claude/commands/ship-issue.md`、`.claude/skills/**`（含 `gen-story` 的 `SKILL.md` 角色清單與 `assets/story-template.html` 的樣式 class：`.hyuse`→`.gintoki`、`.yuma`→`.kagura`、`.osamu`→`.shinpachi`、`.shiori`→`.sadaharu`）、`CLAUDE.md`（同 2.3 補 install.sh 提醒）、`docs/multi-agent-workflow.md`、`openspec/review-template.md`、`spec-sync/README.md`、`.gitignore`、`openspec/specs/agent-ci-gate/spec.md`（`Yuma`→`Kagura`）；`chika`／`Chika` 保持不變
  - **已解除封鎖**：code_city PR #3（SDP-113）已於 2026-11-10 合併進 `origin/main`。code_city 的 `improvement/SDP-88-rename-agents` 分支已 rebase 到含 PR #3 的新 `origin/main`，`agent-ci-gate/spec.md` 的 `Yuma`（2 處）已改為 `Kagura`，並隨 commit `3787b55` 推上遠端。code_city 殘留檢查（不含 chika）在 `agent-ci-gate` 已無命中。
- [x] 3.5 code_city 殘留檢查改用 `grep -rniE 'hyuse|yuma|osamu|shiori' ...`（不含 chika）無結果；`chika` 的剩餘出現處只有 design.md 決策 3 列出的通用說明位置
  - 結果：殘留檢查無結果。`chika` 剩在 `CLAUDE.md`、`docs/multi-agent-workflow.md`、`backend-processor`（SKILL.md、hooks）、`gen-story`（SKILL.md、`.chika` 樣式），以及決策 3 未列、但同屬「本專案沒有 backend 角色」說明的 `.claude/commands/ship-issue.md`（「也沒有 Chika」）與 `issue-architect/SKILL.md`（流程表的 backend 列）。

## 4. 整合驗證

- [x] 4.1 在兩個專案各重開一次 Claude Code session，`/agents` 列出新名稱與指定顏色，舊名不在清單內
  - **已驗證（littleduck）**：重開 session 後 Agent 清單已載入新名 `hikaru`／`yasumi`／`fujimori`／`teshima`／`akafuku`，舊名不在清單內；五個定義檔 frontmatter 的 `color` 為 purple／pink／green／cyan／yellow，與名冊一致。
- [x] 4.2 在 littleduck_website 以 `subagent_type: "akafuku"` 派一個唯讀小任務（例：讀 `openspec/review-template.md` 回報標題），確認能啟動
  - **已驗證**：重開 session 後以 `subagent_type: "akafuku"` 派唯讀任務（讀 `openspec/review-template.md` 回報首行）成功啟動並回報，未修改任何檔案。

## Workflow follow-up

- 本變更命中高風險樣式（`.claude/agents/**`、`.claude/hooks/**`），兩個 PR 都要等 user 看過 diff 後才合併。
- code_city 的 PR 由該 repo 自己合併；littleduck_website 的 PR 內文以連結交叉引用。
- `/issue-solver sdp-88-rename-agents` 歸檔，`agent-roster` 規格同步到 `openspec/specs/`。
