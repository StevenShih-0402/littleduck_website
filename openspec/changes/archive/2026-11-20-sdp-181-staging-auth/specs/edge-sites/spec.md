# Spec Delta

## ADDED Requirements

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
