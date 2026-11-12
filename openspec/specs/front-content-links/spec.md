# front-content-links Specification

## Purpose
規範前台文章內文（Markdown）與頁尾社群連結的可點擊行為與網址安全規則：可點的連結要真的能導向，不安全或佔位的網址不得成為可點擊連結。

## Requirements

### Requirement: 文章內文的連結可點擊
前台渲染文章 Markdown 內文時，連結 SHALL 依網址分類渲染，且 MUST NOT 攔截點擊的預設導向行為：
- 站內路徑（以單一 `/` 開頭）在同分頁導向站內頁面；
- 外部網址（`http://`、`https://`、`mailto:`）以新分頁開啟，且 MUST 帶 `rel="noopener noreferrer"`；
- 其他網址（空字串、`#`、`javascript:`、`data:`、`//` 開頭等）MUST 只顯示連結文字，不渲染成可點擊的連結。

#### Scenario: 點擊站內連結
- **WHEN** 文章內文含 `[看文章](/blog/a1)`，訪客點擊該連結
- **THEN** 瀏覽器在同一分頁導向 `/blog/a1`

#### Scenario: 點擊外部連結
- **WHEN** 文章內文含 `[GitHub](https://github.com/example)`，訪客點擊該連結
- **THEN** 以新分頁開啟該網址，且連結帶 `rel="noopener noreferrer"`

#### Scenario: 不安全的連結網址
- **WHEN** 文章內文含 `[x](javascript:alert(1))`
- **THEN** 頁面只顯示文字「x」，不產生可點擊的連結

### Requirement: 社群連結可點擊，佔位網址顯示為不可點文字
前台頁尾 SHALL 依網站設定的社群連結渲染，網址分類規則同「文章內文的連結可點擊」，MUST NOT 攔截點擊。網址為空、`#` 或不安全時，該社群連結 MUST 仍顯示其名稱，但 MUST 是不可點擊的文字（不產生連結）。

#### Scenario: 設定了真實網址
- **WHEN** 網站設定的社群連結為 `{label: "GitHub", url: "https://github.com/example"}`
- **THEN** 頁尾顯示「GitHub」，點擊後以新分頁開啟該網址

#### Scenario: 網址仍是佔位值
- **WHEN** 某社群連結的網址為 `#` 或空字串
- **THEN** 頁尾仍顯示該社群連結的名稱，但為不可點擊的文字

#### Scenario: 網址不安全
- **WHEN** 某社群連結的網址為 `javascript:alert(1)`
- **THEN** 頁尾顯示該名稱為不可點擊的文字，不產生連結
