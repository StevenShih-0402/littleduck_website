## MODIFIED Requirements

### Requirement: 祕密不寫進 repo
部署用祕密（SSH 私鑰、GHCR token、各環境 env）MUST 存在 Drone 的 repo secret 或主機上的 env 檔，MUST NOT 出現在 repo 或建置記錄的明文中。

#### Scenario: 建置記錄遮蔽祕密
- **WHEN** 流水線步驟引用 secret 並把它印出
- **THEN** Drone 的建置記錄顯示遮蔽後的值

#### Scenario: PR 建置拿不到部署祕密
- **WHEN** 由 PR 觸發的 CI pipeline 執行
- **THEN** 該 pipeline 取不到部署用的 secret
