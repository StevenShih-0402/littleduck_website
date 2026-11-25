## Purpose

規範前台公開聯絡資訊的呈現方式：站主的聯絡 Email 要在所有前台頁面的頁尾穩定可見，兩種主題皆然，供訪客與金流服務審核方聯絡使用。

## ADDED Requirements

### Requirement: 頁尾顯示聯絡 Email
前台頁尾 SHALL 顯示站主的聯絡 Email `steven0210423@gmail.com`，並附「Contact」標示。此 Email MUST 以 `mailto:steven0210423@gmail.com` 的可點擊連結呈現，顯示文字 MUST 為小寫原文（不得被樣式轉成大寫）。此 Email MUST 由前端站台常數提供，MUST NOT 依賴後台可編輯的網站設定內容，也 MUST NOT 受功能開關控制。

#### Scenario: 一般瀏覽者在前台頁面看到 Email
- **WHEN** 未登入的訪客開啟任一前台頁面（如 `/`、`/blog`、`/blog/a1`、`/works`、`/guest`），且網站內容載入完成
- **THEN** 頁尾顯示「Contact」與 `steven0210423@gmail.com`，Email 文字為小寫

#### Scenario: 點擊 Email
- **WHEN** 訪客點擊頁尾的 Email
- **THEN** 瀏覽器以 `mailto:steven0210423@gmail.com` 開啟郵件程式，收件者為該 Email

#### Scenario: 網站設定沒有社群連結
- **WHEN** 後台網站設定的社群連結為空
- **THEN** 頁尾仍顯示聯絡 Email

### Requirement: 兩種主題都顯示聯絡 Email
聯絡 Email SHALL 在 Magazine（day）與 Terminal（night）兩種主題下都顯示，配色依各主題既有的配色規則，切換主題 MUST NOT 讓 Email 消失。

#### Scenario: 切換到 Terminal 主題
- **WHEN** 訪客在前台按主題切換鈕，從 MAG 切到 TERM
- **THEN** 頁尾仍顯示 `steven0210423@gmail.com`，且可點擊

#### Scenario: 以 Magazine 主題瀏覽
- **WHEN** 訪客以 MAG 主題瀏覽前台頁面
- **THEN** 頁尾顯示 `steven0210423@gmail.com`，且可點擊
