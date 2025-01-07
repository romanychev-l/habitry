package user

import (
	"backend/models"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	usersCollection   *mongo.Collection
	historyCollection *mongo.Collection
	habitsCollection  *mongo.Collection
}

func NewHandler(usersCollection, historyCollection, habitsCollection *mongo.Collection) *Handler {
	return &Handler{
		usersCollection:   usersCollection,
		historyCollection: historyCollection,
		habitsCollection:  habitsCollection,
	}
}

func (h *Handler) HandleUser(w http.ResponseWriter, r *http.Request) {
	log.Println("handleUser", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
		user.Habits = []models.Habit{}

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

		// Загружаем привычки пользователя
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"participants.telegram_id": user.TelegramID,
					"$or": []bson.M{
						{"is_archived": false},
						{"$and": []bson.M{
							{"is_archived": true},
							{"creator_id": bson.M{"$ne": user.TelegramID}},
						}},
					},
				},
			},
		}

		log.Printf("Ищем привычки для пользователя %d", user.TelegramID)
		cursor, err := h.habitsCollection.Aggregate(context.Background(), pipeline)
		if err != nil {
			log.Printf("Ошибка при получении привычек: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		var habits []models.Habit
		if err = cursor.All(context.Background(), &habits); err != nil {
			log.Printf("Ошибка при декодировании привычек: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Найдено %d привычек", len(habits))

		// Проверяем кредит только если последний визит был не сегодня
		userTimezone := user.Timezone
		if userTimezone == "" {
			userTimezone = "UTC"
		}
		loc, err := time.LoadLocation(userTimezone)
		if err != nil {
			loc = time.UTC
		}

		today := time.Now().In(loc).Format("2006-01-02")
		if user.LastVisit != today {
			yesterdayStr := time.Now().In(loc).AddDate(0, 0, -1).Format("2006-01-02")

			// Получаем историю за вчера
			var yesterdayHistory models.History
			err := h.historyCollection.FindOne(
				context.Background(),
				bson.M{
					"telegram_id": id,
					"date":        yesterdayStr,
				},
			).Decode(&yesterdayHistory)

			completedHabits := make(map[string]bool)
			if err == nil {
				for _, h := range yesterdayHistory.Habits {
					if h.Done {
						completedHabits[h.HabitID] = true
					}
				}
			}

			// Проверяем какие привычки были запланированы на вчера
			scheduledHabits := make(map[string]bool)
			for _, habit := range habits {
				yesterday := time.Now().In(loc).AddDate(0, 0, -1)
				weekday := int(yesterday.Weekday())
				if weekday == 0 {
					weekday = 6
				} else {
					weekday--
				}

				if habit.IsOneTime {
					habitDate := habit.CreatedAt.Format("2006-01-02")
					if habitDate == yesterdayStr {
						scheduledHabits[habit.ID.Hex()] = true
					}
				} else {
					for _, day := range habit.Days {
						if day == weekday {
							scheduledHabits[habit.ID.Hex()] = true
							break
						}
					}
				}
			}

			missedHabits := len(scheduledHabits) - len(completedHabits)

			update := bson.M{
				"$set": bson.M{
					"credit":     missedHabits,
					"last_visit": today,
				},
			}

			_, err = h.usersCollection.UpdateOne(
				context.Background(),
				bson.M{"telegram_id": id},
				update,
			)

			if err != nil {
				log.Printf("Ошибка при обновлении кредита: %v", err)
			}

			user.Credit = missedHabits
		}

		// Фильтруем привычки для текущего дня
		todayHabits := []models.Habit{}
		today = time.Now().In(loc).Format("2006-01-02")

		for _, habit := range habits {
			if habit.IsOneTime {
				habitDate := habit.CreatedAt.Format("2006-01-02")
				if habitDate == today {
					todayHabits = append(todayHabits, habit)
				}
			} else {
				weekday := int(time.Now().In(loc).Weekday())
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
		log.Printf("Отфильтровано %d привычек для текущего дня", len(todayHabits))

		// Создаем структуру ответа
		type UserResponse struct {
			ID         primitive.ObjectID `json:"_id"`
			TelegramID int64              `json:"telegram_id"`
			Username   string             `json:"username"`
			FirstName  string             `json:"first_name"`
			Language   string             `json:"language_code"`
			PhotoURL   string             `json:"photo_url"`
			CreatedAt  time.Time          `json:"created_at"`
			Credit     int                `json:"credit"`
			LastVisit  string             `json:"last_visit"`
			Timezone   string             `json:"timezone"`
			Habits     []models.Habit     `json:"habits"`
		}

		response := UserResponse{
			ID:         user.ID,
			TelegramID: user.TelegramID,
			Username:   user.Username,
			FirstName:  user.FirstName,
			Language:   user.Language,
			PhotoURL:   user.PhotoURL,
			CreatedAt:  user.CreatedAt,
			Credit:     user.Credit,
			LastVisit:  user.LastVisit,
			Timezone:   user.Timezone,
			Habits:     todayHabits,
		}

		json.NewEncoder(w).Encode(response)
	}
}

func (h *Handler) HandleUpdateLastVisit(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleUpdateLastVisit called")
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

	var request struct {
		TelegramID int64  `json:"telegram_id"`
		Timezone   string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, `{"message": "Неверный формат данных"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Updating last_visit for user %d with timezone %s", request.TelegramID, request.Timezone)

	// Получаем текущее время в часовом поясе пользователя
	loc, err := time.LoadLocation(request.Timezone)
	if err != nil {
		log.Printf("Error loading timezone: %v, using UTC", err)
		loc = time.UTC
	}
	today := time.Now().In(loc).Format("2006-01-02")
	log.Printf("Setting last_visit to: %s", today)

	// Обновляем last_visit и сбрасываем credit
	update := bson.M{
		"$set": bson.M{
			"last_visit": today,
			"credit":     0,
		},
	}

	result, err := h.usersCollection.UpdateOne(
		context.Background(),
		bson.M{"telegram_id": request.TelegramID},
		update,
	)

	if err != nil {
		log.Printf("Error updating last_visit: %v", err)
		http.Error(w, `{"message": "Ошибка при обновлении даты последнего визита"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("Update result: %+v", result)
	w.WriteHeader(http.StatusOK)
}
