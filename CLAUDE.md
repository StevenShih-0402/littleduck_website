# CLAUDE.md

## 開發原則

- 遵循 TDD 測試先行開發原則，使用 pytest 進行測試
- 所有測試遵循 **AAA 模式（Arrange-Act-Assert）**，測試檔案在 tests 架構與 src 互為鏡像
- 遵循 SOLID 原則
- 偏好物件導向開發，可適當使用 Design Pattern
- 完成功能後，視情況同步更新測試、規格文件與 README.md
- 每個檔案的最上方都要有兩行註解：第一行是檔案路徑，第二行是簡述檔案的功能
