# Gophkeeper

## Запуск сервера
```shell
cd deploy/docker-compose
cp .env.example .env
docker-compose up
```

Запустится сервер на порту 8080. При необходимости можно подредактировать файл .env для изменения параметров запуска.

## Запуск клиента
Подразумевается, что сервер запущен на localhost:8080

```shell
cp configs/client/config.example.yaml config.yaml
go run ./cmd/client
```
