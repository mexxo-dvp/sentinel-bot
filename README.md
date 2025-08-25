# kbot

## Що нового у v1.0.3

- Оновлений Dockerfile: версія застосунку (`v1.0.3`) тепер підшивається через `APP_VERSION`.
- Multi-arch (amd64 + arm64) образи зібрані та запушені в Quay.io.
- Створено Helm-чарт (`version 0.1.1`, `appVersion v1.0.3`) із підтримкою:
  - параметра `args` для запуску Cobra-підкоманди;
  - змінної середовища `TELE_TOKEN` із Kubernetes Secret;
  - опційного параметра `arch` (за замовчуванням вимкнений);
  - probes вимкнені за замовчуванням.
- Створено GitHub Release із Helm-пакетом:  
  https://github.com/mexxo-dvp/kbot/releases/download/v1.0.3/kbot-0.1.1.tgz
- Розгортання в Kubernetes тепер використовує immutable digest образу для гарантії коректності.

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
'''
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
  --image=quay.io/paranoidlookup/kbot:v1.0.2 \
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
Chart.yaml: версія 0.1.1, appVersion: v1.0.3.

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
gh release upload v1.0.3 helm/kbot/kbot-0.1.1.tgz
```
9. Деплой через реліз
```bash
helm upgrade --install kbot \
  https://github.com/mexxo-dvp/kbot/releases/download/v1.0.3/kbot-0.1.1.tgz \
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
├── Dockerfile
├── LICENSE 
├── Makefile
├── README.md 
├── cmd 
│  ├── kbot.go 
│  ├── root.go 
│  └── version.go 
├── go.mod 
├── go.sum 
├── kbot 
└── main.go
```
main.go — точка входу, викликає CLI через cobra
cmd/ — основна логіка CLI (запуск, version, root-команда)
Makefile — автоматизація збірки
Dockerfile — контейнеризація для production
go.mod — залежності

# changelog

### v1.0.1

Початкова реалізація:

Базовий CLI на cobra та підтримка запуску Telegram-бота.

Структура проєкту з Makefile, Dockerfile та cmd-пакетом.

Вивід версії був неавтоматизований, не підшивався в білд.

### v1.0.2

Збірка Go-бінарника тепер повністю статична (CGO_ENABLED=0).

### Оновлено Dockerfile:

Базовий образ — Alpine з підтримкою сертифікатів.

Додано автоматичне підшивання версії через -ldflags.

### Makefile оновлено:

Вся крос-компіляція і Docker-збірка використовують актуальну версію із змінної.

Версію бота тепер видно у логах і команді kbot version.


## Розробник
```text
    mexxo (GitHub: mexxo-dvp)
```