# Spec Delta

## Purpose

提供後台的身分驗證：以環境變數設定的單一管理員帳密登入、簽發與驗證 JWT，並以 middleware 保護需要管理員身分的後端端點。

## ADDED Requirements

### Requirement: 管理員以帳密登入並取得權杖
系統 SHALL 提供 `POST /api/v1/auth/login`，接受管理員帳號與密碼，與環境變數設定的單一管理員憑證比對；比對成功回傳有效期限內的 JWT 存取權杖，失敗回傳 401 與統一錯誤格式。管理員憑證與 JWT 秘鑰 MUST 來自環境變數，MUST NOT 寫死於程式碼。

#### Scenario: 帳密正確登入成功
- **WHEN** 以符合環境變數設定的管理員帳號與密碼呼叫 `POST /api/v1/auth/login`
- **THEN** 回傳 200 與一組 JWT 權杖（含過期時間）

#### Scenario: 密碼錯誤登入失敗
- **WHEN** 以錯誤密碼呼叫 `POST /api/v1/auth/login`
- **THEN** 回傳 401 與統一錯誤格式，且不簽發權杖

#### Scenario: 缺少管理員憑證設定
- **WHEN** 伺服器啟動時未設定管理員帳密或 JWT 秘鑰等必要環境變數
- **THEN** 系統於啟動階段 fail-fast 並輸出明確錯誤，不進入服務狀態

### Requirement: 以權杖保護後台端點
系統 SHALL 以 auth middleware 保護所有需要管理員身分的端點（文章／留言／設定的寫入與管理、統計、贊助紀錄清單）；請求 MUST 於 `Authorization: Bearer <token>` 帶入有效且未過期的 JWT 才能通過，否則回傳 401。公開讀取端點（如前台文章列表、公開留言）不受此限制。

#### Scenario: 帶有效權杖存取受保護端點
- **WHEN** 以有效 JWT 呼叫受保護的後台端點
- **THEN** 請求通過驗證並正常處理

#### Scenario: 無權杖或權杖過期
- **WHEN** 未帶權杖、權杖格式錯誤或已過期時呼叫受保護端點
- **THEN** 回傳 401 與統一錯誤格式，且不執行該操作

#### Scenario: 公開端點免驗證
- **WHEN** 未帶權杖呼叫前台公開讀取端點（如 `GET /api/v1/articles`）
- **THEN** 正常回傳資料，不要求驗證
