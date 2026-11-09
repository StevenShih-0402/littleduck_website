<!-- README.md -->
<!-- 專案總覽：三 repo 結構說明（根 / frontend / backend）與開發入口 -->

# LittleDuck Website

個人網站專案：雙主題（雜誌 / 終端機）前台 + 後台管理 + Golang RESTful API。

## Repo 結構

本專案以單一 git repo 管理：

| 路徑 | 內容 |
|---|---|
| `./` | `openspec/` 規格與變更文件、`CLAUDE.md`、本 README |
| `./frontend/` | Next.js 16（App Router）+ React 19 + Tailwind CSS 前端 |
| `./backend/` | Go（Gin + GORM）RESTful API + PostgreSQL + Docker |

## 快速開始

### 前端

```bash
cd frontend
npm install
npm run dev          # http://localhost:3000
```

功能：前台（首頁、部落格、作品、留言板）＋ 後台管理（`/admin`）＋ 雙主題切換（day＝雜誌 / night＝終端機）。資料層目前以 localStorage 持久化，後續改接 backend API。

### 後端

```bash
cd backend
docker compose up --build    # api + PostgreSQL 16
```

- 健康檢查：<http://localhost:8080/health>
- Swagger UI：<http://localhost:8080/swagger/index.html>

詳細架構與環境變數見 [backend/README.md](backend/README.md)。

### Remote Control

想在手機上監看 Claude Code 的進度、派新任務給本機時：

```bash
.claude/scripts/remote-control.sh start   # 在背景 tmux 啟動 server；status／stop／attach 查詢、停止、接回
```

首次啟動步驟、手機端操作與限制見 [docs/remote-control.md](docs/remote-control.md)。

### CI/CD

以自架的 Drone 跑 CI（PR 與 `main` 推送時檢查前後端）與 CD（後端部署到 Hetzner、前端部署到 Vercel；先上 Staging／Preview，在 Drone promote 後才上 Production），設定在 `.drone.yml`。

- 流水線階段、門檻與本機執行 CI 的方式：[docs/cicd.md](docs/cicd.md)
- 初次建置、日常發布與回滾：[docs/deployment-runbook.md](docs/deployment-runbook.md)

## 規格文件

功能規格與變更提案以 [OpenSpec](https://github.com/Fission-AI/OpenSpec) 管理，位於 `openspec/`：

```bash
openspec list          # 進行中的變更
openspec status --change <name>
```

## 開發原則

見 [CLAUDE.md](CLAUDE.md)：TDD 測試先行（AAA 模式）、SOLID、物件導向優先。
