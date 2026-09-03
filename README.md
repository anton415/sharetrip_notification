# ShareTrip Notification

Сервис отвечает за создание и хранение уведомлений ShareTrip. В текущей версии
он предоставляет REST API для создания уведомления и получения его по
идентификатору; отправка во внешние каналы не выполняется.

## Требования

- Go 1.25;
- PostgreSQL;
- `make`.

## Команды

`Makefile` — основная система сборки и единый интерфейс запуска и проверки
проекта:

```bash
make build  # собрать bin/notification
make test   # запустить тесты
make run    # запустить сервис
make check  # форматирование, go vet, тесты и сборка
```

Для запуска сервиса задайте строку подключения к PostgreSQL:

```bash
DATABASE_URL='postgres://postgres:postgres@localhost:5432/sharetrip?sslmode=disable' make run
```

## Миграции

Команды миграций также выполняются через `Makefile` и требуют `DATABASE_URL`:

```bash
DATABASE_URL='postgres://postgres:postgres@localhost:5432/sharetrip?sslmode=disable' make migrate-up
DATABASE_URL='postgres://postgres:postgres@localhost:5432/sharetrip?sslmode=disable' make migrate-status
DATABASE_URL='postgres://postgres:postgres@localhost:5432/sharetrip?sslmode=disable' make migrate-down
```

Локальная версия инструмента миграций устанавливается автоматически при первом
вызове одной из этих команд.
