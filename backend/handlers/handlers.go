package handlers

import (
	"backend/models"
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	tgbotapi "github.com/go-telegram/bot/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	usersCollection *mongo.Collection
	bot             *bot.Bot
}

func NewHandler(collection *mongo.Collection, b *bot.Bot) *Handler {
	return &Handler{
		usersCollection: collection,
		bot:             b,
	}
}

func (h *Handler) HandleUser(w http.ResponseWriter, r *http.Request) {
	log.Println("handleUser", r.Method)
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
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Printf("Ошибка декодирования JSON: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user.CreatedAt = time.Now()
		user.Credit = 0
		user.LastVisit = time.Now().Format("2006-01-02")
		user.History = []models.DayHistory{}
		if user.Habits == nil {
			user.Habits = []models.Habit{}
		}

		result, err := h.usersCollection.InsertOne(context.Background(), user)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(result)

	case "GET":
		telegramID := r.URL.Query().Get("telegram_id")
		id, _ := strconv.ParseInt(telegramID, 10, 64)

		var user models.User
		err := h.usersCollection.FindOne(context.Background(), bson.M{"telegram_id": id}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "Пользователь не найден", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		today := time.Now().Format("2006-01-02")
		if user.LastVisit != today {
			// Проверяем историю за вчерашний день
			yesterday := time.Now().AddDate(0, 0, -1)
			yesterdayStr := yesterday.Format("2006-01-02")
			yesterdayIndex := int(yesterday.Weekday())
			if yesterdayIndex == 0 {
				yesterdayIndex = 6
			} else {
				yesterdayIndex--
			}

			// Получаем все привычки, которые должны были быть выполнены вчера
			scheduledHabits := make(map[string]bool)
			for _, habit := range user.Habits {
				for _, day := range habit.Days {
					if day == yesterdayIndex {
						scheduledHabits[habit.ID] = false
						break
					}
				}
			}

			// Проверяем, какие из них были выполнены
			for _, dayHistory := range user.History {
				if dayHistory.Date == yesterdayStr {
					for _, habitHistory := range dayHistory.Habits {
						delete(scheduledHabits, habitHistory.HabitID)
					}
					break
				}
			}

			// Количество невыполненных привычек - это количество оставшихся в мапе
			missedHabits := len(scheduledHabits)

			// Обновляем кредит и дату последнего визита
			update := bson.M{
				"$set": bson.M{
					"credit":     missedHabits,
					"last_visit": today,
				},
			}

			_, err := h.usersCollection.UpdateOne(
				context.Background(),
				bson.M{"telegram_id": id},
				update,
			)

			if err != nil {
				log.Printf("Ошибка при обновлении кредита: %v", err)
			}

			user.Credit = missedHabits
			// user.LastVisit = today
		}

		// Фильтруем привычки для текущего дня
		todayHabits := []models.Habit{}
		for _, habit := range user.Habits {
			weekday := int(time.Now().Weekday())
			// Преобразуем воскресенье (0) в 6, а остальные дни уменьшаем на 1
			if weekday == 0 {
				weekday = 6
			} else {
				weekday--
			}

			for _, day := range habit.Days {
				if day == weekday {
					todayHabits = append(todayHabits, habit)
					break
				}
			}
		}
		user.Habits = todayHabits

		json.NewEncoder(w).Encode(user)
	}
}

func (h *Handler) HandleHabit(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHabit", r.Method)
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

	var habitRequest models.HabitRequest
	if err := json.NewDecoder(r.Body).Decode(&habitRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	habitRequest.Habit.CreatedAt = time.Now()

	update := bson.M{
		"$push": bson.M{"habits": habitRequest.Habit},
	}

	result, err := h.usersCollection.UpdateOne(
		context.Background(),
		bson.M{"telegram_id": habitRequest.TelegramID},
		update,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func (h *Handler) HandleHabitUpdate(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHabitUpdate", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, `{"message": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	var habitRequest models.HabitRequest
	if err := json.NewDecoder(r.Body).Decode(&habitRequest); err != nil {
		http.Error(w, `{"message": "Неверный формат данных"}`, http.StatusBadRequest)
		return
	}

	// Получаем текущую привычку из БД для проверки last_click_date
	var user models.User
	err := h.usersCollection.FindOne(
		context.Background(),
		bson.M{
			"telegram_id": habitRequest.TelegramID,
			"habits.id":   habitRequest.Habit.ID,
		},
	).Decode(&user)

	if err != nil {
		http.Error(w, `{"message": "Привычка не найдена"}`, http.StatusNotFound)
		return
	}

	var currentHabit models.Habit
	for _, h := range user.Habits {
		if h.ID == habitRequest.Habit.ID {
			currentHabit = h
			break
		}
	}
	log.Println(currentHabit)

	// Проверяем, был ли предыдущий клик в предыдущий разрешенный день
	var newStreak int
	if currentHabit.LastClickDate != "" {
		lastClickTime, _ := time.Parse("2006-01-02", currentHabit.LastClickDate)
		today := time.Now()

		// Получаем индексы дней
		currentDayIndex := int(today.Weekday())
		if currentDayIndex == 0 {
			currentDayIndex = 6
		} else {
			currentDayIndex--
		}

		lastClickDayIndex := int(lastClickTime.Weekday())
		if lastClickDayIndex == 0 {
			lastClickDayIndex = 6
		} else {
			lastClickDayIndex--
		}

		// Находим предыдущий разрешенный день для текущего дня
		currentDayPos := -1
		for i, day := range currentHabit.Days {
			if day == currentDayIndex {
				currentDayPos = i
				break
			}
		}
		prevAllowedDay := currentDayPos

		if currentDayPos > 0 {
			prevAllowedDay = currentHabit.Days[currentDayPos-1]
		} else if len(currentHabit.Days) > 0 {
			prevAllowedDay = currentHabit.Days[len(currentHabit.Days)-1]
		}

		// Если последний клик был в предыдущий разрешенный день, увеличиваем streak
		if lastClickDayIndex == prevAllowedDay {
			newStreak = currentHabit.Streak + 1
		} else {
			newStreak = 1 // Сбрасываем streak
		}
	} else {
		newStreak = 1 // Первое выполнение привычки
	}

	update := bson.M{
		"$set": bson.M{
			"habits.$[habit].score":           currentHabit.Score + 1,
			"habits.$[habit].streak":          newStreak,
			"habits.$[habit].last_click_date": time.Now().Format("2006-01-02"),
		},
	}

	arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"habit.id": habitRequest.Habit.ID},
		},
	})

	result, err := h.usersCollection.UpdateOne(
		context.Background(),
		bson.M{"telegram_id": habitRequest.TelegramID},
		update,
		arrayFilters,
	)

	if err != nil {
		http.Error(w, `{"message": "Ошибка при обновлении в базе данных"}`, http.StatusInternalServerError)
		return
	}

	// Получаем обновленную привычку
	err = h.usersCollection.FindOne(
		context.Background(),
		bson.M{
			"telegram_id": habitRequest.TelegramID,
			"habits.id":   habitRequest.Habit.ID,
		},
	).Decode(&user)

	if err != nil {
		http.Error(w, `{"message": "Ошибка при получении обновленной привычки"}`, http.StatusInternalServerError)
		return
	}

	// Находим обновленную привычку в массиве
	var updatedHabit models.Habit
	for _, h := range user.Habits {
		if h.ID == habitRequest.Habit.ID {
			updatedHabit = h
			break
		}
	}

	log.Println(result)

	// Возвращаем обновленную привычку вместе с результатом операции
	response := struct {
		ModifiedCount int64        `json:"modified_count"`
		Habit         models.Habit `json:"habit"`
	}{
		ModifiedCount: result.ModifiedCount,
		Habit:         updatedHabit,
	}

	json.NewEncoder(w).Encode(response)

	// После успешного обновления привычки обновляем историю
	today := time.Now().Format("2006-01-02")

	// Проверяем, существует ли запись за сегодня
	err = h.usersCollection.FindOne(
		context.Background(),
		bson.M{
			"telegram_id": habitRequest.TelegramID,
			"history": bson.M{
				"$elemMatch": bson.M{
					"date": today,
				},
			},
		},
	).Decode(&user)

	var updateOperation bson.M
	var needsArrayFilter bool

	if err == mongo.ErrNoDocuments {
		// Если записи за сегодня нет, создаем новую
		updateOperation = bson.M{
			"$push": bson.M{
				"history": models.DayHistory{
					Date: today,
					Habits: []models.HabitHistory{
						{
							HabitID: habitRequest.Habit.ID,
							Title:   habitRequest.Habit.Title,
							Done:    true,
						},
					},
				},
			},
		}
		needsArrayFilter = false
	} else {
		// Если запись за сегодня существует, добавляем привычку в существующий массив
		updateOperation = bson.M{
			"$push": bson.M{
				"history.$[elem].habits": models.HabitHistory{
					HabitID: habitRequest.Habit.ID,
					Title:   habitRequest.Habit.Title,
					Done:    true,
				},
			},
		}
		needsArrayFilter = true
	}

	var updateOpts *options.UpdateOptions
	if needsArrayFilter {
		updateOpts = options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: []interface{}{
				bson.M{"elem.date": today},
			},
		})
	}

	var err2 error
	if needsArrayFilter {
		_, err2 = h.usersCollection.UpdateOne(
			context.Background(),
			bson.M{"telegram_id": habitRequest.TelegramID},
			updateOperation,
			updateOpts,
		)
	} else {
		_, err2 = h.usersCollection.UpdateOne(
			context.Background(),
			bson.M{"telegram_id": habitRequest.TelegramID},
			updateOperation,
		)
	}

	if err2 != nil {
		log.Printf("Ошибка при обновлении истории: %v", err2)
		http.Error(w, `{"message": "Ошибка при обновлении истории"}`, http.StatusInternalServerError)
		return
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
