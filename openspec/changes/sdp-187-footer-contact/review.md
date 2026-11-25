# 程式碼審查：sdp-187-footer-contact

## 第 1 輪 — PR 前 — 2026-11-25 23:35

- **範圍**：origin/main...HEAD（第 1 輪）
- **已審到**：0101cba0d98fa9f6ed53946c0b4a5dba14ade0d4
- **審查指令**：
  - PR 前：reviewer 審查模式（已派）
- **高風險檔案**：無

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | FE | reviewer | 中 | 高 | frontend/components/front/front-shell.jsx:127-131 | 頁尾顯示聯絡 Email、兩種主題都顯示聯絡 Email › 全部 Scenario | 規格情境沒有自動化測試（前端尚無元件測試框架）；日後移除 `normal-case`、改掉 `href` 或把 Email 移進條件分支，測試仍會通過，只能靠打包產物 grep 觀察點確認 | PR 備註 |

- **結果**：通過
