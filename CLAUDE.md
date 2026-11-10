# team-hq — 多 Agent 協作中樞

本 repo 同時是 agent-team 的 team-hq 與**唯一的開發 repo**：前端在 `frontend/`，後端在 `backend/`，Agent 定義、總指揮指令與角色 skill 在 `.claude/`。平常從 repo 根目錄啟動 Claude Code。
完整的設計與導入步驟見 `docs/multi-agent-workflow.md`（skill_repo 的通用版）；本專案與通用版不同的地方，以下方「單一 repo 設定」為準。

## 單一 repo 設定（本專案）

agent-team 的通用文件預設「前端 repo、後端 repo 各一個、各自一份 worktree」。本專案是 monorepo，凡是通用文件、skill 或 Agent 定義提到「兩端」「每一端」「前端 repo／後端 repo」，一律依下表解讀：

| 通用說法 | 本專案的意思 |
|----------|--------------|
| 前端 repo、後端 repo、team-hq | 都是 `littleduck_website`（本 repo），主幹 `main` |
| 前後端兩份 worktree | **一份** worktree：`wt/<change-id>`（repo 根目錄下，已列入 `.gitignore`） |
| 前後端兩份變更目錄 | **一份**：`openspec/changes/<change-id>/`，`tasks.md` 同時有 `[BE]` 與 `[FE]` 任務 |
| 後端定稿 `openapi.yaml` 後複製到前端 | 只有一份 `openapi.yaml`，Yasumi 就地定稿，不複製 |
| 兩端各一個 PR、合併依「後端→前端」 | **一個 PR**，合併一次 |
| 前端／後端的 `CLAUDE.md` | 都是本檔 |

- worktree 從 `origin/main` 開：在 repo 根目錄執行 `git worktree add -b <branch> wt/<change-id> origin/main`。
- **Yasumi 與 Fujimori 共用同一份 worktree，必須依序派工**（`repos` 含 backend 時先 Yasumi、再 Fujimori），不可平行，因為兩者共用同一個 git index。
- 寫入範圍以子目錄劃分（`.claude/hooks/guard-write-scope.sh`）：Yasumi 只能寫 `backend/` 與進行中的變更目錄，Fujimori 只能寫 `frontend/` 與進行中的變更目錄。
- Yasumi 的流程 skill 是根目錄的 `.claude/skills/backend-processor/`（`install.sh backend-processor .` 安裝在本 repo）。**不要**在 `backend/`、`frontend/` 底下另裝 `.claude/`：Claude Code 會把子目錄的 skill 也載入，造成新舊版本並存。

## Worktree 慣例

- 每個 issue 一份 worktree：`wt/<change-id>`。
- change-id **全小寫**，以 Linear issue 編號開頭，例：`sdp-123-order-export`。OpenSpec CLI 不接受大寫。
- 分支：`<label>/<ISSUE-ID>-<code>`，ISSUE-ID 維持大寫，例：`feature/SDP-123-order-export`。Linear automation 靠它辨識 issue。
- 開發 Agent 只在該 worktree 內工作。

## Agent 與指令

| 名稱 | 角色 | 定義檔 | 流程步驟在 |
|------|------|--------|-----------|
| hikaru | architect：需求分析、UML、OpenAPI 草稿、提案 | `.claude/agents/hikaru.md` | `.claude/skills/issue-architect/` |
| yasumi | backend：Go API、單元測試、資料表變更 | `.claude/agents/yasumi.md` | `.claude/skills/backend-processor/` |
| fujimori | frontend：Next.js 畫面、API 串接 | `.claude/agents/fujimori.md` | `.claude/skills/frontend-processor/` |
| teshima | integrator：檢查（含執行驗證）、審查通過後收尾開 PR、歸檔、回寫 Linear、模擬合併隊列並合併 | `.claude/agents/teshima.md` | `.claude/skills/issue-integrator/` |
| akafuku | reviewer：PR 前對照 OpenSpec 規格審查 diff 的需求／測試覆蓋與邏輯；PR 中評定 /code-review、/security-review 的發現 | `.claude/agents/akafuku.md` | `.claude/skills/custom-review/` |
| `/ship-issue <id>` | 串起五個角色的總指揮：從提案一路自動做到合併，只在停下條件成立時問 user | `.claude/commands/ship-issue.md` | — |

