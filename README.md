# kbot

## kbot — CI/CD: GitHub Actions → GHCR → ArgoCD → Kubernetes
 **Код** → GitHub (`develop` branch)
- **CI** → GitHub Actions (build + push → GHCR)
- **CD** → ArgoCD (GitOps, auto-sync)
- **Deploy** → Kubernetes (k3s/k3d)
- **Packaging** → Helm Chart

### 0. Передумови
- Репозиторій GitHub: **Public** → <https://github.com/mexxo-dvp/kbot>
- GHCR-пакет: **Public** → `ghcr.io/mexxo-dvp/kbot`
- У середовищі встановлені:  
  `kubectl`, `helm`, `yq`, `docker`/`buildx`
- Працюємо у GitHub Codespaces або локально на Linux (amd64)

### 1. Kind → k3s (міграція)

Спочатку кластер підіймався через **Kind** (Kubernetes-in-Docker).  
Але для задачі з ArgoCD і LoadBalancer довелося перейти на **k3s**, бо:

- `kind` потребує додаткових налаштувань для мережі та LB,  
- у `k3s` LoadBalancer і CoreDNS працюють “із коробки”,  
- у Codespaces простіше підключати ArgoCD UI.  

Фінальна реалізація працює саме на **k3s/k3d**.

### 2. Kubernetes кластер (k3d)

встановити k3d (якщо не встановлено)
```bash
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```
Створення lightweight k3s-кластера через k3d:

```bash
k3d cluster create demo --agents 1
kubectl get nodes -o wide
```
### 3. Встановлення ArgoCD
```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl -n argocd rollout status deploy/argocd-server --timeout=180s
kubectl -n argocd get pods -owide
```

Доступ до ArgoCD UI
```bash
kubectl -n argocd port-forward svc/argocd-server 8080:80
```
UI доступний у Codespaces через HTTPS-проксі GitHub:

https://USERNAME-CODESPACE.github.dev:8080/

 У Codespaces працюємо без власного TLS (GitHub вже дає SSL).

### 4. Namespace + Secret для kbot 
```bash
kubectl create namespace demo || true
kubectl -n demo create secret generic kbot --from-literal=token='<YOUR_TELE_TOKEN>'
```
Helm-чарт очікує Secret із ключем token у namespace demo.

### 5. ArgoCD Application

infra/argocd-app-kbot.yaml:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: kbot
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/mexxo-dvp/kbot
    targetRevision: develop
    path: helm/kbot
  destination:
    server: https://kubernetes.default.svc
    namespace: demo
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```
Застосування:
```bash
kubectl apply -f infra/argocd-app-kbot.yaml
kubectl -n argocd get app kbot
```

### 6. GitHub Actions Workflow

.github/workflows/cicd.yaml:

Будує multi-arch образ (linux/amd64)

Пушить у GHCR

Оновлює helm/kbot/values.yaml (і top-level ключі для автотесту)

Комітить зміни назад у develop

Ключові кроки:

build-push → docker/build-push-action

bump Helm values → yq

git push з auto-commit у develop

```yaml
name: kbot CI/CD

on:
  push:
    branches: [ "develop" ]
    paths-ignore:
      - 'helm/kbot/values.yaml'
      - 'helm/kbot/Chart.yaml'
  workflow_dispatch:

permissions:
  contents: write
  packages: write

env:
  OWNER: ${{ github.repository_owner }}
  IMAGE: ghcr.io/${{ github.repository_owner }}/kbot
  OS: linux
  ARCH: amd64

