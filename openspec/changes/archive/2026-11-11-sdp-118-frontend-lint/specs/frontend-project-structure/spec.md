## ADDED Requirements

### Requirement: 前端提供 lint 與型別檢查腳本
前端 SHALL 在 `package.json` 提供 `lint` 與 `typecheck` 兩個 npm 腳本，在 `frontend/` 執行即可檢查全部前端原始碼（不含建置產物與 `node_modules`）；有錯誤時結束碼 MUST 非 0。

#### Scenario: 乾淨的程式碼
- **WHEN** 在 `frontend/` 對主幹的程式碼執行 `npm run lint` 與 `npm run typecheck`
- **THEN** 兩者結束碼都是 0

#### Scenario: 引入 lint 錯誤
- **WHEN** 在任一前端 `.jsx` 檔寫入違反 ESLint 規則的程式碼後執行 `npm run lint`
- **THEN** 結束碼非 0，輸出列出該檔與規則名稱

#### Scenario: 引入型別錯誤
- **WHEN** 在任一前端檔寫入型別錯誤（例如對 `string` 呼叫不存在的方法）後執行 `npm run typecheck`
- **THEN** 結束碼非 0，輸出列出錯誤位置

### Requirement: 既有違規只能逐行豁免
前端的 lint 與型別檢查 SHALL 以零錯誤為基準。無法修正的既有違規 MUST 以行內註解逐行豁免並寫明理由，MUST NOT 以關閉規則、降級規則或忽略整個檔案的方式通過。

#### Scenario: 豁免帶理由
- **WHEN** 檢視前端原始碼中的 `eslint-disable` 或 `@ts-expect-error` 註解
- **THEN** 每一則都只作用於單行或下一行，並附有說明理由的文字

#### Scenario: 設定不降低規則
- **WHEN** 檢視 `frontend/eslint.config.mjs`
- **THEN** 除了 `eslint-config-next` 預設的忽略路徑，沒有關閉規則、把 error 降為 warn，或忽略原始碼檔案的設定