- 定義檔只寫身分、工具、模型、邊界與交接規則；**調整流程步驟只改 skill**。
- 角色 skill 由 skill_repo（littleduck-skills）的 `install.sh agent-team .` 與 `install.sh backend-processor .` 安裝，重跑即更新。骨架檔（Agent 定義、`ship-issue.md`、hook、範本）不會被覆寫，本專案的單一 repo 調整寫在這些檔案裡。
- **Agent 名稱是本專案自訂的**（SDP-88：hikaru／yasumi／fujimori／teshima／akafuku），littleduck-skills 的 skill 來源仍用範本舊名。重跑 `install.sh` 會把 `.claude/skills/` 覆寫回舊名，重跑後要重新套用 SDP-88 的改名（新舊名稱對照見 `openspec/changes/archive/` 中的 SDP-88 提案），並以不分大小寫的 grep 確認 `.claude/skills/` 沒有舊名殘留。
- hikaru、teshima、yasumi、fujimori 都掛有寫入限制 hook。Akafuku 的 `tools` 只有 Read／Grep／Glob，天生不能寫檔。
- **內建 `/code-review` 不要建立同名 skill**：同名的專案或個人 skill 會整個取代內建審查。Akafuku 用的是另取名的 `/custom-review`。
- subagent 不能再派 subagent，也不能用 AskUserQuestion，所以要問 user 的決定都在主對話（`/ship-issue`）先問完再交棒。

## 設計同步

網站設計放在 claude.ai，改完設計後用 `/design-pull` 同步到 `frontend/`。同步是單向、手動觸發的，定義檔在 `.claude/skills/design-pull/SKILL.md`。

- 同步基準：`frontend/design-sync/manifest.json`，記錄設計來源、各設計檔雜湊與對應的前端檔案
- 差異工具：在 `frontend/` 執行 `npm run design:check -- <拉取目錄>`，測試用 `npm run test:design`

## 交接規則

- 所有交接都落在 `openspec/changes/<change-id>/`，Agent 之間不共享對話。
- 每個 Agent 開工先讀，結束時在 `handoff.md` 追加紀錄（範本：`openspec/handoff-template.md`）；程式碼審查結果寫在 `review.md`（範本：`openspec/review-template.md`）。
- `handoff.md` 開頭的「## 狀態」區塊就地改值，是流程進度的唯一依據；中斷後重新派工時，開發 Agent 依它接續。階段：`proposal → backend → frontend → integrating → reviewing → pr-checking → awaiting-approval（只有高風險）→ archived`。
- **流程中不固定停下等 user 確認**。主對話只在停下條件成立時問 user（完整清單見 `/ship-issue`「停下條件」）。
- 同一目標退回上限 2 次，超過交給使用者；**Teshima 檢查、PR 前審查、CI、PR 中評定的退回合併計算**。合併隊列失敗不退回，直接交給使用者。

## 層層把關

| 階段 | 檢查 | 誰做 |
|------|------|------|
| PR 前 | 確保正常運行：測試、lint、typecheck、執行驗證 | Teshima「檢查」 |
| PR 前 | 檢查代碼品質：對照規格審需求與測試覆蓋 | Akafuku 審查模式 |
| PR 中 | CI：等 Drone 的 `continuous-integration/drone/pr` check 通過（見 `docs/cicd.md`） | 主對話 |
| PR 中 | AI 審核：內建 `/code-review` | 主對話，Akafuku 評定 |
| PR 中 | 安全審核：內建 `/security-review`（只在高風險時） | 主對話，Akafuku 評定 |
| PR 中 | 按風險審批：低／中自動往下，高停下等 user 回覆「合併」 | 主對話 |
| 合併時 | 合併隊列：rebase 到最新 `main`、重跑檢查與 CI，全綠才合併 | Teshima「合併」 |

合併後的部署（Staging → promote → Production）由 Drone 的 CD 負責，不在本流程內，見 `docs/deployment-runbook.md`。

## 審查規則三分法

| 規則性質 | 放哪裡 | 誰檢查 | 何時 |
|----------|--------|--------|------|
| 機械性（格式、import 順序） | `gofmt`／`go vet`（CI 與 Teshima 檢查都會跑） | Teshima「檢查」、Drone CI | PR 前、PR 中 |
| 對照 OpenSpec 的需求覆蓋與測試涵蓋 | `/custom-review` | Akafuku（審查模式） | PR 前 |
| 需要判斷的專案慣例（分層、錯誤格式、命名） | 本檔與 `backend/README.md` | 內建 `/code-review`，Akafuku（評定模式）評級 | PR 中 |