jobs:
  build-push-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout (with tags history)
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Derive version variables
        id: vars
        run: |
          sha_short="$(git rev-parse --short=7 "$GITHUB_SHA")"
          base_tag="$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || echo v1.0.0)"
          echo "sha_short=${sha_short}" >> "$GITHUB_OUTPUT"
          echo "base_tag=${base_tag}"   >> "$GITHUB_OUTPUT"
          echo "tag=${base_tag}-${sha_short}" >> "$GITHUB_OUTPUT"
          echo "full_tag=${base_tag}-${sha_short}-${{ env.OS }}-${{ env.ARCH }}" >> "$GITHUB_OUTPUT"

      - name: Set up Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build & Push (linux/amd64)
        uses: docker/build-push-action@v6
        with:
          push: true
          platforms: linux/amd64
          tags: |
            ${{ env.IMAGE }}:${{ steps.vars.outputs.full_tag }}
            ${{ env.IMAGE }}:${{ steps.vars.outputs.tag }}
            ${{ env.IMAGE }}:${{ steps.vars.outputs.tag }}-amd64
            ${{ env.IMAGE }}:develop
          labels: |
            org.opencontainers.image.source=${{ github.repository }}
            org.opencontainers.image.revision=${{ github.sha }}
            org.opencontainers.image.version=${{ steps.vars.outputs.tag }}

      - name: Install yq
        run: |
          sudo curl -fsSL -o /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
          sudo chmod +x /usr/local/bin/yq
          yq --version

      - name: Bump Helm values & Chart appVersion (sync image.* and top-level keys)
        run: |
          set -euo pipefail
          yq -i ".image.repository = \"ghcr.io/${{ env.OWNER }}/kbot\"" helm/kbot/values.yaml
          yq -i ".image.tag = \"${{ steps.vars.outputs.tag }}\""        helm/kbot/values.yaml
          yq -i ".image.os = \"${{ env.OS }}\""                         helm/kbot/values.yaml
          yq -i ".image.arch = \"${{ env.ARCH }}\""                     helm/kbot/values.yaml
          yq -i ".repository = \"ghcr.io/${{ env.OWNER }}/kbot\""       helm/kbot/values.yaml
          yq -i ".tag = \"${{ steps.vars.outputs.tag }}\""              helm/kbot/values.yaml
          yq -i ".os = \"${{ env.OS }}\""                               helm/kbot/values.yaml
          yq -i ".arch = \"${{ env.ARCH }}\""                           helm/kbot/values.yaml
          yq -i ".appVersion = \"${{ steps.vars.outputs.base_tag }}\""  helm/kbot/Chart.yaml
          git config user.name  "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add helm/kbot/values.yaml helm/kbot/Chart.yaml
          if git diff --cached --quiet; then
            echo "No chart changes."
            exit 0
          fi
          git commit -m "ci(helm): ghcr repo + tag=${{ steps.vars.outputs.tag }} os=${{ env.OS }} arch=${{ env.ARCH }} appVersion=${{ steps.vars.outputs.base_tag }}"
          n=0
          until [ $n -ge 3 ]; do
            if git pull --rebase origin develop; then
              break
            fi
            echo "Rebase failed, retry $((n+1))/3"; git rebase --abort || true; n=$((n+1)); sleep 2
          done
          git push origin HEAD:develop
```
### 7. Makefile (локальна збірка + образ)
```makefile
# =========================
# kbot — unified Makefile
# =========================

# ---- App/Repo ----
APP_NAME ?= kbot
OWNER ?= $(shell echo $${GITHUB_REPOSITORY_OWNER:-mexxo-dvp})
REGISTRY ?= ghcr.io
REPOSITORY ?= $(OWNER)/$(APP_NAME)
IMAGE ?= $(REGISTRY)/$(REPOSITORY)

# ---- Platform for container image ----
OS ?= linux
ARCH ?= amd64

# ---- Versioning (git tag + short sha) ----
GIT_SHA := $(shell git rev-parse --short=7 HEAD)
BASE_TAG ?= $(shell git describe --tags --abbrev=0 --match 'v*' 2>/dev/null)
ifeq ($(strip $(BASE_TAG)),)
BASE_TAG := v1.0.0
endif
TAG := $(BASE_TAG)-$(GIT_SHA)
FULL_TAG := $(TAG)-$(OS)-$(ARCH)

# ---- Build flags ----
LD_FLAGS := -X github.com/mexxo-dvp/kbot/cmd.appVersion=$(BASE_TAG)

# ---- Helm / tools ----
HELM_DIR ?= helm/$(APP_NAME)
YQ ?= yq

