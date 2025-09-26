# sentinel-bot

<a id="en"></a>
[🇬🇧 EN](#english) | [🇺🇦 UA](#українська)

---

## English

### What is this?

**sentinel-bot** is a Telegram bot written in Go (Cobra CLI + telebot). It’s instrumented with **OpenTelemetry** (end‑to‑end traces), emits **metrics**, and writes **structured JSON logs**. In Kubernetes it’s deployed via **Flux (GitOps) + SOPS**, and observed through **Fluent Bit → Loki (logs)**, **OTel Collector → Tempo (traces)**, and **GMP/Prometheus (metrics)**.

**TL;DR**: minimal bot, production‑grade observability (logs ↔ traces correlation, metrics) out of the box.

### Features

* Echo‑style handler (example) — easy to extend
* JSON logging (zerolog, UTC) with `trace_id` / `span_id`
* OpenTelemetry traces & metrics via **OTLP gRPC**
* Kubernetes‑native deploy: **Flux + SOPS** (encrypted secrets)
* Grafana Explore: open a Tempo trace directly from a Loki log line

### Runtime architecture

```
Telegram → sentinel-bot → OTLP gRPC → OTel Collector → Tempo → Grafana (traces)
                               └──────────→ /metrics → GMP/Prometheus (metrics)
sentinel-bot → stdout(JSON) → Fluent Bit → Loki → Grafana (logs)
```

### Requirements

* Go ≥ 1.21 (for local dev)
* Telegram Bot token (`TELE_TOKEN`)
* Docker / Buildx (to build images)
* Kubernetes cluster with Flux, OTel Operator, Tempo, Loki, Grafana, Fluent Bit, and GMP

### Quick start (local)

```bash
export TELE_TOKEN=your_telegram_bot_token
export APP_ENV=dev

# run locally
go run ./cmd sentinel-bot

# or containerized
docker run --rm -e TELE_TOKEN="$TELE_TOKEN" ghcr.io/mexxo-dvp/sentinel-bot:latest
```

### Configuration (environment)

| Name                          | Description                                 | Example                                           |
| ----------------------------- | ------------------------------------------- | ------------------------------------------------- |
| `TELE_TOKEN`                  | Telegram bot token                          | `123:ABC`                                         |
| `APP_ENV`                     | Environment label                           | `dev`                                             |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP gRPC endpoint (cluster OTel Collector) | `otel-collector-collector.observability.svc:4317` |
| `OTEL_SERVICE_NAME`           | Stable OTel service name                    | `sentinel-bot`                                    |
| `LOG_LEVEL`                   | `debug`/`info`/`warn`/`error`               | `info`                                            |

### Instrumentation

* **Logging**: zerolog, ISO‑8601 UTC timestamps; helper injects `trace_id`/`span_id` from context.
* **Tracing**: OTel SDK → OTLP gRPC → OTel Collector → Tempo; propagators include TraceContext + Baggage; semantic attributes (semconv) used.
* **Metrics**: OTel SDK → OTel Collector → Prometheus exporter (`/metrics`) → GMP. Example:

  ```go
  // pkg/metrics
  // counter name: sentinel_messages_total
  metrics.IncMessage(ctx)
  ```

### Kubernetes (Flux GitOps)

**GitOps paths (excerpt):**

```
clusters/gke/apps/sentinel-bot/
  ├─ namespace.yaml
  ├─ secret.enc.yaml        # SOPS, name=sentinel-bot, key=token
  └─ helmrelease.yaml       # chart/image version v0.1.9, env mapping
```

**Secret (one‑time, if needed):**

```bash
kubectl -n sentinel-bot create secret generic sentinel-bot \
  --from-literal=token="$TELE_TOKEN"
```

**HelmRelease essentials:**

* `env.TELE_TOKEN` ← `Secret name=sentinel-bot, key=token`
* `OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector-collector.observability.svc:4317`
* `OTEL_SERVICE_NAME=sentinel-bot`

**Reconcile with Flux:**

```bash
flux reconcile source git flux-system -n flux-system
flux reconcile kustomization flux-system -n flux-system --with-source
flux get helmreleases -n sentinel-bot
```

### Observability

* **Logs (Grafana → Loki / Explore):**

  ```
  {namespace="sentinel-bot", app_kubernetes_io_name="sentinel-bot"}
  ```

  Each log includes `trace_id`/`span_id`. Use the derived field to jump to Tempo.

* **Traces (Tempo):** `service.name=sentinel-bot`. A span is started per incoming message.

* **Metrics (GMP):** OTel Collector exposes `/metrics` for scraping; the app publishes `sentinel_messages_total` counter via OTLP.

### CI/CD (reference)

* **CI**: GitHub Actions → build & push image to GHCR; package & push Helm chart to `ghcr.io/mexxo-dvp/charts` (OCI).
* **CD**: Flux watches the GitOps repo and applies `HelmRelease` changes automatically.

### Project structure (excerpt)

```
.
├─ cmd/                      # Cobra commands (sentinel-bot entry)
├─ pkg/
│  ├─ telemetry/             # OTel init (traces+metrics)
│  ├─ metrics/               # Metric helpers (sentinel_messages_total)
│  └─ logging/               # zerolog configuration & helpers
├─ charts/sentinel-bot       # Helm chart (OCI)
└─ (gitops repo) clusters/gke/apps/sentinel-bot/*
```

### Troubleshooting

* **No logs** in Loki: check pod is running; Fluent Bit DaemonSet is healthy; container logs are JSON; labels match dashboard query.
* **No traces** in Tempo: verify `OTEL_EXPORTER_OTLP_ENDPOINT`; Collector receives OTLP gRPC; Grafana Tempo datasource is healthy.
* **No metrics** in GMP: ensure Collector exposes `:8888/metrics`; PodMonitoring exists; app actually increments counters.

### Security & ops notes

* Secrets managed via **SOPS**; map `sentinel-bot/token` → `TELE_TOKEN`.
* **Gitleaks** in CI/pre‑commit (not in runtime image).
* Minimal runtime: distroless base image; `.dockerignore` excludes local dev stack.

### Release & versioning

* Image: `ghcr.io/mexxo-dvp/sentinel-bot:<version>`
* Chart: `oci://ghcr.io/mexxo-dvp/charts/sentinel-bot:<version>`
* Current: **`0.1.9`**

## Changelog

### v0.1.9

* Flux GitOps: HelmRelease updated to chart/image **0.1.9**; dedicated namespace **`sentinel-bot`**.
* Observability polish: `OTEL_SERVICE_NAME=sentinel-bot`; zerolog central config; UTC timestamps; `trace_id`/`span_id` in logs.
* Metrics: added `pkg/metrics` with counter **`sentinel_messages_total`**; safe no‑op initialization.
* OTel: semconv attributes; TraceContext + Baggage propagators; OTLP endpoint default to OTel Operator service.
* Images: distroless runtime; `.dockerignore` to keep dev stack out of images; multi‑arch build.
* Helm chart: packaged & pushed to OCI (`ghcr.io/mexxo-dvp/charts`).

### v0.1.6

* Initial Flux integration for the app HelmRelease; env mapping for secrets; OTLP endpoint wired to cluster Collector.
* Basic Grafana/Loki dashboard for app logs; derived field to Tempo.
* CI basics for image + chart publishing (reference workflows).

### v0.1.5

* First release with structured logging (zerolog) and initial OTel setup in the codebase.
* Local dev stack (docker‑compose with OTel/Loki/Prometheus/Grafana) prepared for development only.

---
<a id="ua"></a>
[🇬🇧 EN](#english) | [🇺🇦 UA](#українська)

### Що це?

**sentinel-bot** — Telegram‑бот на Go (Cobra + telebot). Інструментовано **OpenTelemetry** (наскрізні трейси), **метрики** і **JSON‑логи**. У Kubernetes деплоїться через **Flux (GitOps) + SOPS**, спостережуваність: **Fluent Bit → Loki**, **Otel Collector → Tempo**, **GMP/Prometheus**.

### Функціонал

* Ехо‑хендлер (приклад), легко розширити
* JSON‑логи (UTC) з `trace_id`/`span_id`
* Трейси та метрики через OTLP gRPC
* GitOps‑деплой (Flux) + SOPS‑секрети
* Перехід з лога у трейс у Grafana

### Архітектура (рантайм)

```
Telegram → sentinel-bot → OTLP gRPC → Otel Collector → Tempo → Grafana (трейси)
                               └────────→ /metrics → GMP/Prometheus (метрики)
sentinel-bot → stdout(JSON) → Fluent Bit → Loki → Grafana (логи)
```

### Вимоги

* Go ≥ 1.21 (локально)
* `TELE_TOKEN` (Telegram)
* Docker / Buildx
* Kubernetes з Flux + стеком спостережуваності

### Швидкий старт (локально)

```bash
export TELE_TOKEN=your_telegram_bot_token
export APP_ENV=dev

go run ./cmd sentinel-bot
# або контейнером
docker run --rm -e TELE_TOKEN="$TELE_TOKEN" ghcr.io/mexxo-dvp/sentinel-bot:latest
```

### Змінні середовища

| Назва                         | Опис                                   | Приклад                                           |
| ----------------------------- | -------------------------------------- | ------------------------------------------------- |
| `TELE_TOKEN`                  | Токен Telegram‑бота                    | `123:ABC`                                         |
| `APP_ENV`                     | Середовище                             | `dev`                                             |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Ендпоінт OTLP до кластерного Collector | `otel-collector-collector.observability.svc:4317` |
| `OTEL_SERVICE_NAME`           | Імʼя сервісу в OTel                    | `sentinel-bot`                                    |
| `LOG_LEVEL`                   | `debug`/`info`/`warn`/`error`          | `info`                                            |

### Kubernetes (Flux)

```
clusters/gke/apps/sentinel-bot/
  ├─ namespace.yaml
  ├─ secret.enc.yaml (SOPS)
  └─ helmrelease.yaml (v0.1.9)
```

Реконсайл:

```bash
flux reconcile source git flux-system -n flux-system
flux reconcile kustomization flux-system -n flux-system --with-source
flux get helmreleases -n sentinel-bot
```

### Спостережуваність

* **Логи (Loki):** `{namespace="sentinel-bot", app_kubernetes_io_name="sentinel-bot"}` → є `trace_id`/`span_id`.
* **Трейси (Tempo):** `service.name=sentinel-bot`.
* **Метрики (GMP):** лічильник `sentinel_messages_total` через Collector `/metrics`.

## Зміни за версіями

* **v0.1.9**: окремий NS, Flux HelmRelease на 0.1.9, `OTEL_SERVICE_NAME`, лічильник `sentinel_messages_total`, distroless, `.dockerignore`, чарт в OCI.
* **v0.1.6**: інтеграція з Flux, секрети та OTLP, дашборд логів, CI‑скелет.
* **v0.1.5**: перший реліз із zerolog та базовим OTel.

**License:** MIT.
