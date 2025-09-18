# ---- builder ----
FROM golang:1.24.5 AS builder
WORKDIR /app

ARG APP_VERSION="dev"

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/mexxo-dvp/sentinel-bot/cmd.appVersion=${APP_VERSION}" \
    -o sentinel-bot main.go

# ---- runtime ----
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /app/sentinel-bot .
ENV OTEL_EXPORTER_OTLP_ENDPOINT="otel-collector:4317" \
    APP_ENV="dev"
EXPOSE 8080
USER 65532:65532

ENTRYPOINT ["/app/sentinel-bot"]
