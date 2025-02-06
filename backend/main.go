package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/db"
	"backend/handlers/follower"
	"backend/handlers/habit"
	"backend/handlers/invoice"
	"backend/handlers/user"
	"backend/middleware"

	tgbot "github.com/go-telegram/bot"
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
	database := client.Database(dbName)
	usersCollection := database.Collection("users")
	historyCollection := database.Collection("history")
	habitsCollection := database.Collection("habits")

	b, err := tgbot.New(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация хендлеров
	userHandler := user.NewHandler(usersCollection, historyCollection, habitsCollection)
	habitHandler := habit.NewHandler(habitsCollection, historyCollection, usersCollection)
	invoiceHandler := invoice.NewHandler(b)
	followerHandler := follower.NewHandler(habitsCollection, usersCollection)

	// Настройка CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Настройка маршрутов
	http.HandleFunc("/api/user", middleware.TelegramAuthMiddleware(userHandler.HandleUser))
	http.HandleFunc("/api/user/settings", middleware.TelegramAuthMiddleware(userHandler.HandleSettings))
	http.HandleFunc("/api/user/last-visit", middleware.TelegramAuthMiddleware(userHandler.HandleUpdateLastVisit))
	http.HandleFunc("/api/user/profile", middleware.TelegramAuthMiddleware(userHandler.HandleUserProfile))
	http.HandleFunc("/api/habit", middleware.TelegramAuthMiddleware(habitHandler.HandleCreate))
	http.HandleFunc("/api/habit/join", middleware.TelegramAuthMiddleware(habitHandler.HandleJoin))
	http.HandleFunc("/api/habit/click", middleware.TelegramAuthMiddleware(habitHandler.HandleUpdate))
	http.HandleFunc("/api/habit/followers", middleware.TelegramAuthMiddleware(habitHandler.HandleGetFollowers))
	http.HandleFunc("/api/habit/following", middleware.TelegramAuthMiddleware(followerHandler.HandleGetFollowing))
	http.HandleFunc("/api/habit/progress", middleware.TelegramAuthMiddleware(followerHandler.HandleHabitProgress))
	http.HandleFunc("/api/habit/unfollow", middleware.TelegramAuthMiddleware(followerHandler.HandleUnfollow))
	http.HandleFunc("/api/habit/activity", middleware.TelegramAuthMiddleware(habitHandler.HandleGetActivity))
	http.HandleFunc("/api/habit/delete", middleware.TelegramAuthMiddleware(habitHandler.HandleDelete))
	http.HandleFunc("/api/habit/undo", middleware.TelegramAuthMiddleware(habitHandler.HandleUndo))
	http.HandleFunc("/api/invoice", middleware.TelegramAuthMiddleware(invoiceHandler.HandleCreateInvoice))

	// Запуск сервера
	wrappedHandler := corsMiddleware.Handler(http.DefaultServeMux)
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(serverAddr, wrappedHandler))
}
