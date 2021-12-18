# Go PostgreSQL

---

## О проекте
Этот репозиторий представляет собой реализацию REST API сервиса для просмотра видео роликов.

## Используемы технологии
- Go
- PostgreSQL
- Redis
- Jaeger

## Запуск проекта
- склонировать репозиторий
- выполнить `docker-compose up -d`

## Запуск интеграционных тестов
```bash
go test ./... -tags=integration_tests -v count=1
```