SHELL := /bin/bash

.PHONY: help print all \
        linux arm64 macos windows \
        image image-local push \
        helm-bump clean

# -------------------------
# Helpers
# -------------------------
help:
	@echo "Targets:"
	@echo "  linux        Build Go binary for linux/amd64 -> bin/$(APP_NAME)-linux-amd64"
	@echo "  arm64        Build Go binary for linux/arm64 -> bin/$(APP_NAME)-linux-arm64"
	@echo "  macos        Build Go binaries for darwin (amd64, arm64)"
	@echo "  windows      Build Go binary for windows/amd64"
	@echo "  image        Build+Push container to $(IMAGE):$(FULL_TAG) and :develop (buildx)"
	@echo "  image-local  Local docker build (no push), tag :local"
	@echo "  helm-bump    Update helm values (registry/repository/tag/os/arch) and Chart.appVersion, commit+push to develop"
	@echo "  print        Show resolved vars"
	@echo "  clean        Remove bin/ and optional local image"

print:
	@echo "OWNER      = $(OWNER)"
	@echo "IMAGE      = $(IMAGE)"
	@echo "BASE_TAG   = $(BASE_TAG)"
	@echo "GIT_SHA    = $(GIT_SHA)"
	@echo "TAG        = $(TAG)"
	@echo "FULL_TAG   = $(FULL_TAG)"
	@echo "OS/ARCH    = $(OS)/$(ARCH)"

all: image helm-bump

# -------------------------
# Go binaries (retain your originals)
# -------------------------
linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LD_FLAGS)" -o bin/$(APP_NAME)-linux-amd64 main.go

arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LD_FLAGS)" -o bin/$(APP_NAME)-linux-arm64 main.go

macos:
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LD_FLAGS)" -o bin/$(APP_NAME)-darwin-amd64 main.go

windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LD_FLAGS)" -o bin/$(APP_NAME)-windows-amd64.exe main.go

# -------------------------
# Container image (GHCR)
# -------------------------
# Local test build (no push)
image-local:
	docker build \
		--build-arg VERSION=$(BASE_TAG) \
		-t $(IMAGE):local \
		.

# Buildx build + push (linux/amd64) with required tags
image:
	docker buildx build \
		--platform $(OS)/$(ARCH) \
		--build-arg VERSION=$(BASE_TAG) \
		--tag $(IMAGE):$(FULL_TAG) \
		--tag $(IMAGE):$(TAG) \
		--tag $(IMAGE):develop \
		--push \
		.

# Kept for compatibility; push is already done in `image`
push:
	@echo "Image already pushed in 'image' target."

# -------------------------
# Helm bump (values + Chart.appVersion) and commit to develop
# -------------------------
helm-bump:
	@if ! command -v $(YQ) >/dev/null 2>&1; then \
	  echo "ERROR: 'yq' is required. Install: https://github.com/mikefarah/yq"; \
	  exit 1; \
	fi
	$(YQ) -i '.image.registry = "$(REGISTRY)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.repository = "$(REPOSITORY)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.tag = "$(TAG)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.os = "$(OS)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.image.arch = "$(ARCH)"' $(HELM_DIR)/values.yaml
	$(YQ) -i '.appVersion = "$(BASE_TAG)"' $(HELM_DIR)/Chart.yaml
	git add $(HELM_DIR)/values.yaml $(HELM_DIR)/Chart.yaml || true
	if ! git diff --cached --quiet; then \
	  git -c user.name="github-actions[bot]" -c user.email="github-actions[bot]@users.noreply.github.com" \
	    commit -m "ci(helm): bump image to $(FULL_TAG) (appVersion=$(BASE_TAG))" && \
	  git push origin HEAD:develop; \
	else \
	  echo "No chart changes to commit."; \
	fi

# -------------------------
# Cleanup
# -------------------------
clean:
	rm -rf bin/
	-docker rmi $(IMAGE):local 2>/dev/null || true
