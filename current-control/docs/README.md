# Swagger

Файл swagger.json содержит описание API и встраивается
в приложение при сборке.

После запуска сервера:

- Swagger UI: http://localhost:8080/swagger/index.html
- JSON: http://localhost:8080/swagger/doc.json

Для работы Swagger UI нужен доступ к интернету.

Для обновления документации выполнить в корне проекта:

```bash
make swagger
```

После генерации нужно перезапустить сервер.