#!/bin/bash

# Проверка наличия .env файла
if [ ! -f .env ]; then
  echo "Создание нового .env файла..."
  echo "# GophKeeper server configuration" > .env
  echo "SERVER_PORT=8080" >> .env
  echo "DB_HOST=localhost" >> .env
  echo "DB_PORT=5432" >> .env
  echo "DB_USER=postgres" >> .env
  echo "DB_PASS=postgres" >> .env
  echo "DB_NAME=gophkeeper" >> .env
  echo "JWT_SECRET=gophkeeper_secret_key" >> .env
  echo "Файл .env создан с настройками по умолчанию."
else
  # Проверка наличия JWT_SECRET
  if grep -q "JWT_SECRET" .env; then
    # Показываем текущее значение
    JWT_SECRET=$(grep "JWT_SECRET" .env | cut -d '=' -f2)
    echo "Текущий JWT_SECRET: $JWT_SECRET"
  else
    # Добавляем JWT_SECRET если его нет
    echo "JWT_SECRET=gophkeeper_secret_key" >> .env
    echo "JWT_SECRET добавлен в .env файл."
  fi
fi

echo "Чтобы изменить JWT_SECRET, отредактируйте файл .env"
echo "После изменения JWT_SECRET перезапустите сервер." 