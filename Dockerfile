FROM golang:1.24.5 AS builder

WORKDIR /app
COPY . .

ARG VERSION=1.0.2

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X github.com/mexxo-dvp/kbot/cmd.appVersion=$VERSION" -o kbot main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/kbot .
RUN apk add --no-cache ca-certificates
ENTRYPOINT ["./kbot"]

