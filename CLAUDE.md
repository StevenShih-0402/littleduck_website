# littleduck_website

前端在 `frontend/`（Next.js），後端在 `backend/`（Go），多 Agent 流程的 Agent 定義、總指揮指令與角色 skill 在 `.claude/`。平常從 repo 根目錄啟動 Claude Code。
多 Agent 流程由 littleduck-skills 的 `install.sh agent-team-fullstack` 安裝，完整的設計與導入步驟見 `docs/multi-agent-workflow.md`。

## agent-team 設定

<!-- 角色 skill、/ship-issue 與寫入範圍 hook 都讀這一段；欄位名不要改。 -->

- **結構**：fullstack
- **主幹**：main
- **worktree 根目錄**：wt/
- **後端子目錄**：backend/
- **前端子目錄**：frontend/

- `wt/` 與 `review-diff.patch` 已列入 `.gitignore`。
- **不要在 `backend/`、`frontend/` 底下建立 `.claude/`**：Claude Code 會把子目錄的 skill 也載入，造成新舊版本並存。

## 檢查指令

開發 Agent 與 Teshima 依這張表在各子目錄跑檢查；與 Drone CI 的 `ci-backend`、`ci-frontend` 相同。

| 項目 | 後端（在 `backend/` 執行） | 前端（在 `frontend/` 執行） |
|------|------|------|
| 安裝依賴 | 無（`go` 指令自動下載模組） | `npm ci` |
| 測試 | `go test ./...` | `npm run test:design` |
| lint | `gofmt -l .`（要沒有輸出）、`go vet ./...` | `npm run lint` |
| 型別檢查 | 無（`go vet` 與編譯已涵蓋） | `npm run typecheck`（`checkJs`，不轉 TypeScript） |
| 建置 | — | `npm run build` |
| 相依套件弱點 | — | `npm audit --audit-level=high` |
| migration 正反向 | 無獨立 migration 檔，見下 | — |
| API schema 產出 | 無 | — |

- **資料表變更走 GORM AutoMigrate**（`backend/internal/database/`），沒有獨立的 migration 檔。「migration 正反向」改為：新增的 model 欄位能在空資料庫與既有資料庫上啟動成功（執行驗證時確認），並在 `handoff.md` 註明是否為不相容的 schema 變更。
- 編輯後的 lint hook：前端 hook 會在 `frontend/` 已 `npm ci` 時，對剛編輯的 JS 檔跑 ESLint（設定在 `frontend/eslint.config.mjs`）；後端 hook 只處理 `.py`，在本專案不會作用，`gofmt`／`go vet` 要自己跑。
- 前端既有的 lint／型別違規只能以帶理由的單行註解豁免（`eslint-disable-next-line <規則> -- 理由`、`@ts-expect-error 理由`），不關閉、不降級規則。
- 改到設計稿對應的前端檔案時，`frontend/design-sync/manifest.json` 的對應關係要維持正確（見「設計同步」）。

## Worktree 慣例

- 每個 issue **一份** worktree：在 repo 根目錄執行 `git worktree add -b <branch> wt/<change-id> origin/main`。
- change-id **全小寫**，以 Linear issue 編號開頭，例：`sdp-123-order-export`。OpenSpec CLI 不接受大寫。
- 分支：`<label>/<ISSUE-ID>-<code>`，ISSUE-ID 維持大寫，例：`feature/SDP-123-order-export`。Linear automation 靠它辨識 issue。
- 一份變更目錄（`wt/<change-id>/openspec/changes/<change-id>/`）、一份 `openapi.yaml`、一份 `handoff.md`、一份 `review.md`、一個 PR。
- Yasumi 與 Fujimori **共用這份 worktree，依序派工**（先後端、後前端），不平行：兩者共用同一個 git index。

## Agent 與指令

