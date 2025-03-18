package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"backend/db"
	"backend/handlers/follower"
	"backend/handlers/habit"
	"backend/handlers/invoice"
	"backend/handlers/ton"
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
	tonWalletAddress := os.Getenv("TON_WALLET_ADDRESS")

	if botToken == "" || mongoHost == "" || mongoPort == "" || dbName == "" {
		log.Fatal("Не все переменные окружения установлены")
	}

	if tonWalletAddress == "" {
		log.Println("Предупреждение: TON_WALLET_ADDRESS не установлен. Платежи TON будут недоступны.")
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
	tonTxCollection := database.Collection("ton_transactions")
	settingsCollection := database.Collection("settings")

	b, err := tgbot.New(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация хендлеров
	userHandler := user.NewHandler(usersCollection, historyCollection, habitsCollection)
	habitHandler := habit.NewHandler(habitsCollection, historyCollection, usersCollection)
	invoiceHandler := invoice.NewHandler(b)
	followerHandler := follower.NewHandler(habitsCollection, usersCollection)
	tonHandler := ton.NewHandler(usersCollection, tonTxCollection, settingsCollection)

	// Запускаем процессор транзакций в отдельной горутине
	go runTonTransactionProcessor(tonHandler)

	// Запускаем процессор вывода средств в отдельной горутине
	go runWithdrawalsProcessor(tonHandler)

	// Настройка CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Разрешаем запросы с любого источника в режиме разработки
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-CSRF-Token", "X-Telegram-Data"},
		AllowCredentials: true,
		Debug:            true, // Включаем отладочный режим для CORS
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

	// Добавляем маршруты для TON платежей
	// http.HandleFunc("/api/ton/deposit", middleware.TelegramAuthMiddleware(tonHandler.HandleDeposit))
	// http.HandleFunc("/api/ton/transaction", middleware.TelegramAuthMiddleware(tonHandler.HandleCheckTransaction))

	// Добавляем новые маршруты для USDT платежей
	http.HandleFunc("/api/ton/usdt-deposit", middleware.TelegramAuthMiddleware(tonHandler.HandleUsdtDeposit))
	http.HandleFunc("/api/ton/check-usdt-transaction", middleware.TelegramAuthMiddleware(tonHandler.HandleCheckUsdtTransaction))

	// Добавляем маршрут для обработки запросов на вывод WILL
	http.HandleFunc("/api/ton/withdraw", middleware.TelegramAuthMiddleware(tonHandler.HandleWithdraw))

	// Запуск сервера
	wrappedHandler := corsMiddleware.Handler(http.DefaultServeMux)
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(serverAddr, wrappedHandler))
}

// runTonTransactionProcessor запускает периодическую обработку TON транзакций
func runTonTransactionProcessor(handler *ton.TonHandler) {
	for {
		log.Println("Проверка транзакций")
		ctx := context.Background()
		_, err := handler.CheckUsdtTransaction(ctx)
		if err != nil {
			log.Printf("Ошибка при проверке транзакций: %v", err)
		}
		log.Println("Транзакции проверены")
		time.Sleep(2 * time.Minute)
	}
}

// runWithdrawalsProcessor запускает периодическую обработку запросов на вывод
func runWithdrawalsProcessor(handler *ton.TonHandler) {
	for {
		log.Println("Начинаем обработку запросов на вывод")
		ctx := context.Background()

		err := handler.ProcessWithdrawals(ctx)
		if err != nil {
			log.Printf("Ошибка при обработке запросов на вывод: %v", err)
		}

		log.Println("Обработка запросов на вывод завершена")
		time.Sleep(2 * time.Minute)
	}
}
