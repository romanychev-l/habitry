package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/handlers/follower"
	"backend/handlers/habit"
	"backend/handlers/invoice"
	"backend/handlers/ping"
	"backend/handlers/ton"
	"backend/handlers/user"

	"github.com/gin-gonic/gin"
	tgbot "github.com/go-telegram/bot"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Удаляем загрузку .env файла
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Ошибка загрузки .env файла: %v", err)
	}

	// Устанавливаем режим Gin
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Запуск в режиме release")
	} else {
		log.Println("Запуск в режиме debug")
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

	// Инициализация подключения к MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Получаем коллекции
	db := client.Database(dbName)
	usersCollection := db.Collection("users")
	historyCollection := db.Collection("history")
	habitsCollection := db.Collection("habits")
	txCollection := db.Collection("transactions")
	settingsCollection := db.Collection("settings")
	pingsCollection := db.Collection("pings")

	b, err := tgbot.New(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация обработчиков
	userHandler := user.NewHandler(usersCollection, historyCollection, habitsCollection)
	habitHandler := habit.NewHandler(habitsCollection, historyCollection, usersCollection)
	invoiceHandler := invoice.NewHandler(b)
	followerHandler := follower.NewHandler(habitsCollection, usersCollection)
	tonHandler := ton.NewHandler(usersCollection, txCollection, settingsCollection)
	pingHandler := ping.NewHandler(pingsCollection)

	// Запускаем процесс транзакций в отдельной горутине
	go runTonTransactionProcessor(tonHandler)

	// Запускаем процесс вывода средств в отдельной горутине
	go runWithdrawalsProcessor(tonHandler)

	// Настройка CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Разрешаем запросы с любого источника в режиме разработки
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-CSRF-Token", "X-Telegram-Data"},
		AllowCredentials: true,
		Debug:            true, // Включаем отладочный режим для CORS
	})

	// Настройка роутера
	r := setupGinRouter(userHandler, habitHandler, invoiceHandler, followerHandler, tonHandler, pingHandler, botToken)

	// Запуск сервера
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: corsMiddleware.Handler(r),
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Даем серверу 5 секунд для завершения активных соединений
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
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