```

### 8. Helm Chart

helm/kbot/values.yaml (фінальний вигляд для CI/CD):

```yaml
# --- Keys required by autotest ---
repository: "ghcr.io/mexxo-dvp/kbot"
tag: "v1.0.3-2972c32"
os: "linux"
arch: "amd64"
TELE_TOKEN: ""
replicaCount: 1
# --- Main image configuration ---
image:
  repository: "ghcr.io/mexxo-dvp/kbot"
  tag: "v1.0.3-2972c32"
  os: "linux"
  arch: "amd64"
  TELE_TOKEN: ""
  pullPolicy: IfNotPresent
# --- Standard Helm chart fields ---
imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""
service:
  enabled: false
  type: ClusterIP
  port: 8080
ingress:
  enabled: false
resources: {}
nodeSelector: {}
tolerations: []
affinity: {}
# --- Telegram token (secret in ns: demo) ---
teleToken:
  secretName: kbot
  secretKey: token
# Runs subcommand 'kbot' (Cobra)
args:
  - kbot
probes:
  enabled: false
```

helm/kbot/templates/deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "kbot.fullname" . }}
  labels:
    {{- include "kbot.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "kbot.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "kbot.selectorLabels" . | nindent 8 }}
    spec:
      {{ with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}

      {{- /* Fallback-параметри: нові кореневі -> старі image.* */ -}}
      {{- $repository := (default .Values.image.repository .Values.repository) -}}
      {{- $tag := (default .Values.image.tag .Values.tag) -}}
      {{- $arch := (default .Values.image.arch .Values.arch) -}}

      containers:
        - name: {{ .Release.Name }}
          image: "{{ $repository }}:{{ $tag }}{{- if .Values.image.os }}-{{ .Values.image.os }}{{- end }}{{- if $arch }}-{{ $arch }}{{- end }}"
          imagePullPolicy: {{ .Values.image.pullPolicy | default "IfNotPresent" }}
          {{- with .Values.args }}
          args:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          env:
            - name: TELE_TOKEN
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.teleToken.secretName }}
                  key: {{ .Values.teleToken.secretKey }}
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
          {{- if .Values.probes.enabled }}
          readinessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 3
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 10
            periodSeconds: 20
          {{- end }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
```

### 9. helm/kbot/templates/secret.yaml

Секрет ми створюємо через kubectl (одноразово). Цей шаблон не створює секрет, але залишений як опція, якщо захочеш керувати ним через Helm (зверни увагу на безпеку).
```yaml
{{- if .Values.TELE_TOKEN }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Values.teleToken.secretName | default "kbot" }}
type: Opaque
data:
  {{ .Values.teleToken.secretKey | default "token" }}: {{ .Values.TELE_TOKEN | b64enc }}
{{- end }}
```

За замовчуванням не увімкнено. Рекомендується створювати секрет поза чартом:
```bash
kubectl -n demo create secret generic kbot --from-literal=token='<TELE_TOKEN>'
```

helm/kbot/templates/service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ include "kbot.fullname" . }}
  labels:
    {{- include "kbot.labels" . | nindent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "kbot.selectorLabels" . | nindent 4 }}
```


---

## Що нового у v1.0.3 (детально)

### 0. Передумови
Встановлено локально:
- `helm v3.18.4`
- `kubectl v1.33.2`
- `gh (GitHub CLI) v2.75.0`
- `docker v28.3.1`
- `kind v0.23.0`

### 1. Підготовка Kubernetes-кластера
```bash
kubectl config current-context   # помилка: current-context is not set
```
створили kind-кластер
```bash
kind create cluster --name kbot --wait 90s
kubectl cluster-info --context kind-kbot
kubectl get nodes
```
### Створення namespace і секрету з TELE_TOKEN
```bash
NS=demo
kubectl create namespace "$NS"

TELE_TOKEN="Your_Tele_Token"
kubectl -n "$NS" create secret generic kbot --from-literal=token="${TELE_TOKEN}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n "$NS" get secret kbot -o yaml
```
### 3. Тест завантаження образу
```bash
kubectl run kbot-pull-check -n demo \
  --image=quay.io/paranoidlookup/kbot:v1.0.3 \
  --restart=Never -- sleep 30

