# Текущий контроль № 1. Сервис заказов

REST API для работы с заказами. Данные хранятся в памяти.

## Запуск

Требуется Go 1.22 или выше.

```bash
go mod tidy
go run ./cmd/api
```

Сервер запускается на http://localhost:8080.

Swagger UI: http://localhost:8080/swagger/index.html

Для отображения Swagger UI нужен доступ к интернету.

## Эндпоинты

| Метод | Путь | Назначение |
|---|---|---|
| GET | /health | Проверка сервиса |
| POST | /api/v1/orders | Создание заказа |
| GET | /api/v1/orders | Список заказов |
| GET | /api/v1/orders/{id} | Получение заказа |
| PATCH | /api/v1/orders/{id}/status | Изменение статуса |
| GET | /api/v1/orders/{id}/history | История статусов |

## Примеры запросов

Создание заказа:

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"items":[{"name":"Keyboard","quantity":2,"price":500000}]}'
```

Изменение статуса:

```bash
curl -X PATCH http://localhost:8080/api/v1/orders/1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"processing"}'
```

## Бизнес-логика

Заказ должен содержать хотя бы один товар. Название не должно
быть пустым, цена и количество должны быть больше нуля.

Цена хранится в копейках. Стоимость заказа рассчитывается
как сумма произведений цены на количество.

Начальный статус заказа — new.

Допустимые переходы:

- new -> processing
- new -> cancelled
- processing -> completed
- processing -> cancelled

Статусы completed и cancelled являются конечными.
Все изменения статусов сохраняются в истории.

## Ошибки

- 400 — некорректные данные.
- 404 — заказ не найден.
- 409 — недопустимый переход статуса.
- 500 — внутренняя ошибка сервера.

Пример ответа:

```json
{
  "error": "invalid status transition",
  "code": "INVALID_STATUS_TRANSITION"
}
```

## Структура проекта

- cmd/api — запуск сервера.
- internal/order — модели, ошибки и интерфейс репозитория.
- internal/service — бизнес-логика и тесты.
- internal/repository/memory — хранение заказов.
- internal/handler — HTTP-обработчики и middleware.
- docs — Swagger.

Направление зависимостей: handler -> service -> repository
(через интерфейс).

Зависимости передаются через конструкторы.
В тестах используется тестовая реализация репозитория.

Для маршрутизации используется chi, для логирования — slog.
Логи содержат method, path, status, duration и request_id.
Recovery middleware обрабатывает паники.

## Тестирование

```bash
go test ./...
go test ./internal/service/... -cover
```

Тесты написаны с использованием testing и testify/require.

Покрытие пакета service — 100%.

## Swagger

Генерация документации:

```bash
make swagger
```