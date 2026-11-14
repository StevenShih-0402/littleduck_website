# Proposal

## Why

SDP-113 把 Drone 的 CI／CD 設定、部署腳本與文件合併進 repo；SDP-157 再把前端從 Vercel 改成與後端一樣自架在 Hetzner（同一套「GHCR 映像 → Staging → promote 同一映像到 Production」）。前端的主機名已在 SDP-157 寫成實際值，但後端與 Drone 的位址仍是 `<網域>`、`api-staging.<網域>`、`<staging>`／`<production>` 佔位，前端 Staging／Production 的 `.env` 範例甚至直接引用這些佔位。user 照 runbook 架設時無從複製正確的值，也就無法開始 SDP-117 的實機上線與驗收。

網域已定案，本次把它落實到 repo，讓 user 能照文件直接上線：

| 用途 | 主機名 | 取代原佔位 |
|------|--------|-----------|
| 前端 Production | `stevenshih-littleduck.com` | （SDP-157 已是實際值） |
| 前端 Staging | `staging.stevenshih-littleduck.com` | （SDP-157 已是實際值） |
| Drone | `drone.stevenshih-littleduck.com` | `drone.<網域>` |
| 後端 Production | `api.stevenshih-littleduck.com` | `api.<網域>`、`<production>` |
| 後端 Staging | `api.staging.stevenshih-littleduck.com` | `api-staging.<網域>`、`<staging>` |

## What Changes

- `docs/deployment-runbook.md`：
  - 開頭「`<網域>` 尚未決定」的說明與前端主機名表，合併成上表五列的主機名對照表（含對應的 Drone secret 或設定位置）。
  - 全文 `drone.<網域>`、`api.<網域>`、`api-staging.<網域>` 換成實際主機名：第 0 節 DNS、第 1 節 OAuth App URL／certbot／`DRONE_HOST=` render 範例／登入驗證、第 2 節環境表與憑證指令（改成兩行明確的 `certbot` 指令，原本的 `-d <網域>` 有歧義）、第 3 節前端 `.env` 的 `API_BASE_URL`、第 5 節 Staging 後端檢查、第 6 節回滾的 `curl`。
  - 第 0 節 DNS 前置補上三筆 A 記錄（`drone.`、`api.`、`api.staging.`）與 Cloudflare DNS only（灰雲）的要求，與第 3 節前端 DNS 的寫法一致。
- `docs/cicd.md`：
  - 後端 CD 冒煙目標 `https://<staging>`、`https://<production>` 換成實際主機名。
  - Secret 清單加「值」欄，填入四個主機名 secret 的實際值；其他 secret 寫值的來源。
  - 文末新增「## 各階段耗時」段落與空白紀錄表，由 user 實機量測後填寫（對應 SDP-113 tasks 6.1）。
- `deploy/` 只改註解與範本值，不改邏輯：
  - `deploy/drone/.env.example`：`DRONE_SERVER_HOST` 預填 `drone.stevenshih-littleduck.com`，註解範例改實際主機名。
  - `deploy/ci/smoke-backend.sh`、`deploy/ci/smoke-frontend.sh`：用法註解的 `example.com` 範例改成實際主機名。
- 規格 delta：`ci-server`、`backend-delivery`、`frontend-delivery` 的場景把佔位改成實際主機名。

不需修改（系統分析結論，見 design.md Context）：

- `.drone.yml`：所有主機名都以 secret（`api_host`、`api_staging_host`、`frontend_host`、`frontend_staging_host`）帶入，換值只改 Drone secret。
- `deploy/edge/*.conf.template`、`deploy/edge/render.sh`、`deploy/ci/*-bundle.sh`、`remote-deploy.sh`、`install-sites.sh`、`deploy/web/*.yml`、`backend/docker-compose.*.yml`：樣板以變數帶入 `server_name` 與憑證路徑，多一層子網域（`api.staging.`）不影響渲染；`api-staging`、`web-staging` 是**環境名稱**（樣板檔名、容器名、目錄名），不是主機名，維持不變。
- 後端 CORS 目前為 `Access-Control-Allow-Origin: *`，兩個前端主機名都能呼叫，不需改後端程式碼。

無破壞性 API 變更，本次不動任何 HTTP 端點；也不動 `backend/`、`frontend/` 的任何檔案。

### 原 issue 中 Vercel 相關項目的處理

SDP-157 已停用 Vercel，issue 原文的 Vercel 步驟與驗收改為自架的等價項目（本次只改文件，實機項目由 user 手動做，清單見 handoff.md）：

