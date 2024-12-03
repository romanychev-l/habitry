package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"backend/bot"
	"backend/db"
	"backend/handlers"

	"github.com/rs/cors"
)

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN не установлен")
	}

	// Подключение к БД
	client, err := db.Connect("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Инициализация обработчиков
	usersCollection := client.Database("ht_db").Collection("users")
	b, err := bot.New(botToken)
	if err != nil {
		log.Fatal(err)
	}
	handler := handlers.NewHandler(usersCollection, b)

	// Настройка CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Роутинг
	http.HandleFunc("/user", handler.HandleUser)
	http.HandleFunc("/habit", handler.HandleHabit)
	http.HandleFunc("/habit/update", handler.HandleHabitUpdate)
	http.HandleFunc("/create-invoice", handler.HandleCreateInvoice)
	http.HandleFunc("/habit/undo", handler.HandleHabitUndo)

	// Запуск сервера
	wrappedHandler := corsMiddleware.Handler(http.DefaultServeMux)
	log.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", wrappedHandler))
}
