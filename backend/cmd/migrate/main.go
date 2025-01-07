package main

import (
	"backend/db"
	"backend/migrations"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Ошибка загрузки .env файла: %v", err)
	}

	// Получаем переменные окружения
	mongoHost := os.Getenv("MONGO_HOST")
	mongoPort := os.Getenv("MONGO_PORT")
	dbName := os.Getenv("MONGO_DB_NAME")

	if mongoHost == "" || mongoPort == "" || dbName == "" {
		log.Fatal("Не все переменные окружения установлены")
	}

	// Формируем строку подключения к MongoDB
	mongoURI := fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort)

	// Подключение к БД
	client, err := db.Connect(mongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(nil)

	// Запускаем миграцию
	log.Println("Начинаем миграцию привычек...")
	if err := migrations.MigrateHabits(client, dbName); err != nil {
		log.Fatal(err)
	}
	log.Println("Ми��рация успешно завершена")
}