| 名稱 | 角色 | 定義檔 | 流程步驟在 | 寫入限制 hook |
|------|------|--------|-----------|---------------|
| hikaru | architect：需求分析、UML、OpenAPI 草稿、提案 | `.claude/agents/hikaru.md` | `.claude/skills/issue-architect/` | `architect`：只寫 `openspec/` |
| yasumi | 後端 Agent：Go API、單元測試、資料表變更 | `.claude/agents/yasumi.md` | `.claude/skills/backend-processor/` | `backend`：只寫 `backend/` 與變更目錄 |
| fujimori | 前端 Agent：Next.js 畫面、API 串接 | `.claude/agents/fujimori.md` | `.claude/skills/frontend-processor/` | `frontend`：只寫 `frontend/` 與變更目錄 |
| teshima | integrator：檢查（含執行驗證）、審查通過後收尾開 PR、歸檔、回寫 Linear、模擬合併隊列並合併 | `.claude/agents/teshima.md` | `.claude/skills/issue-integrator/` | `integrator`：只寫 `handoff.md`、`.linear.md` 與歸檔同步的 `openspec/specs/` |
| akafuku | reviewer：PR 前對照 OpenSpec 規格審查 diff；PR 中評定 /code-review、/security-review 的發現 | `.claude/agents/akafuku.md` | `.claude/skills/custom-review/` | 不需要：`tools` 只有 Read／Grep／Glob |
| `/ship-issue <id>` | 串起五個角色的總指揮：從提案一路自動做到合併，只在停下條件成立時問 user | `.claude/commands/ship-issue.md` | — | — |

- 定義檔只寫身分、工具、模型、邊界與交接規則；**調整流程步驟只改 skill**。
- 角色 skill 由 littleduck-skills 的 `install.sh agent-team-fullstack .` 安裝，重跑即更新；骨架檔（本檔、Agent 定義、`ship-issue.md`、hook、範本）已存在時不會被覆寫。
- **Agent 名稱是本專案自訂的**（SDP-88）。角色 skill 只用角色稱呼、hook 的參數是角色，所以重跑 installer 後**不必**再改 `.claude/skills/`；名稱只寫在 Agent 定義、`ship-issue.md`、`docs/multi-agent-workflow.md` 與本檔。**改名後重跑 installer 會再建立範本名稱的 Agent 定義**（同名檔不存在就建立），重跑後要刪掉多出來的範本名稱定義檔，否則會有兩套 Agent。
- **內建 `/code-review` 不要建立同名 skill**：同名的專案或個人 skill 會整個取代內建審查。Akafuku 用的是另取名的 `/custom-review`。
- subagent 不能再派 subagent，也不能用 AskUserQuestion，所以要問 user 的決定都在主對話（`/ship-issue`）先問完再交棒。

## 設計同步

網站設計放在 claude.ai，改完設計後用 `/design-pull` 同步到 `frontend/`。同步是單向、手動觸發的，定義檔在 `.claude/skills/design-pull/SKILL.md`。

- 同步基準：`frontend/design-sync/manifest.json`，記錄設計來源、各設計檔雜湊與對應的前端檔案
- 差異工具：在 `frontend/` 執行 `npm run design:check -- <拉取目錄>`，測試用 `npm run test:design`
- 移植完成後要通過 `npm run lint` 與 `npm run typecheck`（設計稿的寫法可能觸發 lint 規則，移植時一併處理）

## 交接規則

- 所有交接都落在 `openspec/changes/<change-id>/`，Agent 之間不共享對話。
- 每個 Agent 開工先讀，結束時在 `handoff.md` 追加紀錄（範本：`openspec/handoff-template.md`）；程式碼審查結果寫在 `review.md`（範本：`openspec/review-template.md`）。
- `handoff.md` 開頭的「## 狀態」區塊就地改值，是流程進度的唯一依據；中斷後重新派工時，開發 Agent 依它接續。階段：`proposal → backend → frontend → integrating → reviewing → pr-checking → awaiting-approval（只有高風險）→ archived`。沒有 `[BE]` 任務時跳過 `backend`，沒有 `[FE]` 任務時 Yasumi 直接交 `integrating`。
- Hikaru 沒有待確認問題時，主對話直接提交並推送提案（不等 user 核准）；開發 Agent 每完成一項任務就 WIP commit 並 push；PR 前審查通過後，Teshima 收尾時把 WIP commit 重切成功能 commit。
- **流程中不固定停下等 user 確認**。主對話只在停下條件成立時問 user（完整清單見 `/ship-issue`「停下條件」）。
- 同一目標（Hikaru／Yasumi／Fujimori）退回上限 2 次，超過交給使用者；**Teshima 檢查、PR 前審查、CI、PR 中評定的退回合併計算**。退回對象依發現所屬的端決定。合併隊列失敗不退回，直接交給使用者。

## 層層把關

