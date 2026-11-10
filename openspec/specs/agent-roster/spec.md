# agent-roster Specification

## Purpose
規定 littleduck_website 多 Agent 團隊每個角色的 Agent 名稱與顯示顏色，讓定義檔、派工指令、寫入限制 hook 與文件都用同一組名字，並和其他專案的團隊區分開來。

## Requirements

### Requirement: 角色名冊
多 Agent 團隊 SHALL 依下表為每個角色提供一個 Agent，其定義檔位於 `.claude/agents/<名稱>.md`，frontmatter 的 `name` 與 `color` MUST 與下表一致：

| 角色 | 名稱 | 顏色 |
|------|------|------|
| architect | hikaru | purple |
| backend | yasumi | pink |
| frontend | fujimori | green |
| integrator | teshima | cyan |
| reviewer | akafuku | yellow |

#### Scenario: 定義檔與名冊一致
- **WHEN** 列出 `.claude/agents/` 並讀取各檔 frontmatter
- **THEN** 恰有 `hikaru.md`、`yasumi.md`、`fujimori.md`、`teshima.md`、`akafuku.md` 五個定義檔，各自的 `name` 等於檔名、`color` 等於上表

#### Scenario: 用新名稱派工
- **WHEN** 主對話以 `subagent_type: "yasumi"` 派工
- **THEN** 啟動的是 backend 角色的 Agent

### Requirement: 舊名稱停用
`hyuse`、`chika`、`yuma`、`osamu`、`shiori` MUST NOT 再作為 Agent 名稱、定義檔名或寫入限制 hook 的角色參數出現；歷史歸檔（`openspec/changes/archive/**`）不在此限。

#### Scenario: 現行檔案沒有舊名稱
- **WHEN** 在 repo 中排除 `openspec/changes/archive/` 與 `.git/` 後以不分大小寫搜尋五個舊名稱
- **THEN** 沒有任何結果

#### Scenario: 舊名呼叫失效
- **WHEN** 主對話以 `subagent_type: "hyuse"` 派工
- **THEN** 找不到該 Agent，不會啟動任何角色

### Requirement: 寫入限制 hook 認得新名稱
寫入限制 hook SHALL 只接受 `hikaru`、`teshima`、`yasumi`、`fujimori` 作為角色參數，各名稱的寫入範圍 MUST 與改名前對應角色相同；收到其他參數 MUST 拒絕寫入。

#### Scenario: backend 角色只能寫 backend
- **WHEN** 以角色 `yasumi` 寫入 `frontend/` 底下的檔案
- **THEN** hook 拒絕，訊息指出該檔由 Fujimori 負責

#### Scenario: architect 可寫 openspec
- **WHEN** 以角色 `hikaru` 寫入 `openspec/changes/<id>/proposal.md`
- **THEN** hook 放行

#### Scenario: 舊角色參數被拒
- **WHEN** 以角色 `chika` 呼叫 hook
- **THEN** hook 以未知角色拒絕，並列出可接受的新名稱
