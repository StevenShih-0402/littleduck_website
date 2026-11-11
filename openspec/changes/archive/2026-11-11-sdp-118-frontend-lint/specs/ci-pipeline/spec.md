## MODIFIED Requirements

### Requirement: 前端 CI 門檻
前端 CI SHALL 依序執行依賴安裝、lint、型別檢查、設計同步工具測試、正式建置與依賴弱點稽核；lint 與型別檢查 MUST 在正式建置之前執行；任一步驟失敗 MUST 讓前端 check 失敗。

#### Scenario: 建置失敗
- **WHEN** PR 讓 `npm run build` 失敗
- **THEN** 前端 check 失敗

#### Scenario: 依賴有高風險弱點
- **WHEN** `npm audit` 回報 high 以上的弱點
- **THEN** 前端 check 失敗

#### Scenario: lint 錯誤
- **WHEN** PR 中有前端檔案違反 ESLint 規則（error 等級）
- **THEN** 前端 check 失敗，記錄中列出違規的檔案與規則，正式建置不執行

#### Scenario: 型別錯誤
- **WHEN** PR 中有前端檔案無法通過型別檢查
- **THEN** 前端 check 失敗，記錄中列出錯誤位置，正式建置不執行
