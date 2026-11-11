# frontend-project-structure Specification

## Purpose
前端以標準 Next.js 專案結構組織（App Router、模組化元件、靜態資源、npm scripts），且維持原型的功能行為

## Requirements

### Requirement: 前端採用標準 Next.js 專案結構
前端 SHALL 以 Next.js（App Router）組織：頁面位於 `app/`、共用元件位於 `components/`、資料層位於 `store/`、靜態內容資料位於 `data/`、靜態資源位於 `public/`，並以 npm 管理依賴（不得使用 CDN 載入 React、Babel Standalone 或 Tailwind）。

#### Scenario: 開發伺服器可啟動
- **WHEN** 在 `frontend/` 執行 `npm install && npm run dev`
- **THEN** Next.js 開發伺服器啟動，首頁可於瀏覽器正常渲染

#### Scenario: 專案可建置
- **WHEN** 執行 `npm run build`
- **THEN** 建置成功結束，無編譯錯誤

#### Scenario: 不殘留原型載入方式
- **WHEN** 檢視專案原始碼
- **THEN** 不存在 UMD React、Babel Standalone、Tailwind CDN 的 `<script>`/`<link>` 引用，所有元件皆為 ES module

### Requirement: 頁面路由對應原型各頁
系統 SHALL 以 Next.js 檔案式路由提供原型的全部頁面：首頁 `/`、部落格 `/blog`、文章 `/blog/[id]`、作品展示 `/works`、留言板 `/guest`、後台 `/admin`。

#### Scenario: 前台頁面可直接造訪
- **WHEN** 使用者直接開啟 `/blog`、`/works`、`/guest` 任一 URL
- **THEN** 對應頁面正常渲染，站內導覽列連結可互相切換

#### Scenario: 文章頁以動態路由開啟
- **WHEN** 使用者從部落格列表點擊一篇文章
- **THEN** 導向 `/blog/<文章 id>` 並顯示該文章內容

### Requirement: 保留原型功能行為
重構後 SHALL 保留原型的使用者可見行為：雙主題切換（day 雜誌／night 終端機，選擇持久化於 localStorage）、前台與後台全部功能。文章、留言、設定、統計與贊助紀錄等業務資料 SHALL 以後端 API 作為來源與儲存，不再以 localStorage 作為這些資料的持久層（主題偏好仍持久化於 localStorage）。

#### Scenario: 主題切換與持久化
- **WHEN** 使用者切換主題後重新整理頁面
- **THEN** 主題維持上次選擇，且全站 theme tokens（背景、文字、邊框）正確套用

#### Scenario: 後台編輯反映到前台
- **WHEN** 使用者在 `/admin` 修改文章或站台設定
- **THEN** 變更經後端 API 寫入，前台頁面讀取到更新後的內容，且換裝置或清除瀏覽器資料後仍一致

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
