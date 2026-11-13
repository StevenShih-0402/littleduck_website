# Spec Delta

## MODIFIED Requirements

### Requirement: 訪客張貼留言
系統 SHALL 提供公開的留言張貼端點，接受留言者名稱、內容與所屬對象（如文章 id 或留言板）；新留言 MUST 以待審核（pending）狀態建立，不立即公開。所屬對象對應的功能開關停用時（留言板為 `guestbook`、文章留言為 `article_comments`），MUST 回傳 403 與錯誤碼 `FEATURE_DISABLED`，不建立留言；但請求帶有效管理員權杖（`Authorization: Bearer <jwt>`）時 MUST 不受開關限制（管理員預覽）。此端點 MUST NOT 因權杖缺失、格式錯誤或過期而回 401，這些情況一律視同訪客處理。

#### Scenario: 成功張貼留言
- **WHEN** 訪客以合法名稱與內容呼叫留言張貼端點
- **THEN** 回傳 201（或 200）並建立一筆狀態為 pending 的留言

#### Scenario: 留言內容缺漏
- **WHEN** 留言缺少必要欄位（如空白內容）
- **THEN** 回傳 400 與統一錯誤格式，不建立留言

#### Scenario: 留言板開關停用
- **WHEN** `guestbook` 停用時，訪客以空的 `articleId` 張貼留言
- **THEN** 回傳 403，錯誤碼為 `FEATURE_DISABLED`，不建立留言

#### Scenario: 文章留言開關停用
- **WHEN** `article_comments` 停用、`guestbook` 啟用時，訪客以非空的 `articleId` 張貼留言
- **THEN** 回傳 403，錯誤碼為 `FEATURE_DISABLED`，不建立留言

#### Scenario: 管理員預覽時放行
- **WHEN** `guestbook` 停用時，以有效管理員權杖張貼留言板留言
- **THEN** 回傳 201 並建立一筆狀態為 pending 的留言

#### Scenario: 權杖無效視同訪客
- **WHEN** `guestbook` 停用時，以過期或格式錯誤的權杖張貼留言板留言
- **THEN** 回傳 403，錯誤碼為 `FEATURE_DISABLED`（不是 401），不建立留言

#### Scenario: 開關啟用時權杖無效不影響張貼
- **WHEN** `guestbook` 啟用時，以過期的權杖張貼留言板留言
- **THEN** 回傳 201 並建立一筆狀態為 pending 的留言
