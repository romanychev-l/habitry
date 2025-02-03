package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/bot"
	"backend/db"
	"backend/handlers/follower"
	"backend/handlers/habit"
	"backend/handlers/invoice"
	"backend/handlers/user"

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
	followersCollection := database.Collection("followers")

	b, err := bot.New(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация хендлеров
	userHandler := user.NewHandler(usersCollection, historyCollection, habitsCollection, followersCollection)
	habitHandler := habit.NewHandler(habitsCollection, historyCollection, followersCollection, usersCollection)
	invoiceHandler := invoice.NewHandler(b)
	followerHandler := follower.NewHandler(followersCollection)

	// Настройка CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Роутинг
	http.HandleFunc("/user", userHandler.HandleUser)
	http.HandleFunc("/user/visit", userHandler.HandleUpdateLastVisit)
	http.HandleFunc("/user/settings", userHandler.HandleSettings)
	http.HandleFunc("/user/profile", userHandler.HandleUserProfile)
	http.HandleFunc("/habit", habitHandler.HandleCreate)
	http.HandleFunc("/habit/update", habitHandler.HandleUpdate)
	http.HandleFunc("/habit/delete", habitHandler.HandleDelete)
	http.HandleFunc("/create-invoice", invoiceHandler.HandleCreateInvoice)
	http.HandleFunc("/habit/undo", habitHandler.HandleUndo)
	http.HandleFunc("/habit/join", habitHandler.HandleJoin)
	http.HandleFunc("/followers", followerHandler.HandleFollowers)
	http.HandleFunc("/habit/progress", followerHandler.HandleHabitProgress)
	http.HandleFunc("/habit/edit", habitHandler.HandleEdit)
	http.HandleFunc("/habit/followers", habitHandler.HandleGetFollowers)
	http.HandleFunc("/habit/unfollow", followerHandler.HandleUnfollow)

	// Запуск сервера
	wrappedHandler := corsMiddleware.Handler(http.DefaultServeMux)
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(serverAddr, wrappedHandler))
}
