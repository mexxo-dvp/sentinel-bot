# ---- builder ----
FROM golang:1.24.5 AS builder
WORKDIR /app

ARG APP_VERSION="dev"

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/mexxo-dvp/kbot/cmd.appVersion=${APP_VERSION}" \
    -o kbot main.go

# ---- runtime ----
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/kbot .

# Бінарник — це Cobra CLI. Subcommand передаємо через Helm args (["kbot","kbot"])
ENTRYPOINT ["/app/kbot"]