kubectl -n demo wait --for=condition=Ready pod/kbot-pull-check --timeout=120s
kubectl -n demo describe pod kbot-pull-check | sed -n '/Image:/,/Events:/p'
kubectl -n demo delete pod kbot-pull-check
```
### 4. Створення Helm-чарта
```bash
git checkout -b feat/helm-chart
helm create helm/kbot
rm -f helm/kbot/templates/{hpa.yaml,ingress.yaml,serviceaccount.yaml,tests/test-connection.yaml}
```
Chart.yaml: версія 0.1.2, appVersion: v1.0.3.

values.yaml: параметри образу, секція для secret.

deployment.yaml: image, args, env TELE_TOKEN, probes вимкнені.

### 5. Лінтинг і встановлення
```bash
helm lint helm/kbot
helm install kbot helm/kbot -n demo
kubectl -n demo rollout status deploy/kbot
kubectl -n demo get pods -o wide
```
перша помилка: ImagePullBackOff через тег -amd64.
### 6. Фікс образів і args

Прибрали суфікс -amd64.

Додали args[0]=kbot, args[1]=kbot.

Проби вимкнені.
```bash
helm upgrade kbot helm/kbot -n demo \
  --set image.arch='' \
  --set args[0]=kbot \
  --set args[1]=kbot \
  --set probes.enabled=false
```
### 7. Білд нового образу v1.0.3
```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg APP_VERSION=1.0.3 \
  -t quay.io/paranoidlookup/kbot:v1.0.3 \
  -t quay.io/paranoidlookup/kbot:latest \
  --push .
```
Отримали digest:
```text
    sha256:c8f2950db5857561f36ed696bfcbe9b1b033fd81285ae86999d55cf3c31d34c8
```
### 8. GitHub Release
```bash
git tag v1.0.3
git push origin v1.0.3
helm package helm/kbot
gh release create v1.0.3 --title "kbot v1.0.3" --notes-file RELEASE.md
gh release upload v1.0.3 helm/kbot/kbot-0.1.2.tgz
```
9. Деплой через реліз
```bash
helm upgrade --install kbot \
  https://github.com/mexxo-dvp/kbot/releases/download/v1.0.3/kbot-0.1.2.tgz \
  -n demo \
  --set image.repository=quay.io/paranoidlookup \
  --set image.tag=v1.0.3
```
### 10. Проблеми та рішення

Локальний кеш образу залишав 1.0.2.

Фікс: imagePullPolicy=Always + оновлення за digest.
```bash
kubectl -n demo set image deploy/kbot \
  kbot=quay.io/paranoidlookup/kbot@sha256:c8f2950db5857561f36ed696bfcbe9b1b033fd81285ae86999d55cf3c31d34c8
