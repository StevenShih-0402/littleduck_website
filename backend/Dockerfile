# backend/Dockerfile
# multi-stage 建置：golang:1.26 編譯 → alpine 最小映像執行，組態僅由環境變數注入

# ── build stage ─────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 先複製 go.mod / go.sum 以利用 layer cache
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# ── run stage ───────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
