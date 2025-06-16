# Docker Setup для GophKeeper

## Быстрый старт

### Запуск сервисов
```bash
docker-compose up -d --build
```

### Остановка сервисов
```bash
docker-compose down
```

### Просмотр логов
```bash
# Все сервисы
docker-compose logs

# Только приложение
docker-compose logs app

# Только база данных
docker-compose logs postgres
```

### Проверка статуса
```bash
docker-compose ps
```

## Доступ к сервисам

- **API**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **PostgreSQL**: localhost:5432

## Переменные окружения

Сервис использует следующие переменные окружения:

- `SERVER_PORT=8080` - порт сервера
- `DB_HOST=postgres` - хост базы данных
- `DB_PORT=5432` - порт базы данных
- `DB_USER=postgres` - пользователь БД
- `DB_PASS=postgres` - пароль БД
- `DB_NAME=gophkeeper` - имя базы данных
- `DB_SSL_MODE=disable` - режим SSL для БД
- `JWT_SECRET=your-super-secret-jwt-key-change-in-production` - секрет для JWT

## Очистка данных

Для полной очистки данных (включая базу данных):

```bash
docker-compose down
docker volume rm gophkeeper_postgres_data
```

## Структура

- **app** - основное приложение GophKeeper
- **postgres** - база данных PostgreSQL 16
- **postgres_data** - том для хранения данных БД

## Миграции

Миграции базы данных автоматически применяются при первом запуске PostgreSQL контейнера из папки `./migrations/`. 