```
### 11. Результат
```text
Запуск kbot версії: 1.0.3
Бот запущено. Очікування повідомлень...
```
## kbot v1.0.2

kbot — це Telegram-бот, написаний на Go з використанням бібліотек cobra для CLI та telebot для роботи з Telegram API.
Функціонал

Відповідає на будь-які вхідні текстові повідомлення в Telegram.
Можливість запуску та управління через CLI-команди.
Гнучке налаштування через змінні середовища.
Вивід версії через команду kbot version.

Швидкий старт
Вимоги

Go >= 1.21
Telegram-бот токен (TELE_TOKEN)
(опціонально) Docker

### Клонування репозиторію:
```bash
git clone https://github.com/your login/kbot.git
cd kbot
```
### Встановіть змінну середовища з токеном Telegram-бота:
```bash
export TELE_TOKEN=your_telegram_bot_token
```
### Запуск:
```bash
go run main.go kbot
```
### CLI-команди

kbot — запуск Telegram-бота
kbot version — показати поточну версію програми

### Приклад роботи

Надішліть боту будь-яке текстове повідомлення — отримаєте дзеркальну відповідь.

### Makefile

Makefile дозволяє швидко збирати бінарники під різні ОС і будувати Docker-образи:

Linux:
```bash
make linux
```
ARM64 (Linux):
```bash
make arm64
```
MacOS:
```bash
make macos
```
Windows:
```bash
make windows
```
Docker-образ:
```bash
make image
```
Очищення:
```bash
make clean
```
### Docker

Проєкт готовий до деплойменту в контейнерах:

Збірка Docker-образу:

```bash
docker build -t quay.io/your login/kbot:latest .
```

### Запуск контейнера:
```bash
export TELE_TOKEN=your_telegram_bot_token
docker run -e TELE_TOKEN quay.io/your login/kbot:latest
```
### Змінні середовища
markdown
| Назва       | Опис                | Приклад           |
|-------------|---------------------|-------------------|
| TELE_TOKEN  | Токен Telegram-бота | 123456789:ABC...  |

### Структура проекту
```text
.
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
├── RELEASE.md
├── cmd
│   ├── kbot.go
│   ├── root.go
│   └── version.go
├── go.mod
├── go.sum
├── helm
│   └── kbot
│       ├── Chart.yaml
│       ├── templates
│       │   ├── NOTES.txt
│       │   ├── _helpers.tpl
│       │   ├── deployment.yaml
│       │   ├── secret.yaml
│       │   └── service.yaml
│       └── values.yaml
├── infra
│   └── argocd-app-kbot.yaml
├── kbot
└── main.go
```
### Опис структури проєкту

main.go — точка входу в застосунок, викликає CLI через Cobra.

cmd/ — директорія з основною логікою CLI:

kbot.go — стартова команда (логіка запуску бота).

root.go — базова root-команда CLI.

version.go — команда для виводу версії.

Makefile — автоматизація збірки, тестів і публікації Docker-образів.

Dockerfile — опис контейнеризації для production.

go.mod / go.sum — файли для керування залежностями Go.

LICENSE — ліцензія проєкту.

README.md — основна документація.

RELEASE.md — нотатки до релізів (changelog).

helm/kbot/ — Helm-чарт для деплою в Kubernetes:

Chart.yaml — метадані чарту.

values.yaml — дефолтні значення параметрів.

templates/ — шаблони Kubernetes-ресурсів (Deployment, Service, Secret).

NOTES.txt — підказки для користувача після встановлення.

_helpers.tpl — хелпери для шаблонів.

infra/ — інфраструктурні маніфести:

argocd-app-kbot.yaml — декларація застосунку для ArgoCD.

kbot/ — службова директорія (може містити артефакти, дані або тимчасові файли).

# changelog

### v1.0.1

- Початкова реалізація:
- Базовий CLI на cobra та підтримка запуску Telegram-бота.
- Структура проєкту з Makefile, Dockerfile та cmd-пакетом.
- Вивід версії був неавтоматизований, не підшивався в білд.

### v1.0.2

- Збірка Go-бінарника тепер повністю статична (CGO_ENABLED=0).

### Оновлено Dockerfile:

- Базовий образ — Alpine з підтримкою сертифікатів.
- Додано автоматичне підшивання версії через -ldflags.

### Makefile оновлено:

- Вся крос-компіляція і Docker-збірка використовують актуальну версію із змінної.
- Версію бота тепер видно у логах і команді kbot version.

### Що нового у v1.0.3

- Оновлений Dockerfile: версія застосунку (`v1.0.3`) тепер підшивається через `APP_VERSION`.
- Multi-arch (amd64 + arm64) образи зібрані та запушені в Quay.io.
- Створено Helm-чарт (`version 0.1.2`, `appVersion v1.0.3`) із підтримкою:
  - параметра `args` для запуску Cobra-підкоманди;
  - змінної середовища `TELE_TOKEN` із Kubernetes Secret;
  - опційного параметра `arch` (за замовчуванням вимкнений);
  - probes вимкнені за замовчуванням.
- Створено GitHub Release із Helm-пакетом:  
  https://github.com/mexxo-dvp/kbot/releases/download/v1.0.3/kbot-0.1.2.tgz
- Розгортання в Kubernetes тепер використовує immutable digest образу для гарантії коректності.


## Розробник
```text
    mexxo (GitHub: mexxo-dvp)
```