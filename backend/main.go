package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InvoiceResponse struct {
	URL string `json:"url"`
}

// Структуры для базы данных
type Habit struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string             `bson:"title"`
	Periodicity []string           `bson:"periodicity"` // дни недели ["monday", "wednesday", "friday"]
	CreatedAt   time.Time          `bson:"created_at"`
}

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	TelegramID int64              `bson:"telegram_id"`
	Username   string             `bson:"username"`
	FirstName  string             `bson:"first_name"`
	Habits     []Habit            `bson:"habits"`
	CreatedAt  time.Time          `bson:"created_at"`
}

var client *mongo.Client
var usersCollection *mongo.Collection

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN не установлен")
	}

	b, _ := bot.New(botToken)

	// Создаем CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // В продакшене лучше указать конкретные домены
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Проверка подключения
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Получаем коллекцию
	usersCollection = client.Database("ht_db").Collection("users")

	// Добавляем новые эндпоинты для работы с пользователями
	http.HandleFunc("/user", handleUser)
	http.HandleFunc("/habit", handleHabit)

	http.HandleFunc("/create-invoice", func(w http.ResponseWriter, r *http.Request) {
		// Включаем CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Получаем сумму из параметров запроса
		amountStr := r.URL.Query().Get("amount")
		amount, err := strconv.Atoi(amountStr)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Неверная сумма", http.StatusBadRequest)
			return
		}

		params := bot.CreateInvoiceLinkParams{
			Title:         "Some title",
			Description:   "Some description",
			Payload:       "some_payload",
			ProviderToken: "",
			Currency:      "XTR",
			Prices: []models.LabeledPrice{
				{Label: "Some label", Amount: amount},
			},
		}

		invoiceURL, err := b.CreateInvoiceLink(context.Background(), &params)

		if err != nil {
			fmt.Println(err)
			http.Error(w, "Ошибка при создании инвойса", http.StatusInternalServerError)
			return
		}

		response := InvoiceResponse{URL: invoiceURL}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Оборачиваем наш handler в CORS middleware
	handler := corsMiddleware.Handler(http.DefaultServeMux)

	log.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Обработчики для пользователей и привычек
func handleUser(w http.ResponseWriter, r *http.Request) {
	// Добавляем CORS заголовки
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Обрабатываем префлайт запросы
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "POST":
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user.CreatedAt = time.Now()
		result, err := usersCollection.InsertOne(context.Background(), user)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(result)

	case "GET":
		telegramID := r.URL.Query().Get("telegram_id")
		id, _ := strconv.ParseInt(telegramID, 10, 64)

		var user User
		err := usersCollection.FindOne(context.Background(), bson.M{"telegram_id": id}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "Пользователь не найден", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(user)
	}
}

func handleHabit(w http.ResponseWriter, r *http.Request) {
	// Добавляем CORS заголовки
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Обрабатываем префлайт запросы
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var habit Habit
	if err := json.NewDecoder(r.Body).Decode(&habit); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	telegramID := r.URL.Query().Get("telegram_id")
	id, _ := strconv.ParseInt(telegramID, 10, 64)

	habit.CreatedAt = time.Now()

	update := bson.M{
		"$push": bson.M{"habits": habit},
	}

	result, err := usersCollection.UpdateOne(
		context.Background(),
		bson.M{"telegram_id": id},
		update,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}
