# edge-sites Specification

## Purpose
定義 littleduck 站台設定安裝到 Hetzner 主機上 mygo 共用 Nginx 的行為：安裝前確認站台引用的憑證在 Nginx 容器內存在，安裝後以 `nginx -t` 驗證，任一步失敗都不讓 mygo 的 Nginx 受影響。

## Requirements

### Requirement: 安裝站台前在 Nginx 容器內檢查憑證
站台安裝腳本 SHALL 在放入任何站台設定之前，對每個 `littleduck-*.conf` 引用的 `/etc/letsencrypt/live/...` 檔案，在 mygo-v2 的 Nginx 容器內確認檔案存在。檢查 MUST NOT 依賴執行帳號（`deploy`）對主機 `/etc/letsencrypt/live`、`/etc/letsencrypt/archive` 的讀取或進入權限。

#### Scenario: 憑證存在且主機憑證目錄為 700
- **WHEN** 主機 `/etc/letsencrypt/live`、`/etc/letsencrypt/archive` 為 root 專屬的 700，站台引用的憑證已核發且掛進 Nginx 容器，以 `deploy` 執行安裝腳本
- **THEN** 憑證檢查通過，站台設定放入站台目錄，`nginx -t` 通過後 reload，輸出「站台設定已更新並 reload」，結束碼為 0

#### Scenario: 憑證不存在
- **WHEN** 站台設定引用的憑證在 Nginx 容器內不存在
- **THEN** 腳本輸出 `<conf 檔名>：憑證 <路徑> 不存在，請先依 runbook 核發` 到 stderr，以非 0 結束碼中止，站台目錄內容與執行前相同，Nginx 不 reload

#### Scenario: 無法在容器內執行檢查
- **WHEN** 對 Nginx 容器執行檢查指令本身失敗（例：容器在檢查途中停止）
- **THEN** 腳本以有別於「憑證不存在」的錯誤訊息中止，結束碼非 0，站台目錄內容與執行前相同

### Requirement: 安裝後驗證失敗即還原
站台安裝腳本 SHALL 在放入設定後於 Nginx 容器內執行 `nginx -t`；通過才 reload，失敗 MUST 還原為安裝前的 `littleduck-*.conf` 並以非 0 結束碼中止。

#### Scenario: 設定語法錯誤
- **WHEN** 放入的站台設定讓 `nginx -t` 失敗
- **THEN** 站台目錄還原為安裝前的 `littleduck-*.conf`，不 reload，輸出「nginx -t 失敗，還原舊設定」

### Requirement: 安裝站台前在 Nginx 容器內檢查帳密檔
站台安裝腳本 SHALL 在放入任何站台設定之前，對每個 `littleduck-*.conf` 以 `auth_basic_user_file` 引用的檔案，在 mygo-v2 的 Nginx 容器內確認檔案存在；不存在 MUST 中止且不更動站台目錄。「無法在容器內檢查」與「檔案不存在」MUST 以不同訊息回報。

#### Scenario: 帳密檔存在
- **WHEN** 站台設定引用的帳密檔已放在主機站台目錄並掛進 Nginx 容器
- **THEN** 檢查通過，安裝照常進行（`nginx -t` 通過後 reload），結束碼為 0

#### Scenario: 帳密檔不存在
- **WHEN** 站台設定引用的帳密檔在 Nginx 容器內不存在
- **THEN** 腳本輸出 `<conf 檔名>：帳密檔 <路徑> 不存在，請先依 runbook 建立` 到 stderr，以非 0 結束碼中止，站台目錄內容與執行前相同，Nginx 不 reload

#### Scenario: 未引用帳密檔的站台
- **WHEN** 安裝的站台設定沒有 `auth_basic_user_file`（例：Production 站台）
- **THEN** 不做帳密檔檢查，行為與本變更前相同
