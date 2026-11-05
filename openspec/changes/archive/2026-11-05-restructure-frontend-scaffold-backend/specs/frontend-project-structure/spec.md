# frontend-project-structure

## ADDED Requirements

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
重構後 SHALL 保留原型行為：雙主題切換（day 雜誌／night 終端機，選擇持久化於 localStorage）、localStorage 資料層（`Lake` 之 seed、load、save、reset 介面不變）、前台與後台全部功能。

#### Scenario: 主題切換與持久化
- **WHEN** 使用者切換主題後重新整理頁面
- **THEN** 主題維持上次選擇，且全站 theme tokens（背景、文字、邊框）正確套用

#### Scenario: 後台編輯反映到前台
- **WHEN** 使用者在 `/admin` 修改文章或站台設定
- **THEN** 變更寫入 localStorage，前台頁面讀取到更新後的內容