| 階段 | 檢查 | 誰做 |
|------|------|------|
| PR 前 | 確保正常運行：`backend/`、`frontend/` 各自的檢查指令，以及執行驗證 | Teshima「檢查」 |
| PR 前 | 驗證測試有效性：對關鍵模組跑變異測試 | Teshima「檢查」 |
| PR 前 | 檢查代碼品質：對照規格審需求與測試覆蓋 | Akafuku 審查模式 |
| PR 中 | CI：等 Drone 的 `continuous-integration/drone/pr` check 通過（見 `docs/cicd.md`） | 主對話 |
| PR 中 | AI 審核：內建 `/code-review` | 主對話，Akafuku 評定 |
| PR 中 | 安全審核：內建 `/security-review`（只在高風險時） | 主對話，Akafuku 評定 |
| PR 中 | 按風險審批：低／中自動往下，高停下等 user 回覆「合併」 | 主對話 |
| 合併時 | 合併隊列：rebase 到最新 `main`、重跑檢查與 CI，全綠才合併 | Teshima「合併」 |

合併後的部署（Staging → promote → Production）由 Drone 的 CD 負責，不在本流程內，見 `docs/deployment-runbook.md`。

## 功能開關（SDP-149）

- 前台新增一般瀏覽者可操作的功能（表單、按鈕、互動區塊）時，MUST 同時：
  1. 在後端功能開關登記表（`backend/internal/service/feature_service.go` 的 `featureRegistry`）新增一個 key，新功能預設**停用**（`Default: false`）；
  2. 在前端 `frontend/lib/feature-flags.mjs` 的 `FEATURES` 加上同一個 key，並以 `FeatureGate`（`frontend/components/feature-gate.jsx`）包住該操作區塊：啟用時顯示，停用時一般瀏覽者看不到；
  3. 若該功能有公開寫入端點，後端在開關停用時回 403 `FEATURE_DISABLED`，但帶有效管理員 JWT 的請求放行（管理員預覽，路由掛 `middleware.OptionalAuth`）。
- 只藏操作區塊，不藏導覽與既有公開內容。
- 管理員預覽：管理員在後台登入後，同一分頁的前台會顯示停用中的區塊並標示「停用中・預覽」，可實際試用；確認無誤後到後台「功能開關」分頁啟用才對外開放。預覽時送出的資料會寫進真實資料庫，試用後要自行清理。
- 開關讀取失敗一律視為停用（fail-closed），頁面其他內容照常顯示。
- 功能穩定、不再需要開關時，另開 issue 移除開關與相關判斷，不留永久開關。

## 審查規則三分法

| 規則性質 | 放哪裡 | 誰檢查 | 何時 |
|----------|--------|--------|------|
| 機械性（格式、import 順序、前端 lint 規則） | 後端 `gofmt`／`go vet`、前端 ESLint `npm run lint`（CI 與 Teshima 檢查都會跑） | Teshima「檢查」、Drone CI | PR 前、PR 中 |
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

## 執行驗證

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

後端用 18080 埠，避開本機常被佔用的 8080；同一時間只能檢查一個 worktree。本次變更特有的觀察點由 Hikaru 寫在 `design.md` 的「## 執行驗證觀察點」，Teshima 兩者都跑。截圖觀察點的格式與注意事項見 `docs/multi-agent-workflow.md`。

## 變異測試

Teshima「檢查」會對 Hikaru 在 `design.md`「## 變異測試目標」標示的關鍵模組跑變異測試；開發 Agent 與 Akafuku 都不跑。

本專案**尚未設定**變異測試指令，Teshima 會記為「略過」、不擋流程。要啟用時在這裡補上（前後端工具不同時分別寫）：

<!--
- **後端指令**：`<變異測試指令，接受目標模組/路徑參數>`
- **前端指令**：`<變異測試指令>`
- **通過閾值**：`<變異分數下限，例：80%>`
-->

## 新舊流程如何選用

| 情境 | 用哪條 |
|------|--------|
| 前後端都要改，或要讓 Agent 分工開發 | `/ship-issue <id>` |
| 小 issue、只動一端 | `/ship-issue <id>`（只有一端有任務時只派那一端），或單點呼叫 Yasumi（後端）、Fujimori（前端） |
| 自己帶著做 | `/issue-reader` → `/issue-processor` → `/issue-solver` |

兩條流程**不可同時操作同一張 issue**。
