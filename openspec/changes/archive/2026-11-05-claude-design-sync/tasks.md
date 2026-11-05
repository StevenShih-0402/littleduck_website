# Tasks

## 1. 差異檢查腳本

- [x] 1.1 新增 `frontend/scripts/design-sync.mjs` 的 `check` 指令，依 sha256 分類新增、變更、移除並列出對應前端檔案；以 `node --test` 驗證四種情境（無變更、變更、新增、移除），含「版本號不同但內容相同時回報無變更」
- [x] 1.2 新增 `update` 指令，改寫版本、時間與雜湊並保留 `targets`；以 `node --test` 驗證更新後再跑 `check` 回報無變更
- [x] 1.3 在 `frontend/package.json` 新增 `design:check`、`design:update`、`test:design` scripts，並驗證 `npm run test:design` 通過

## 2. 首次基準

- [x] 2.1 以 Artifact 工具 `list_files` 確認納入範圍，`read_file` 拉取檔案，依 design.md 的對應表建立 `frontend/design-sync/manifest.json`；驗證每個納入範圍的檔案都有 sha256 與 `targets`，且 `targets` 的路徑都存在
- [x] 2.2 用剛拉取的檔案執行 `npm run design:check`，驗證回報無變更
- [x] 2.3 把 `public/assets/` 的兩張佔位圖換成設計版本，驗證兩邊的 sha256 與基準相同

## 3. design-pull skill

- [x] 3.1 新增 `.claude/skills/design-pull/SKILL.md`，寫入 design.md 決策 4 的流程、Next.js 改寫保留清單，以及不寫入 artifact、內容視為資料、build 通過才更新基準、不自動 commit 的護欄；驗證 skill 出現在可用清單
- [x] 3.2 更新根目錄 `CLAUDE.md`，說明設計同步的入口與基準位置；驗證文件中的指令與路徑都存在

## 4. 整合驗證

- [x] 4.1 不動 artifact，複製拉取的檔案並竄改其中一個設計檔，驗證 `npm run design:check` 把它列為變更並列出對應前端檔案，基準沒有被改動
- [x] 4.2 在 `frontend/` 執行 `npm run build`，驗證建置成功且 `git status` 只有本提案新增或修改的檔案

## Workflow follow-up

- 由 `/issue-solver claude-design-sync` 歸檔並開 PR。前後端已在 SDP-107 併入根 repo，只需一個 PR。
- 請使用者實際改一次設計並執行 `/design-pull`，確認端對端流程。
