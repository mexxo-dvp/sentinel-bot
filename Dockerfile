# ---- builder ----
FROM golang:1.24.5 AS builder
WORKDIR /app

ARG APP_VERSION="dev"

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/mexxo-dvp/sentinel-bot/cmd.appVersion=${APP_VERSION}" \
    -o sentinel-bot main.go

# ---- runtime ----
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/sentinel-bot .

# Бінарник — це Cobra CLI. Subcommand передаємо через Helm args (["sentinel-bot","sentinel-bot"])
ENTRYPOINT ["/app/sentinel-bot"]
