package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/bot"
	"backend/db"
	"backend/handlers"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Удаляем загрузку .env файла
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Ошибка загрузки .env файла: %v", err)
	}

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080" // значение по умолчанию
	}

	// Получаем все переменные окружения
	botToken := os.Getenv("BOT_TOKEN")
	mongoHost := os.Getenv("MONGO_HOST")
	mongoPort := os.Getenv("MONGO_PORT")
	dbName := os.Getenv("MONGO_DB_NAME")

	if botToken == "" || mongoHost == "" || mongoPort == "" || dbName == "" {
		log.Fatal("Не все переменные окружения установлены")
	}

	// Формируем строку подключения к MongoDB
	mongoURI := fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort)

	// Подключение к БД
	client, err := db.Connect(mongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Инициализация обработчиков
	usersCollection := client.Database(dbName).Collection("users")
	historyCollection := client.Database(dbName).Collection("history")
	b, err := bot.New(botToken)
	if err != nil {
		log.Fatal(err)
	}
	handler := handlers.NewHandler(usersCollection, historyCollection, b)

	// Настройка CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Роутинг
	http.HandleFunc("/user", handler.HandleUser)
	http.HandleFunc("/user/visit", handler.HandleUpdateLastVisit)
	http.HandleFunc("/habit", handler.HandleHabit)
	http.HandleFunc("/habit/update", handler.HandleHabitUpdate)
	http.HandleFunc("/habit/delete", handler.HandleHabitDelete)
	http.HandleFunc("/create-invoice", handler.HandleCreateInvoice)
	http.HandleFunc("/habit/undo", handler.HandleHabitUndo)

	// Запуск сервера
	wrappedHandler := corsMiddleware.Handler(http.DefaultServeMux)
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(serverAddr, wrappedHandler))
}