- 機械性規則不交給 AI 審；Akafuku 遇到 `/code-review` 回報機械性項目會判為不成立。
- 只有 Akafuku 評為「高嚴重度且高信心」的發現退回，其餘寫成 PR 備註並留言到 PR。
- 第 2 輪起只審上一輪之後的新改動，前幾輪已記錄的發現不再重提。

### 高風險檔案樣式

`/ship-issue` 第 7 步以 `git diff --name-only` 比對本清單，命中就把 `/code-review` 的 effort 調為 `high`、跑 `/security-review`，並把變更判定為高風險。

| 類別 | 樣式 |
|------|------|
| CI／CD 與部署 | `.drone.yml`、`deploy/**`、`**/Dockerfile`、`backend/docker-compose*.yml`、`frontend/vercel.json` |
| 流程設定 | `.claude/hooks/**`、`.claude/agents/**` |
| 資料表結構 | `backend/internal/model/**`、`backend/internal/database/**`（GORM AutoMigrate） |
| 金流 | `backend/internal/service/payment/**`、`backend/internal/controller/payment_controller.go` |
| 認證與權限 | 路徑或檔名含 `auth`、`token`、`session`、`jwt`、`permission`、`guard`；`backend/internal/middleware/**` |

### 風險分級

依下表判定**整個變更**的風險等級：

| 等級 | 條件（任一成立） | 處置 |
|------|-----------------|------|
| 高 | 命中上方「高風險檔案樣式」；破壞性 API 變更；`/security-review` 有成立的發現（列為 PR 備註） | 停下，等 user 看過高風險檔案的 diff 後回覆「合併」 |
| 中 | 動到 `backend/`、`frontend/` 的執行期程式碼或相依版本（非破壞性） | 自動往下，最終報告列出 |
| 低 | 只動文件、測試、註解 | 自動往下 |

### 合併設定

| 項目 | 值 |
|------|----|
| 合併方式 | `--merge`（保留重切後的功能 commit） |
| CI 等待上限 | 30 分鐘（超過就停下回報，不視為通過） |
| `/security-review` | 審查「目前分支 vs `origin/main`」，本 repo 的主幹就是 `origin` 預設分支，可直接使用 |

### 執行驗證

Teshima「檢查」依本段把產品跑起來觀察。指令在 worktree 根目錄執行。

- **啟動**：
  - 後端：`API_PORT=18080 docker compose -f backend/docker-compose.yml -p "ld-check-$(basename "$PWD")" up -d --build`
  - 前端：`(cd frontend && npm ci && npm run build && { nohup ./node_modules/.bin/next start -p 3100 > /tmp/ld-next.log 2>&1 & echo $! > /tmp/ld-next.pid; })`
- **就緒檢查**（等待上限 60 秒）：`curl -fsS localhost:18080/health` 與 `curl -fsS -o /dev/null localhost:3100/`
- **觀察點**：
  - `curl -s localhost:18080/health` → `{"status":"ok","database":"up"}`
  - `curl -s -o /dev/null -w '%{http_code}' localhost:18080/api/v1/articles` → `200`
  - `curl -s -o /dev/null -w '%{http_code}' localhost:3100/blog/a1` → `200`
- **停止**：`API_PORT=18080 docker compose -f backend/docker-compose.yml -p "ld-check-$(basename "$PWD")" down -v` 與 `kill "$(cat /tmp/ld-next.pid)"`（直接執行 `node_modules/.bin/next` 而非 `npx`，pid 檔記到的才是服務本身；也不要用 `pkill -f`，會連下指令的 shell 一起比對到）

後端用 18080 埠，避開本機常被佔用的 8080；同一時間只能檢查一個 worktree。本次變更特有的觀察點由 Hikaru 寫在 `design.md` 的「## 執行驗證觀察點」，Teshima 兩者都跑。

### review-diff.patch 不進版控

`/ship-issue` 第 5、7 步會在變更目錄產生 `review-diff.patch` 給唯讀的 Akafuku 讀，審完即刪、不進 commit。根目錄 `.gitignore` 已列入。

## 新舊流程如何選用

| 情境 | 用哪條 |
|------|--------|
| 前後端都要改，或要讓 Agent 分工開發 | `/ship-issue <id>` |
| 小 issue、只動一端 | 單點呼叫 Yasumi（後端）或 Fujimori（前端），或 `/ship-issue` 只填一端 |
| 自己帶著做 | `/issue-reader` → `/issue-processor` → `/issue-solver` |

兩條流程**不可同時操作同一張 issue**。
