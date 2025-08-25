# kbot
###  Що нового у v1.0.3

У цьому релізі підготовлений **kbot** до продакшн-розгортання:

- **Оновлений Dockerfile**: версія застосунку (`v1.0.3`) тепер передається через `build-arg` (`APP_VERSION`), що гарантує правильний друк версії в контейнері.
- **Multi-arch (amd64 + arm64)** білд та публікація для `v1.0.3`, Helm-чарт посилається на manifest list.
- **Оновлений Helm-чарт** (`version 0.1.1`, `appVersion "v1.0.3"`):
  - Параметр `args`: запуск Cobra-підкоманди (`["kbot","kbot"]`).
  - Змінна `TELE_TOKEN` безпечно підтягується з Kubernetes Secret.
  - Опційний параметр архітектури (`arch`) для multi-arch (за замовчуванням вимкнено).
  - Проби (healthchecks) вимкнені за замовчуванням для швидшого розгортання, їх можна вмикати після додавання health endpoint.
- **GitHub Release**: створено реліз `v1.0.3` з Helm-пакетом:

###  Інсталяція (v1.0.3)
```bash
kubectl create namespace demo || true
kubectl -n demo create secret generic kbot --from-literal=token='<ВАШ_СПРАВЖНІЙ_ТОКЕН>'

helm upgrade --install kbot \
https://github.com/mexxo-dvp/kbot/releases/download/v1.0.3/kbot-0.1.1.tgz \
-n demo \
--set image.repository=quay.io/paranoidlookup \
--set image.tag=v1.0.3
```

### devops application from scatch
### kbot v1.0.2

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