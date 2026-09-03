# Сервис уведомлений для проекта ShareTrip

Спецификация фиксирует только требования урока «Зачем мы дробим процесс».
Дополнительные возможности и инфраструктурные решения в неё не входят.

## Цель

Создать отдельный сервис `notification`, который хранит уведомления, создаёт их
через REST API и позволяет получить уведомление по идентификатору.

На этом этапе сервис не отправляет уведомления во внешние системы.

## Объём первой версии

В первую версию входят:

- отдельный репозиторий `sharetrip_notification`;
- `Makefile` как основная система сборки и интерфейс команд проекта;
- ядро сервиса;
- REST API из двух методов;
- хранение уведомлений;
- миграция таблицы `notifications`;
- единственный статус `created`.

В первую версию не входят адаптеры для email, SMS, push-уведомлений, Telegram и
других внешних каналов. Статусы `sent`, `failed` и повторные попытки доставки
также откладываются до следующих уроков.

## Минимальный стек

- Go;
- Make;
- REST API;
- PostgreSQL.

Выбор HTTP-фреймворка, драйвера PostgreSQL и инструмента миграций является
технической реализацией, а не дополнительным требованием урока.

## Команды проекта

Сборка, тестирование, запуск и полный набор проверок выполняются через
`Makefile`:

```bash
make build
make test
DATABASE_URL='...' make run
make check
```

Миграции выполняются тем же интерфейсом; для каждой команды требуется строка
подключения к PostgreSQL в `DATABASE_URL`:

```bash
DATABASE_URL='...' make migrate-up
DATABASE_URL='...' make migrate-status
DATABASE_URL='...' make migrate-down
```

## REST API

### Формат ошибок

Все предусмотренные контрактом ошибки возвращаются как JSON с единым набором
полей:

```json
{
  "code": "VALIDATION_ERROR",
  "message": "recipient_id is required"
}
```

`code` предназначен для программной обработки, а `message` сообщает клиенту
конкретную причину ошибки. Используются следующие коды:

| Код | HTTP-статус | Значение |
| --- | --- | --- |
| `VALIDATION_ERROR` | `400 Bad Request` | Ошибка тела или обязательного поля запроса |
| `NOT_FOUND` | `404 Not Found` | Уведомление не найдено |
| `INTERNAL_ERROR` | `500 Internal Server Error` | Внутренняя ошибка сервиса |

### Создать уведомление

```http
POST /notifications
Content-Type: application/json
```

Request:

```json
{
  "recipient_id": "client-123",
  "type": "trip_published",
  "payload": {
    "trip_id": "trip-456"
  }
}
```

Response `201 Created`:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "recipient_id": "client-123",
  "type": "trip_published",
  "status": "created",
  "payload": {
    "trip_id": "trip-456"
  },
  "created_at": "2026-07-06T12:00:00Z"
}
```

Ошибки:

| HTTP-статус | `code` | `message` | Причина |
| --- | --- | --- | --- |
| `400 Bad Request` | `VALIDATION_ERROR` | `invalid request body` | Тело запроса нельзя разобрать как JSON |
| `400 Bad Request` | `VALIDATION_ERROR` | `recipient_id is required` | Не задан `recipient_id` |
| `400 Bad Request` | `VALIDATION_ERROR` | `type is required` | Не задан `type` |
| `400 Bad Request` | `VALIDATION_ERROR` | `payload is required` | Не задан `payload` |
| `500 Internal Server Error` | `INTERNAL_ERROR` | `internal server error` | Внутренняя ошибка сервиса |

### Получить уведомление

```http
GET /notifications/{id}
```

Response `200 OK`:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "recipient_id": "client-123",
  "type": "trip_published",
  "status": "created",
  "payload": {
    "trip_id": "trip-456"
  },
  "created_at": "2026-07-06T12:00:00Z"
}
```

Ошибки:

| HTTP-статус | `code` | `message` | Причина |
| --- | --- | --- | --- |
| `400 Bad Request` | `VALIDATION_ERROR` | `notification id must be a valid UUID` | `id` имеет неправильный формат |
| `404 Not Found` | `NOT_FOUND` | `notification not found` | Уведомление не найдено |
| `500 Internal Server Error` | `INTERNAL_ERROR` | `internal server error` | Внутренняя ошибка сервиса |

## Схема хранения

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    recipient_id TEXT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

| Поле | Назначение |
| --- | --- |
| `id` | UUID уведомления |
| `recipient_id` | Получатель уведомления |
| `type` | Тип уведомления, например `trip_published` |
| `status` | Состояние уведомления; в этой версии только `created` |
| `payload` | Данные уведомления в формате JSON |
| `created_at` | Дата создания уведомления |

`recipient_id` остаётся `TEXT`, как прямо указано в миграции урока. Значения
внутри `payload` сервис хранит без дополнительных требований к их формату.

## Критерии завершения

1. Миграция создаёт таблицу `notifications` по указанной схеме.
2. Корректный `POST /notifications` сохраняет уведомление и возвращает
   `201 Created` с полями из контракта.
3. Ошибка в POST-запросе возвращает `400 Bad Request`, код
   `VALIDATION_ERROR` и конкретное сообщение из контракта.
4. Внутренняя ошибка создания возвращает `500 Internal Server Error` с кодом
   `INTERNAL_ERROR`.
5. `GET /notifications/{id}` возвращает существующее уведомление с `200 OK`.
6. Запрос отсутствующего уведомления возвращает `404 Not Found` с кодом
   `NOT_FOUND`.
7. Внутренняя ошибка получения возвращает `500 Internal Server Error` с кодом
   `INTERNAL_ERROR`.
8. В реализации нет адаптеров внешних каналов доставки.
