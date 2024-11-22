package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/cors"
)

type InvoiceResponse struct {
	URL string `json:"url"`
}

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

		// Создаем URL для инвойса
		// invoiceURL := fmt.Sprintf(
		// 	"https://t.me/%s?start=invoice_%d",
		// 	botToken,
		// 	amount,
		// )

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

	log.Println("Сервер запущен на порту 3030")
	log.Fatal(http.ListenAndServe(":3030", handler))
}
