package invoice

import (
	"backend/models"
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-telegram/bot"
	tgbotapi "github.com/go-telegram/bot/models"
)

type Handler struct {
	bot *bot.Bot
}

func NewHandler(b *bot.Bot) *Handler {
	return &Handler{
		bot: b,
	}
}

func (h *Handler) HandleCreateInvoice(w http.ResponseWriter, r *http.Request) {
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
	log.Println(amount)
	if err != nil {
		log.Println(err)
		http.Error(w, "Неверная сумма", http.StatusBadRequest)
		return
	}

	titles := []string{"Штраф за лень", "Дань привычке", "Налог на прокрастинацию", "Ленькопошлина", "Фонд упущенных возможностей"}

	params := bot.CreateInvoiceLinkParams{
		Title:         titles[rand.Intn(len(titles))],
		Description:   "Плата равна количеству пропущенных привычек за последний день",
		Payload:       "some_payload",
		ProviderToken: "",
		Currency:      "XTR",
		Prices: []tgbotapi.LabeledPrice{
			{Label: "Some label", Amount: amount},
		},
	}

	invoiceURL, err := h.bot.CreateInvoiceLink(context.Background(), &params)

	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка при создании инвойса", http.StatusInternalServerError)
		return
	}

	response := models.InvoiceResponse{URL: invoiceURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
