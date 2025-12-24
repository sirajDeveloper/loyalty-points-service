# Loyalty Points Service

Система накопления баллов лояльности «Гофермарт» — это платформа для управления заказами, начислениями и балансом пользователей.

## Описание

Проект реализует HTTP API для работы с системой лояльности, включающий:
- Регистрацию и аутентификацию пользователей
- Загрузку и обработку номеров заказов
- Начисление баллов лояльности через внешний сервис
- Управление балансом и списание средств
- Взаимодействие с системой расчёта начислений

Архитектура проекта следует принципам Clean Architecture и SOLID, обеспечивая разделение на слои: domain, application, infrastructure и presentation.

## Требования

- Go 1.24 или выше
- PostgreSQL 12 или выше
- Docker (для интеграционных тестов)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (для миграций)

## Установка

```bash
git clone https://github.com/sirajDeveloper/loyalty-points-service.git
cd loyalty-points-service
go mod download
```

## Конфигурация

Сервис поддерживает конфигурацию через переменные окружения или флаги командной строки:

| Переменная окружения | Флаг | Описание | По умолчанию |
|---------------------|------|----------|--------------|
| `RUN_ADDRESS` | `-a` | Адрес и порт запуска сервиса | `localhost:8080` |
| `DATABASE_URI` | `-d` | URI подключения к PostgreSQL | - |
| `ACCRUAL_SYSTEM_ADDRESS` | `-r` | Адрес системы расчёта начислений | - |
| `JWT_SECRET` | `-j` | Секретный ключ для JWT токенов | `your-secret-key-change-in-production` |
| `JWT_EXPIRY` | - | Время жизни JWT токена | `30m` |

Пример запуска:
```bash
RUN_ADDRESS=localhost:8080 \
DATABASE_URI=postgresql://user:password@localhost:5432/gophermart?sslmode=disable \
ACCRUAL_SYSTEM_ADDRESS=http://localhost:8081 \
JWT_SECRET=your-secret-key \
./cmd/gophermart/gophermart
```

Или с флагами:
```bash
./cmd/gophermart/gophermart \
  -a localhost:8080 \
  -d postgresql://user:password@localhost:5432/gophermart?sslmode=disable \
  -r http://localhost:8081 \
  -j your-secret-key
```

## Миграции базы данных

Перед запуском сервиса необходимо применить миграции базы данных.

```bash
migrate -path migrations/gophermart \
  -database "postgresql://user:password@localhost:5432/gophermart?sslmode=disable" \
  up
```

Откат миграций:
```bash
migrate -path migrations/gophermart \
  -database "postgresql://user:password@localhost:5432/gophermart?sslmode=disable" \
  down
```

## Сборка

```bash
cd cmd/gophermart
go build -o gophermart
```

## Запуск

1. Запустите PostgreSQL и создайте базу данных:
```bash
createdb gophermart
```

2. Примените миграции (см. раздел "Миграции базы данных")

3. Запустите сервис:
```bash
./cmd/gophermart/gophermart
```

## API

- `POST /api/user/register` — регистрация пользователя
- `POST /api/user/login` — аутентификация пользователя
- `POST /api/auth/validate` — валидация JWT токена
- `GET /api/auth/health` — проверка здоровья сервиса
- `POST /api/user/orders` — загрузка номера заказа (требует аутентификации)
- `GET /api/user/orders` — получение списка заказов (требует аутентификации)
- `GET /api/user/balance` — получение текущего баланса (требует аутентификации)
- `POST /api/user/balance/withdraw` — списание средств (требует аутентификации)
- `GET /api/user/withdrawals` — получение истории списаний (требует аутентификации)

Подробная спецификация API доступна в файле [SPECIFICATION.md](SPECIFICATION.md).

## Тестирование

### Unit тесты

```bash
go test ./...
```

### Интеграционные тесты

Интеграционные тесты используют `testcontainers` для создания изолированных PostgreSQL контейнеров. Требуется Docker.

```bash
# Интеграционные тесты для gophermart
go test -tags=integration -v ./internal/gophermart/infrastructure/datastorage/postgres/...

# Интеграционные тесты для user-service
go test -tags=integration -v ./internal/user-service/infrastructure/datastorage/postgres/...

# Интеграционные тесты для HTTP клиентов
go test -v ./internal/gophermart/infrastructure/httpclient/...
```

### Покрытие кода

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Структура проекта

```
.
├── cmd/
│   ├── gophermart/          # Основной сервис (включает регистрацию, логин и бизнес-логику)
│   └── user-service/        # Отдельный сервис аутентификации (опционально)
├── internal/
│   ├── gophermart/
│   │   ├── application/     # Use cases (бизнес-логика)
│   │   ├── domain/          # Доменные модели и интерфейсы
│   │   ├── infrastructure/  # Реализации (БД, HTTP клиенты)
│   │   └── presentation/    # HTTP handlers и middleware
│   └── user-service/
│       ├── application/     # Use cases
│       ├── domain/          # Доменные модели
│       ├── infrastructure/  # Реализации
│       └── presentation/    # HTTP handlers
├── migrations/              # SQL миграции
│   └── gophermart/
└── go.mod
```

## Технологии

- [chi](https://github.com/go-chi/chi) — HTTP роутер
- [pgx](https://github.com/jackc/pgx) — PostgreSQL драйвер
- [golang-migrate](https://github.com/golang-migrate/migrate) — миграции БД
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) — JWT токены
- [testcontainers-go](https://github.com/testcontainers/testcontainers-go) — интеграционные тесты
- [testify](https://github.com/stretchr/testify) — тестирование

## Лицензия

SirajDeveloper Inc
