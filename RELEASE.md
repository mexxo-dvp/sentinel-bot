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
