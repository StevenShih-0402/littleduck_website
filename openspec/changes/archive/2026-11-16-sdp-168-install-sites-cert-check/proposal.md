## Why

`deploy/edge/install-sites.sh` 以 `deploy` 帳號在主機上用 `[ -f /etc/letsencrypt/live/<主機名>/fullchain.pem ]` 檢查憑證。certbot 預設把 `/etc/letsencrypt/live`、`/etc/letsencrypt/archive` 設為 700，`deploy` 看不到已核發的憑證，腳本誤報「憑證不存在」並中止。手動安裝 Drone 站台與每次 CD（`remote-deploy.sh` 以 `deploy` 身分呼叫 `install-sites.sh`）都會在這一步失敗（SDP-168；2026-11-14 SDP-117 實機架設時發現，目前以 `chmod 755` 暫時繞過）。

## What Changes

- `deploy/edge/install-sites.sh`：憑證檢查改在 mygo-v2 的 Nginx 容器內執行（`docker exec "$nginx_container" test -f "$cert"`），檢查的是 Nginx 實際讀得到的檔案，不再依賴主機憑證目錄對 `deploy` 開放權限。
  - 憑證不存在（`test -f` 回 1）：維持原錯誤訊息 `<conf>：憑證 <路徑> 不存在，請先依 runbook 核發`，並在放設定前中止。
  - 無法在容器內執行檢查（`docker exec` 本身失敗，回傳碼非 1）：另以明確訊息中止，不誤報成「憑證不存在」。
  - 其餘流程（找容器、備份、放設定、`nginx -t`、reload、失敗還原）不變。
- 新增 `deploy/edge/test-install-sites.sh`：本機以假的 Nginx 容器（帶 mygo-v2 compose label、只在容器內放測試憑證）驗證「憑證存在→安裝並 reload」「憑證不存在→原錯誤訊息並中止、站台目錄不變」。只在本機與 integrator 檢查時執行，不進 CD bundle。
- `docs/deployment-runbook.md`、`docs/cicd.md`：註明憑證檢查在 Nginx 容器內進行，**不需要**放寬 `/etc/letsencrypt/live`、`archive` 的權限；並說明 SDP-117 期間的 `chmod 755` 繞過可在本變更上線後改回 700。
- 非破壞性變更；沒有 API、前端或後端 Go 程式碼變更。

## Capabilities

### New Capabilities
- `edge-sites`：littleduck 站台設定安裝到 mygo 共用 Nginx 的行為契約（安裝前的憑證檢查、`nginx -t` 失敗還原）。

### Modified Capabilities
（無）

## Impact

- **預計修改的檔案範圍**（供判斷能否與其他 issue 並行）：
  - `deploy/edge/install-sites.sh`（修改）
  - `deploy/edge/test-install-sites.sh`（新增）
  - `docs/deployment-runbook.md`（修改：第 1 節 Drone 安裝、第 2 節後端憑證、第 3 節前端憑證附近的說明）
  - `docs/cicd.md`（修改：「Nginx 站台」一節）
  - 不動 `backend/`、`frontend/`、`.drone.yml`、`deploy/ci/*`。
- **風險**：命中高風險檔案樣式 `deploy/**`，整體判定為**高風險**，PR 需等 user 回覆「合併」。
- **相依**：mygo_v2_anon 的 `docker-compose.prod.yml` 須維持把 `/etc/letsencrypt:/etc/letsencrypt:ro` 掛進 Nginx 容器（路徑相同）；`deploy` 已在 docker 群組。
- **合併後**：user 手動把主機 `/etc/letsencrypt/live`、`archive` 改回 700，再以 `deploy` 實機執行 `install-sites.sh` 驗收（不屬於 PR 內任務）。
- **API**：無變動（`openapi.yaml` 為 `paths: {}`）。
