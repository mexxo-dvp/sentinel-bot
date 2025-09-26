<a id="en"></a>
[EN](#en) | [UA](#ua)

# EN

# sentinel-bot — Release v0.1.9

Short and to the point: a stable release with clean GitOps deployment via Flux, full integration with OTel/Loki/Tempo/GMP, and minimal runtime overhead.

## TL;DR — What’s implemented

* **GitOps/Flux**: HelmRelease for `sentinel-bot`, wired via `clusters/gke/apps/...`. The SOPS-encrypted secret is picked up by Flux.
* **Namespace**: the app now runs in its own namespace `sentinel-bot` (no dependency on `apps`).
* **Image**: distroless, multi‑arch build; `.dockerignore` keeps dev artifacts out of the image.
* **Helm Chart**: published to OCI as version `0.1.9`; image values and env vars are passed via HelmRelease.
* **Environment**: added `OTEL_SERVICE_NAME=sentinel-bot`, `APP_ENV=dev`, and `OTEL_EXPORTER_OTLP_ENDPOINT` → proper binding to the OTel Operator.
* **Logs**: JSON (zerolog, UTC). Fluent Bit collects cluster‑wide and ships to **Loki**.

  * Logs include `trace_id`/`span_id`; Grafana derived field links to Tempo.
* **Traces**: OTel SDK → **OTel Collector** (operator) → **Tempo**. OTLP endpoint is correct.
* **Metrics**: App → **OTLP** → Collector → **Prometheus exporter** (`/metrics`) → **GMP** (Google Managed Prometheus).

  * Added counter `sentinel_messages_total` (via `pkg/metrics`).
* **Code hygiene**: telemetry/metrics moved under `pkg/*`; centralized logging; ArgoCD manifests marked legacy (Flux only).

## Change details

### Infrastructure / GitOps

* **HelmRelease**: `clusters/gke/apps/sentinel-bot/helmrelease.yaml` → chart/image version `0.1.9`, secret mapping `sentinel-bot/token → TELE_TOKEN`, env for OTel.
* **Kustomize**: controller resource `apps.yaml` in Flux; `apps/` points to `sentinel-bot/` subfolder.
* **Namespace switch**: created the `sentinel-bot` namespace, moved the SOPS secret, updated references in HelmRelease. The old `apps` can be removed after prune.
* **OTel Collector**: enabled **Prometheus exporter** on `:8888`; GMP scrapes `/metrics` (PodMonitoring enabled).

### Application

* **Telemetry**: OTel (Traces+Metrics) init with semantic attributes; Baggage included in propagator.
* **Metrics**: new `pkg/metrics` package, safe no‑op if not initialized; increment happens in the text handler.
* **Logging**: centralized config (UTC, `LOG_LEVEL` controls levels, base service fields).

## Compatibility / Risks

* Breaking changes: **none**. If the Collector is temporarily unavailable, the app stays up and simply doesn’t export traces/metrics.
* Secret: expects `Secret name=sentinel-bot`, key `token` (SOPS). In HelmRelease, it maps to env `TELE_TOKEN`.

## How to deploy

1. Publish **image + chart** to GHCR (tags `0.1.9`).
2. Commit to GitOps repo with the namespace manifests for `sentinel-bot` and HelmRelease `0.1.9`.
3. Run `flux reconcile` (source + kustomization).
4. Verify the Pod in `sentinel-bot` and the HelmRelease status.

## Smoke tests

* **Logs (Grafana/Loki)**: query `{namespace="sentinel-bot", app_kubernetes_io_name="sentinel-bot"}` → JSON logs + `trace_id` present.
* **Traces (Tempo)**: click `trace_id` in Loki → opens the trace; `service.name=sentinel-bot` visible.
* **Metrics (GMP)**: find `sentinel_messages_total` or runtime metrics via OTel exporter.

## Rollback

* Bump chart/image versions back (e.g., `0.1.5`) in the HelmRelease, commit, run `flux reconcile`.

## Security/Ops notes

* **Gitleaks**: enabled in CI/pre‑commit; don’t ship it into the runtime image.
* **ArgoCD**: not used for this app (Flux is the single source of truth).

— end —

---

<a id="ua"></a>
[EN](#en) | [UA](#ua)

# sentinel-bot — Release v0.1.9

Коротко і по суті: стабільний реліз із чистим GitOps-деплоєм через Flux, повною інтеграцією Otel/Loki/Tempo/GMP, і мінімальним runtime‑овергендом.

## TL;DR — що імплементовано

* **GitOps/Flux**: HelmRelease для `sentinel-bot`, підключений шлях `clusters/gke/apps/...`. SOPS‑секрет підхоплюється Flux.
* **Namespace**: застосунок перенесено в окремий НС `sentinel-bot` (без залежності від `apps`).
* **Образ**: дистролес, multi‑arch build, `.dockerignore` → локальні dev‑артефакти не потрапляють у образ.
* **Helm Chart**: опубліковано в OCI з версією `0.1.9`; значення для образу та env‑змінні прокидано через HelmRelease.
* **Оточення**: додано `OTEL_SERVICE_NAME=sentinel-bot`, `APP_ENV=dev`, `OTEL_EXPORTER_OTLP_ENDPOINT` → правильна прив’язка до Otel Operator.
* **Логи**: JSON (zerolog, UTC). Fluent Bit збирає кластер‑вайд і шипить у **Loki**.

  * В логах є `trace_id`/`span_id`; у Grafana derived field лінкує в Tempo.
* **Трейси**: Otel SDK → **Otel Collector** (operator) → **Tempo**. Ендпоінт OTLP коректний.
* **Метрики**: App → **OTLP** → Collector → **Prometheus exporter** (`/metrics`) → **GMP**.

  * Додано лічильник `sentinel_messages_total` (через `pkg/metrics`).
* **Кодова гігієна**: винесено telemetry/metrics у `pkg/*`; логування централізовано; ArgoCD‑маніфести переведено в legacy (Flux only).

## Деталі по змінах

### Інфраструктура / GitOps

* **HelmRelease**: `clusters/gke/apps/sentinel-bot/helmrelease.yaml` → версія чарту/образу `0.1.9`, мапа секрету `sentinel-bot/token → TELE_TOKEN`, env для Otel.
* **Kustomize**: доданий керуючий ресурс `apps.yaml` у Flux; `apps/` вказує на підпапку `sentinel-bot/`.
* **Namespace switch**: створено `sentinel-bot` НС, перенесено SOPS‑секрет, оновлено посилання в HelmRelease. `apps` можна видалити після prune.
* **Otel Collector**: додано **Prometheus exporter** на `:8888`; GMP знімає `/metrics` (PodMonitoring активний).

### Додаток

* **Telemetry**: ініціалізація Otel (Trace+Metrics) з семантичними атрибутами; Baggage у пропагаторі.
* **Metrics**: новий пакет `pkg/metrics`, safe no‑op якщо не ініціалізований; інкремент у текстовому хендлері.
* **Logging**: централізована конфігурація (UTC, рівні через `LOG_LEVEL`, базові поля сервісу).

## Сумісність / ризики

* Breaking changes: **нема**. Якщо Collector тимчасово недоступний — додаток не падає, просто не експортує трейс/метрики.
* Секрет: очікується `Secret name=sentinel-bot`, ключ `token` (SOPS). В HelmRelease мапиться на env `TELE_TOKEN`.

## Як задеплоїти

1. **Образ + чарт** у GHCR (теги `0.1.9`).
2. **Коміт у GitOps** із файлами для `sentinel-bot` НС та HelmRelease `0.1.9`.
3. `flux reconcile` (source + kustomization).
4. Перевірити pod у `sentinel-bot` та статус HelmRelease.

## Смоук‑тести

* **Логи (Grafana/Loki)**: `{namespace="sentinel-bot", app_kubernetes_io_name="sentinel-bot"}` → JSON логи + `trace_id`.
* **Трейси (Tempo)**: клік по `trace_id` в Loki → відкривається трейс; у сервісі видно `service.name=sentinel-bot`.
* **Метрики (GMP)**: шукай `sentinel_messages_total` або runtime‑метрики від Otel.

## Rollback

* Змінити версії чарту/образу на попередні (напр., `0.1.5`) у HelmRelease, закомітити, `flux reconcile`.

## Нотатки безпеки/опса

* **Gitleaks**: у CI/pre‑commit; у runtime‑образ не кладемо.
* **ArgoCD**: не використовуємо для цього застосунку (Flux — single source of truth).

— end —
