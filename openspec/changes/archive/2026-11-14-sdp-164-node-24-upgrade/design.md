# Design

## Context

升級動機見 proposal.md 的「Why」。目前 repo 指定 Node 版本的地方只有三處（以 `origin/main` 4caa687 盤點）：

| 位置 | 現值 |
|------|------|
| `.drone.yml` 的 `ci-frontend`：`install`、`lint`、`typecheck`、`test-design`、`build`、`audit` | `node:20-alpine` ×6 |
| `frontend/Dockerfile` 的 `deps`、`build`、`runner` | `node:20-alpine` ×3 |
| `frontend/package.json` 的 `devDependencies` | `@types/node: ^20.19.43` |

`package.json` 沒有 `engines`，也沒有 `.nvmrc`。docs、README、`openspec/specs/` 都沒寫死 Node 版本，所以不用同步文件。CD pipeline 的前端映像是用 `frontend/Dockerfile` 建的，Dockerfile 改了，CD 就跟著用新版，`.drone.yml` 的 CD 段落不必動。

已知限制：開發機目前裝的是 Node 20.20.0，沒有 Node 24。`CLAUDE.md`「執行驗證」用本機的 `next start` 啟動，跑的是本機 Node。

## Goals / Non-Goals

**Goals:**
- 前端 CI 和前端映像實際用 Node 24 執行，結果與升級前相同（檢查全綠、映像行為不變）。
- 用 `engines` 與 `.nvmrc` 宣告 Node 版本，之後 drift 時看得出來。

**Non-Goals:**
- 不升 Next.js、React、TypeScript 等其他相依套件，只動 `@types/node`。
- 不改其他 repo（`batting_record`、`luke`），也不更新開發機的 Node。
- 不開 `engine-strict`，避免在開發機升到 Node 24 之前 `npm ci` 直接失敗。

## Decisions

### D1：映像 tag 用 `node:24-alpine`，不釘小版本

沿用現有的 `node:<major>-alpine` 慣例，每次建置都會拿到 24.x 最新的安全修補；映像弱點掃描（trivy）會把關。
- 替代方案 `node:24.x.y-alpine3.xx`：建置可重現性較好，但要有人定期手動升版。本專案沒有 Renovate 之類的自動升版工具，釘死反而容易過期，所以不採用。

### D2：選 Node 24，不選 Node 22

issue 指定 Node 24。Node 24 目前是 Active LTS，支援到 2028-04；Node 22 已進入 Maintenance，在 2027-04 EOL。Next.js 16.4.5 要求 `>=20.9.0`，兩者都符合，選 24 下次升級前能撐比較久。

### D3：`engines.node` 寫 `">=24"`，加 `.nvmrc`，不開 engine-strict

- `>=24`：表示最低支援版本，與 Next 的 `engines` 寫法一致；不寫上限，之後升 Node 26 不必改這裡。
- `.nvmrc` 只寫大版本 `24`，nvm／fnm 都認得。放在 `frontend/` 而非 repo 根目錄，因為只有前端用 Node。
- 不開 engine-strict：在 Node 20 上 `npm ci` 只會發 `EBADENGINE` 警告。這樣開發機和執行驗證在使用者升級前仍能照常運作。

### D4：`@types/node` 跟著升到 `^24`

型別大版本要對齊執行環境，否則 `checkJs` 可能放行 Node 24 已移除的 API，或擋下新 API。本專案的 Node API 用量只有 `scripts/design-sync.mjs` 與測試中的 `node:test`、`fs`、`path`、`crypto`，預期升級後型別檢查不會產生新錯誤；若出現，就地修正，不用豁免註解繞過。

### D5：在 Node 24 上驗證，不依賴本機 Node

開發機沒有 Node 24，所以驗證走容器：
- **CI**：用 `docs/cicd.md`「在本機跑 CI」的 `drone exec --pipeline ci-frontend`，它本身就跑在 `node:24-alpine`。
- **映像**：`docker build` 之後執行 `node -v`，並跑 SDP-157 留下的冒煙腳本。

### D6：runner stage 移除映像內建的套件管理工具（實作時新增，2026-11-14 經 user 確認）

`node:24-alpine` 內建的 npm 11.19.1 自帶 `undici` 6.28.0 與 `brace-expansion` 5.0.9，trivy 對它們報出 6 個 HIGH CVE，`ci-frontend` 的 `trivy` 步驟因此失敗。Node 20 映像沒有這個問題，因為它的 npm 10.8.2 不含 undici。這些檔案屬於映像內建的 npm，不是網站自己的相依。執行期容器只跑 `node server.js`，用不到 npm。