| 原文 | 自架後的等價項目 |
|------|------------------|
| 建立 Vercel 專案（Root Directory、Node 版本） | 不適用。改為 runbook 第 3 節「前端主機初次建置」：記憶體量測、DNS、清理 mygo 舊 `staging.` 記錄、核發兩張前端憑證、建 `/opt/littleduck-web*` 與 `.env`。Node 版本由 `frontend/Dockerfile` 決定（SDP-164 已升到 24） |
| 取得 Vercel token、org／project ID | 不適用。前端與後端共用 `ghcr_*`、`hetzner_host`、`deploy_ssh_key`；前端只新增 `frontend_host`、`frontend_staging_host` |
| Preview 的 Protection Bypass | 不適用。前端 Staging 是公開的自架站台，冒煙直接打 `https://staging.stevenshih-littleduck.com` |
| 綁定正式網域 | apex 與 `staging.` 改 A 記錄指向主機 IP，Nginx 站台由 CD 渲染安裝 |
| 「GHCR 與 Vercel 都沒有新部署」（2.4） | GHCR 沒有新的 `littleduck-api`／`littleduck-web` 映像，主機容器不變 |
| 「推送非 main 分支，Vercel 不會自動部署」（4.1） | 推送非 main 分支時 Drone 只跑 CI，沒有任何 `cd-*` pipeline，主機上的前端容器不變 |
| 「Vercel Preview 冒煙通過」（4.2） | `cd-frontend-staging` 冒煙通過（`/`、`/blog` 回 2xx，`/runtime-config.js` 指向 `api.staging.`） |
| 「前端正式網域更新」（4.3） | promote 後 `cd-frontend-production` 冒煙通過，`/runtime-config.js` 指向 `api.` |
| （新增）刪除舊 Vercel 專案與殘留 secret | runbook 第 3 節第 8 步，Production 穩定後由 user 處理 |

## Non-goals

以下由 user 照 runbook 手動完成，不在 Agent 的實作範圍：

- 架設 Drone（GitHub OAuth App、`/opt/drone` 啟動、啟用 repo、Trusted、ignore forks、secret）。
- 後端與前端主機（`/opt/littleduck-api*`、`/opt/littleduck-web*` 目錄與 `.env`、`drone-deploy` 金鑰、GHCR 登入、記憶體量測）。
- DNS 記錄、certbot 核發憑證、清除 mygo 舊 `staging.` 的 A 記錄與憑證、刪除舊託管服務專案與殘留 secret。
- SDP-113 歸檔 tasks.md 的 11 項 env-blocked 實機驗收（清單見 `handoff.md`，已改寫成自架版本）。
- `docs/cicd.md`「各階段耗時」與 runbook 記憶體量測表的實際數值。
- 收斂後端 CORS（要收斂另開 issue，且 `backend/internal/middleware/**` 屬高風險檔案）。
- `www.` 子網域。
- 只對 Vercel 成立的項目（Preview alias、Protection Bypass、`NEXT_PUBLIC_API_BASE_URL` 在 Vercel 的設定）全部作廢；前端後端位址改由主機 `.env` 的 `API_BASE_URL` 於執行期決定（SDP-157）。

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `ci-server`：Drone 對外子網域由 `drone.<網域>` 定為 `drone.stevenshih-littleduck.com`。
- `backend-delivery`：Staging／Production 的對外主機名定為 `api.staging.stevenshih-littleduck.com`、`api.stevenshih-littleduck.com`，冒煙與健康檢查場景改用實際主機名。
- `frontend-delivery`：前端 Staging／Production 的冒煙與隔離場景把 `<前端 Staging 主機名>`、「正式網域」改成實際主機名，並寫明各自對應的後端主機名。

## Impact

預計修改的檔案（供判斷能否與其他 issue 並行）：

| 端 | 檔案 | 改動 |
|----|------|------|
| BE | `docs/deployment-runbook.md` | 主機名對照表、全文佔位換實際值、第 0 節 DNS、第 2 節憑證指令 |
| BE | `docs/cicd.md` | 冒煙目標、Secret 值欄、「各階段耗時」段落 |
| BE | `deploy/drone/.env.example` | `DRONE_SERVER_HOST` 預填值與註解 |
| BE | `deploy/ci/smoke-backend.sh`、`deploy/ci/smoke-frontend.sh` | 只改用法註解 |
| — | `openspec/changes/sdp-117-cicd-rollout/**` | 本變更目錄；歸檔時同步 `openspec/specs/ci-server`、`backend-delivery`、`frontend-delivery` |

- 不動 `backend/`、`frontend/`、`.drone.yml`；`openapi.yaml` 為 `paths: {}`。
- 全部任務的檔案都在後端子目錄 `backend/` 之外：依 user 決定，主對話派後端 Agent 前會暫時放寬主 repo `CLAUDE.md` 的「後端子目錄」（需要的路徑見 handoff.md），完成後改回；放寬本身不進 PR。
- 命中 `CLAUDE.md`「高風險檔案樣式」的 `deploy/**`（`deploy/drone/.env.example`、`deploy/ci/smoke-*.sh`），整個變更判定為**高風險**，PR 中要等 user 回覆「合併」。實際只改註解與範本值。
- 與其他改 `docs/cicd.md`、`docs/deployment-runbook.md`、`deploy/**` 的 issue 不宜並行；與只動 `backend/`、`frontend/` 程式碼的 issue 可並行。
