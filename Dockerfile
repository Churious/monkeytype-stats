# [Stage 1] 빌드 단계
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY main.go .
RUN go mod init mkt-stats
# SSL 인증서 문제 해결을 위해 CGO 비활성화 상태에서 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server main.go

# ---------------------------------------------------------

# [Stage 2] 실행 단계
FROM alpine:latest

# ⭐ 핵심 수정: GitHub(HTTPS) 접속을 위한 보안 인증서 설치
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]