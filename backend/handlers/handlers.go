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
	usersCollection   *mongo.Collection
	historyCollection *mongo.Collection
	bot               *bot.Bot
}

func NewHandler(usersCollection, historyCollection *mongo.Collection, b *bot.Bot) *Handler {
	return &Handler{
		usersCollection:   usersCollection,
		historyCollection: historyCollection,
		bot:               b,
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

			var yesterdayHistory models.History
			err = h.historyCollection.FindOne(
				context.Background(),
				bson.M{
					"telegram_id": id,
					"date":        yesterdayStr,
				},
			).Decode(&yesterdayHistory)

			// Проверяем, какие привычки были выполнены
			completedHabits := make(map[string]bool)
			if err == nil {
				for _, habitHistory := range yesterdayHistory.Habits {
					completedHabits[habitHistory.HabitID] = true
				}
			}

			scheduledHabits := make(map[string]bool)
			for _, habit := range user.Habits {
				// Получаем вчерашний день недели
				yesterday := time.Now().AddDate(0, 0, -1)
				weekday := int(yesterday.Weekday())
				if weekday == 0 {
					weekday = 6
				} else {
					weekday--
				}

				// Проверяем, была ли привычка запланирована на вчера
				if habit.IsOneTime {
					// Для одноразовых дел проверяем дату создания
					habitDate := habit.CreatedAt.Format("2006-01-02")
					if habitDate == yesterdayStr {
						scheduledHabits[habit.ID] = true
					}
				} else {
					// Для регулярных привычек проверяем день недели
					for _, day := range habit.Days {
						if day == weekday {
							scheduledHabits[habit.ID] = true
							break
						}
					}
				}
			}

			// Количество невыполненных привычек - это количество оставшихся в мапе
			missedHabits := len(scheduledHabits) - len(completedHabits)

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
		today = time.Now().Format("2006-01-02")

		for _, habit := range user.Habits {
			if habit.IsOneTime {
				// Для одноразовых дел покаываем их в день создания
				habitDate := habit.CreatedAt.Format("2006-01-02")
				if habitDate == today {
					todayHabits = append(todayHabits, habit)
				}
			} else {
				// Существующая логика для обычных привычек

				weekday := int(time.Now().Weekday())
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
	log.Println(habitRequest)
	log.Println("OneTime", habitRequest.Habit.IsOneTime)

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
		http.Error(w, `{"message": "Ошибк�� при обновлении в базе данных"}`, http.StatusInternalServerError)
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

	// Проверяем существование записи за сегодня
	var existingHistory models.History
	err = h.historyCollection.FindOne(
		context.Background(),
		bson.M{
			"telegram_id": habitRequest.TelegramID,
			"date":        today,
		},
	).Decode(&existingHistory)

	if err == mongo.ErrNoDocuments {
		// Создаем новую запись истории
		newHistory := models.History{
			TelegramID: habitRequest.TelegramID,
			Date:       today,
			Habits: []models.HabitHistory{
				{
					HabitID: habitRequest.Habit.ID,
					Title:   habitRequest.Habit.Title,
					Done:    true,
				},
			},
		}

		_, err = h.historyCollection.InsertOne(context.Background(), newHistory)
	} else if err == nil {
		// Добавляем привычку в существующую запись
		_, err = h.historyCollection.UpdateOne(
			context.Background(),
			bson.M{
				"telegram_id": habitRequest.TelegramID,
				"date":        today,
			},
			bson.M{
				"$push": bson.M{
					"habits": models.HabitHistory{
						HabitID: habitRequest.Habit.ID,
						Title:   habitRequest.Habit.Title,
						Done:    true,
					},
				},
			},
		)
	}

	if err != nil {
		log.Printf("Ошибка при обновлении истории: %v", err)
		return
	}
}

func (h *Handler) HandleHabitUndo(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHabitUndo", r.Method)
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

	// Получаем текущую привычку для правильного обновления score
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

	// Обновляем score и streak с проверкой на отрицательные значения
	newStreak := currentHabit.Streak - 1
	if newStreak < 0 {
		newStreak = 0
	}

	newScore := currentHabit.Score - 1
	if newScore < 0 {
		newScore = 0
	}

	update := bson.M{
		"$set": bson.M{
			"habits.$[habit].last_click_date": "",
			"habits.$[habit].streak":          newStreak,
			"habits.$[habit].score":           newScore,
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

	// Находим об��овленную привычку в массиве
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
