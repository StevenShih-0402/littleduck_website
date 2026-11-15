# Tasks

本次沒有 `[FE]` 任務，也沒有 `backend/` 的 Go 程式碼變更；以下 `[BE]` 任務改的是 `deploy/`、`docs/`（主 repo `CLAUDE.md` 已暫時把後端子目錄放寬為 `backend/、deploy/、docs/`）。不需跑 `go test`／`gofmt`，各項驗證寫在任務內，指令在 worktree 根目錄執行。

## 1. 後端：腳本

- [x] 1.1 [BE] 改 `deploy/edge/install-sites.sh`：憑證檢查改為 `docker exec "$nginx_container" test -f "$cert"`；回傳 1 維持原錯誤訊息 `<conf>：憑證 <路徑> 不存在，請先依 runbook 核發` 並 exit 1；其他非 0 以包含 conf 檔名、憑證路徑與「無法在 Nginx 容器內檢查」的訊息 exit 1；維持 POSIX `sh` 與 `set -eu`，其餘流程不變；開頭註解補上「Nginx 容器須以相同路徑掛載 /etc/letsencrypt」（變異測試目標：`deploy/edge/install-sites.sh` 的憑證檢查分流）。驗證：`sh -n deploy/edge/install-sites.sh` 無輸出、`grep -c '\[ -f "\$cert" \]' deploy/edge/install-sites.sh` 為 `0` — 憑證檢查改在 Nginx 容器內進行
- [x] 1.2 [BE] 新增 `deploy/edge/test-install-sites.sh`（依 design.md 決策 3）：以帶 mygo-v2 label 的 `nginx:1.27-alpine` 假容器、只在容器內建立的測試憑證、`EDGE_SITES_DIR` 暫存目錄，驗證三個案例：憑證存在→exit 0、輸出「站台設定已更新並 reload」、站台目錄有新 conf；憑證不存在→exit 非 0、stderr 為原錯誤訊息、站台目錄不變；無法檢查（PATH 前置 docker 包裝回 126）→exit 非 0、訊息不含「請先依 runbook 核發」、站台目錄不變。開頭偵測到既有同 label 容器就中止，`trap` 清除容器與暫存目錄。驗證：`sh deploy/edge/test-install-sites.sh` 印出三行 `PASS:` 與最後一行 `ALL PASS`、結束碼 0，結束後 `docker ps -a --filter label=com.docker.compose.project=mygo-v2 -q` 無輸出；另把 1.1 暫時改回主機 `[ -f ]` 檢查時「憑證存在」案例要 FAIL（確認測試有效，改回後不 commit 該暫時修改） — 新增本機驗證腳本，三案例通過且變異確認有效
- [x] 1.3 [BE] 兩支腳本通過 shellcheck。驗證：`docker run --rm -v "$PWD:/mnt" -w /mnt koalaman/shellcheck:v0.10.0 -s sh -S warning deploy/edge/install-sites.sh deploy/edge/test-install-sites.sh` 無輸出、結束碼 0 — 兩支腳本通過 shellcheck

## 2. 後端：文件

- [x] 2.1 [BE] 更新 `docs/cicd.md`「Nginx 站台」一節：憑證檢查在 mygo 的 Nginx 容器內進行（前置：Nginx 以相同路徑唯讀掛載 `/etc/letsencrypt`），不需放寬主機憑證目錄權限；並提一句本機驗證可跑 `sh deploy/edge/test-install-sites.sh`。驗證：`grep -n "容器內" docs/cicd.md` 命中「Nginx 站台」一節 — cicd.md 註明容器內檢查
- [x] 2.2 [BE] 更新 `docs/deployment-runbook.md`：第 1 節第 4 步（放 Nginx 站台）、第 2 節「憑證」與第 3 節「核發憑證」附近註明檢查在 Nginx 容器內，**不需要**放寬 `/etc/letsencrypt/live`、`archive` 權限（維持 certbot 預設 700）；記一行 SDP-117 期間的 `chmod 755` 暫時繞過，本變更上線後以 root `chmod 700 /etc/letsencrypt/live /etc/letsencrypt/archive` 改回。驗證：`grep -n "700" docs/deployment-runbook.md` 命中上述段落，且文件中沒有要求放寬憑證目錄權限的步驟 — runbook 註明不需放寬權限

## Workflow follow-up

- 合併後由 user 手動：主機以 root `chmod 700 /etc/letsencrypt/live /etc/letsencrypt/archive`，再以 `deploy` 實機執行 `install-sites.sh` 驗收「憑證存在→安裝並 reload」「憑證不存在→原錯誤訊息並中止」。