做法：在 `runner` stage 刪除 `/usr/local/lib/node_modules/{npm,corepack}`、`/opt/yarn-*` 與 `/usr/local/bin` 底下的 `npm`、`npx`、`corepack`、`yarn`、`yarnpkg`。`deps`、`build` 兩個 stage 仍用 npm，不受影響。
- 替代方案「把這些 CVE 登記到 `frontend/.trivyignore`」：改動較小，但有弱點的檔案仍留在映像裡；而且 npm 每次升版都可能帶出新的 CVE，例外清單會越來越長。不採用。
- 代價：要進容器除錯時沒有 `npm`／`npx` 可用。這不影響現有運維，因為 runbook 的操作都不在容器內使用 npm。

## Risks / Trade-offs

- [`npm ci` 在 Node 24 下的 lockfile 相容性]：Node 24 內建 npm 11，lockfileVersion 3 可直接使用 → 由 CI 的 `install` 步驟驗證；若 npm 11 要改寫 lockfile，就以 Node 24 容器重新產生並一併 commit。
- [`@types/node` 24 引入新的型別錯誤] → 以 D4 的方式就地修正；`npm run typecheck` 在 CI 與 4.x 驗證會攔下。
- [Node 24 的 OpenSSL／預設行為改變影響 `next build` 或 runtime] → 由映像實際啟動與冒煙測試（`/`、`/blog`、`/runtime-config.js`）攔下；Staging 部署後 CD 冒煙會再驗一次。
- [本機與 CI 的 Node 版本不一致（20 對 24）] → 接受；`EBADENGINE` 警告會提醒開發者，`.nvmrc` 提供切換依據。在 Linear 備註建議開發機升級。
- [變更命中高風險樣式（`.drone.yml`、`**/Dockerfile`）] → 依流程停在等 user 審批；diff 只有映像 tag，容易審。

## Migration Plan

1. 合併後 Drone 的 `ci-frontend` 與 `cd-frontend-staging` 會用 Node 24 建置，並部署到 Staging。
2. Staging 冒煙通過後，再以既有方式 promote 到 Production。
3. **回滾**：Production 依 `docs/deployment-runbook.md` 換回先前的 `<version>-<sha>` 映像（Node 20 版），不必重建；程式碼面 revert 這個 PR 即可。

## 執行驗證觀察點

以下指令在 worktree 根目錄執行，與 `CLAUDE.md` 的通用啟動並行。映像以 3101 埠另外啟動。

- `grep -c 'node:20' .drone.yml frontend/Dockerfile` → 兩檔各為 `0`
- `grep -c 'image: node:24-alpine' .drone.yml` → `6`
- `grep -c '^FROM node:24-alpine' frontend/Dockerfile` → `3`
- `cat frontend/.nvmrc` → `24`
- `node -p "const p=require('./frontend/package.json');p.engines.node+' '+p.devDependencies['@types/node']"` → `>=24 ^24.<x>.<y>`
- `docker build -q -t ld-web-check frontend >/dev/null && docker run --rm ld-web-check node -v` → 以 `v24.` 開頭
- `docker run -d --rm --name ld-web-check -p 3101:3000 -e API_BASE_URL=http://localhost:18080 ld-web-check >/dev/null && sleep 5 && sh deploy/ci/smoke-frontend.sh http://localhost:3101 http://localhost:18080; echo "exit=$?"` → 印出三個「通過」且 `exit=0`
- `docker inspect -f '{{.Config.User}}' ld-web-check` → `node`
- `docker exec ld-web-check sh -c 'command -v npm npx corepack yarn; echo n=$?'` → 只印 `n=127`（沒有任何路徑；D6：runner 不含套件管理工具）
- `docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy:0.58.1 image --quiet --severity CRITICAL,HIGH --exit-code 1 --ignorefile /dev/null ld-web-check >/dev/null; echo "exit=$?"` → `exit=0`（在停止、刪除映像之前跑）
- `docker stop ld-web-check >/dev/null && docker image rm ld-web-check >/dev/null && echo stopped` → `stopped`（收尾，必跑）
- `docker run --rm -v "$PWD:/w" -w /w drone/cli:1.8.0 lint --trusted .drone.yml && echo ok` → `ok`

## 變異測試目標

無。本變更只動版本設定，沒有新增或修改邏輯；改由 CI 全綠與映像冒煙驗證。
