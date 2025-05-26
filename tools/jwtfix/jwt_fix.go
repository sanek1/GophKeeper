package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Константа для теста - всегда явно задаем значение
const hardcodedSecret = "gophkeeper_secret_key"

func main() {
	// 1. Создаем и проверяем токен локально для диагностики
	testToken()

	// 2. Обновляем .env файл с явным hardcoded значением
	updateEnvFile()

	// 3. Очищаем клиентские токены
	clearClientTokens()

	// 4. Выводим инструкции
	fmt.Println("\nВыполните следующие шаги:")
	fmt.Println("1. Перезапустите сервер: go run cmd/server/main.go")
	fmt.Println("2. В новом окне запустите клиент: go run cmd/client/main.go")
	fmt.Println("3. Зарегистрируйтесь и войдите заново")
}

// Проверяем, что подпись и проверка токена работают с одним ключом
func testToken() {
	fmt.Println("Тестирование JWT токенов...")

	// Создаем токен с hardcoded секретом
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "test_user",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	// Подписываем токен
	tokenString, _ := token.SignedString([]byte(hardcodedSecret))
	fmt.Println("Токен создан:", tokenString)

	// Проверяем токен с тем же секретом
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(hardcodedSecret), nil
	})

	if err != nil {
		fmt.Println("КРИТИЧЕСКАЯ ОШИБКА: Токен не прошел проверку даже с тем же ключом!")
		fmt.Println("Ошибка:", err)
	} else {
		fmt.Println("Токен успешно проверен локально.")
		fmt.Println("JWT функционирует корректно.")
	}
}

// Создаем .env файл с явно заданным значением JWT_SECRET
func updateEnvFile() {
	// Создаем содержимое .env
	envContent := "# GophKeeper server configuration\n" +
		"SERVER_PORT=8080\n" +
		"DB_HOST=localhost\n" +
		"DB_PORT=5432\n" +
		"DB_USER=postgres\n" +
		"DB_PASS=postgres\n" +
		"DB_NAME=gophkeeper\n" +
		"# Явно зафиксированный JWT_SECRET\n" +
		fmt.Sprintf("JWT_SECRET=%s\n", hardcodedSecret)

	// Записываем в .env
	err := os.WriteFile("../../.env", []byte(envContent), 0644)
	if err != nil {
		fmt.Printf("Ошибка при создании .env файла: %v\n", err)
		return
	}

	fmt.Println("\nФайл .env успешно обновлен с JWT_SECRET:", hardcodedSecret)

	// Обновляем также переменные окружения для текущего процесса
	os.Setenv("JWT_SECRET", hardcodedSecret)
	fmt.Println("Установлена переменная окружения JWT_SECRET =", hardcodedSecret)
}

// Очищаем клиентские токены
func clearClientTokens() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Ошибка при получении домашней директории: %v\n", err)
		return
	}

	gophkeeperDir := filepath.Join(homeDir, ".gophkeeper")
	if _, err := os.Stat(gophkeeperDir); os.IsNotExist(err) {
		fmt.Println("Директория клиента не найдена, токен не требует очистки")
		return
	}

	tokenFile := filepath.Join(gophkeeperDir, "token")
	if _, err := os.Stat(tokenFile); err == nil {
		if err := os.Remove(tokenFile); err != nil {
			fmt.Printf("Ошибка при удалении файла токена: %v\n", err)
			return
		}
		fmt.Println("Токен клиента успешно удален")
	} else {
		fmt.Println("Файл токена не найден, очистка не требуется")
	}
}
