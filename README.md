# GophKeeper

[![CI](https://github.com/sanek1/GophKeeper/actions/workflows/ci.yml/badge.svg)](https://github.com/sanek1/GophKeeper/actions/workflows/ci.yml)
[![CodeQL](https://github.com/sanek1/GophKeeper/actions/workflows/codeql.yml/badge.svg)](https://github.com/sanek1/GophKeeper/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/sanek1/GophKeeper/branch/main/graph/badge.svg)](https://codecov.io/gh/sanek1/GophKeeper)
[![Go Report Card](https://goreportcard.com/badge/github.com/sanek1/GophKeeper)](https://goreportcard.com/report/github.com/sanek1/GophKeeper)
[![Go Reference](https://pkg.go.dev/badge/github.com/sanek1/GophKeeper.svg)](https://pkg.go.dev/github.com/sanek1/GophKeeper)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

GophKeeper - это защищенный менеджер паролей и конфиденциальных данных, состоящий из серверной и клиентской части.

## 🚀 Функциональность

### Сервер
- REST API для доступа к хранилищу секретов
- Аутентификация пользователей и авторизация доступа к информации с использованием JWT-токенов
- Защищенное хранение данных в базе PostgreSQL
- Документация API с помощью Swagger

### Клиент
- CLI-приложение для взаимодействия с сервером
- Локальное шифрование данных с помощью мастер-пароля
- Поддержка различных типов секретов (пароли, текст, карты, заметки и т.д.)
- Кроссплатформенность (Windows, Linux, Mac OS)

## 📦 Типы хранимых данных

- `password` - пароли
- `card` - банковские карты
- `text` - текстовые данные
- `file` - файлы
- `note` - заметки
- `binary` - бинарные данные

## 🛠️ Сборка и запуск

### Предварительные требования
- Go 1.21+
- PostgreSQL
- Make (опционально)
- golangci-lint (для разработки)

### Установка инструментов разработки

```bash
# Установка всех необходимых инструментов
make install-tools
```

### Сборка

Для сборки проекта используйте Makefile:

```bash
# Показать все доступные команды
make help

# Сборка сервера и клиента для текущей платформы
make build

# Только сервер
make server

# Только клиент
make client

# Сборка клиентов для всех платформ
make client-all
```

### Запуск сервера

```bash
./bin/server
```

Или с помощью Docker:

```bash
docker-compose up -d
```

## 📋 Использование клиента

```bash
# Показать версию клиента
client version

# Регистрация нового пользователя
client register <логин> <пароль>

# Вход в систему
client login <логин> <пароль>

# Установка мастер-пароля для шифрования данных
client set-master-password <пароль>

# Создание нового секрета
client create <тип> <метаданные> <данные>

# Получение списка секретов
client list

# Получение секрета по ID
client get <id>

# Обновление секрета
client update <id> <метаданные> <данные>

# Удаление секрета
client delete <id>
```

## 🔒 Безопасность

- Аутентификация пользователей через JWT токены
- Данные шифруются на стороне клиента с использованием AES-GCM
- Каждый пользователь имеет доступ только к своим секретам
- Пароли хранятся в хешированном виде с использованием bcrypt

## 🧪 Разработка

### Запуск тестов

```bash
# Запуск всех тестов
make test

# Запуск тестов с покрытием
make test-coverage

# Запуск тестов с детектором гонок
make test-race

# Проверка покрытия тестами (минимум 80%)
make coverage-check
```

### Качество кода

```bash
# Запуск линтера
make lint

# Запуск линтера с автоисправлением
make lint-fix

# Форматирование кода
make fmt

# Полная проверка кода
make check
```

### CI/CD команды

```bash
# Команды для CI/CD
make ci-test      # Тесты с проверкой покрытия
make ci-lint      # Линтинг кода
make ci-build     # Сборка всех артефактов
make ci           # Полная проверка
```

### Docker

```bash
# Сборка Docker образа
make docker-build

# Запуск в Docker
make docker-run

# Остановка Docker
make docker-stop
```

## 📊 Покрытие тестами

Проект имеет высокое покрытие тестами:

| Модуль | Покрытие |
|--------|----------|
| Config | 95.8% |
| Repository | 89.2% |
| Database | 88.9% |
| Crypto | 82.1% |
| Auth | 82.5% |
| API | 40.1% |

Общее покрытие: **>80%**

## 🏗️ CI/CD

Проект использует GitHub Actions для автоматизации:

- **Линтинг кода** с golangci-lint
- **Запуск тестов** с проверкой покрытия
- **Сборка** для множества платформ
- **Анализ безопасности** с CodeQL и Gosec
- **Docker образы** для продакшена

## 🗂️ Генерация документации

```bash
# Генерация Swagger документации
make swagger
```

## 📝 Лицензия

MIT 