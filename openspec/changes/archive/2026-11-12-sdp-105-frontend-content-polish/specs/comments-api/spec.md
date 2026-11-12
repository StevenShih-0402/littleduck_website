## ADDED Requirements

### Requirement: 留言顯示日期以台灣時間計算
新留言的顯示日期（`date`，格式 `YYYY / MM / DD`）SHALL 以台灣時間（UTC+8）的當天日期產生，MUST NOT 受伺服器所在時區影響。

#### Scenario: 台灣凌晨張貼留言
- **WHEN** 訪客在台灣時間 2026-11-12 03:00（UTC 2026-11-11 19:00）張貼留言
- **THEN** 回傳的留言 `date` 為 `2026 / 11 / 12`

#### Scenario: 台灣白天張貼留言
- **WHEN** 訪客在台灣時間 2026-11-12 15:00 張貼留言
- **THEN** 回傳的留言 `date` 為 `2026 / 11 / 12`
