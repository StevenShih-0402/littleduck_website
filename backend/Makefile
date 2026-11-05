# backend/Makefile
# 常用開發指令：run、test、build、swag（重新產生 Swagger 文件）

SWAG ?= $(shell go env GOPATH)/bin/swag

.PHONY: run test build swag

run: ## 啟動開發伺服器
	go run ./cmd/server

test: ## 執行全部測試
	go test ./...

build: swag ## 產生文件並編譯 binary
	go build -o bin/server ./cmd/server

swag: ## 依 swag 註解重新產生 docs/
	$(SWAG) init -g cmd/server/main.go -o docs
