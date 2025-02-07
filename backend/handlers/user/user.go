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

const ObjectIDHexRegex = "^[0-9a-fA-F]{24}$"

// Вспомогательная функция для проверки наличия элемента в массиве
func contains(arr []int, val int) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}

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
		log.Println("POST")
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Printf("Ошибка декодирования JSON: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user.CreatedAt = time.Now()
		user.LastVisit = time.Now().Format("2006-01-02")

		result, err := h.usersCollection.InsertOne(context.Background(), user)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	case "GET":
		telegramID := r.URL.Query().Get("telegram_id")
		id, _ := strconv.ParseInt(telegramID, 10, 64)

		log.Println("GET ", id, telegramID)

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
		cursor, err := h.habitsCollection.Find(
			context.Background(),
			bson.M{"telegram_id": user.TelegramID},
		)
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
		log.Printf("Привычки: %+v", habits)

		// Проверяем кредит только если последний визит был не сегодня
		userTimezone := user.Timezone
		if userTimezone == "" {
			userTimezone = "UTC"
		}
		loc, err := time.LoadLocation(userTimezone)
		if err != nil {
			log.Printf("Ошибка при загрузке таймзоны: %v", err)
			loc = time.UTC
		}

		today := time.Now().In(loc).Format("2006-01-02")
		if user.LastVisit != today {
			// Получаем историю за вчера
			yesterday := time.Now().In(loc).AddDate(0, 0, -1)
			yesterdayStr := yesterday.Format("2006-01-02")

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
						completedHabits[h.HabitID.Hex()] = true
					}
				}
			}

			// Проверяем какие привычки были запланированы на вчера
			scheduledHabits := make(map[string]bool)
			for _, habit := range habits {
				// Пропускаем автопривычки при проверке пропущенных
				if habit.IsAuto {
					continue
				}

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
					if len(habit.Days) == 0 || contains(habit.Days, weekday) {
						scheduledHabits[habit.ID.Hex()] = true
					}
				}
			}

			// Обнуляем streak для пропущенных привычек
			for habitID := range scheduledHabits {
				if !completedHabits[habitID] {
					habitObjectID, err := primitive.ObjectIDFromHex(habitID)
					if err != nil {
						log.Printf("Ошибка при преобразовании habit_id: %v", err)
						continue
					}

					_, err = h.habitsCollection.UpdateOne(
						context.Background(),
						bson.M{
							"_id": habitObjectID,
						},
						bson.M{
							"$set": bson.M{
								"streak": 0,
							},
						},
					)
					if err != nil {
						log.Printf("Ошибка при обнулении streak для привычки %s: %v", habitID, err)
					}
				}
			}

			// Автоматически отмечаем автопривычки
			for _, habit := range habits {
				if habit.IsAuto {
					// Проверяем, должна ли привычка быть выполнена сегодня
					shouldComplete := false
					if habit.IsOneTime {
						habitDate := habit.CreatedAt.Format("2006-01-02")
						shouldComplete = habitDate == today
					} else {
						weekday := int(time.Now().In(loc).Weekday())
						if weekday == 0 {
							weekday = 6
						} else {
							weekday--
						}
						shouldComplete = len(habit.Days) == 0 || contains(habit.Days, weekday)
					}

					if shouldComplete && habit.LastClickDate != today {
						// Обновляем привычку
						_, err = h.habitsCollection.UpdateOne(
							context.Background(),
							bson.M{"_id": habit.ID},
							bson.M{
								"$set": bson.M{
									"last_click_date": today,
									"streak":          habit.Streak + 1,
									"score":           habit.Score + 1,
								},
							},
						)
						if err != nil {
							log.Printf("Ошибка при автоматическом выполнении привычки: %v", err)
							continue
						}

						// Обновляем историю
						history := models.History{
							TelegramID: user.TelegramID,
							Date:       today,
							Habits: []models.HabitHistory{
								{
									HabitID: habit.ID,
									Title:   habit.Title,
									Done:    true,
								},
							},
						}

						// Проверяем существование записи в истории
						var existingHistory models.History
						err = h.historyCollection.FindOne(
							context.Background(),
							bson.M{
								"telegram_id": user.TelegramID,
								"date":        today,
							},
						).Decode(&existingHistory)

						if err == mongo.ErrNoDocuments {
							// Если записи нет, создаем новую
							_, err = h.historyCollection.InsertOne(context.Background(), history)
						} else if err == nil {
							// Если запись существует, добавляем привычку
							_, err = h.historyCollection.UpdateOne(
								context.Background(),
								bson.M{
									"telegram_id": user.TelegramID,
									"date":        today,
								},
								bson.M{
									"$push": bson.M{
										"habits": models.HabitHistory{
											HabitID: habit.ID,
											Title:   habit.Title,
											Done:    true,
										},
									},
								},
							)
						}

						if err != nil {
							log.Printf("Ошибка при обновлении истории для автопривычки: %v", err)
						}
					}
				}
			}

			// Получаем обновленные привычки
			cursor, err = h.habitsCollection.Find(
				context.Background(),
				bson.M{"telegram_id": user.TelegramID},
			)
			if err != nil {
				log.Printf("Ошибка при получении обновленных привычек: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cursor.Close(context.Background())

			habits = nil
			if err = cursor.All(context.Background(), &habits); err != nil {
				log.Printf("Ошибка при декодировании обновленных привычек: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Обновляем последний визит пользователя
			_, err = h.usersCollection.UpdateOne(
				context.Background(),
				bson.M{"telegram_id": id},
				bson.M{
					"$set": bson.M{
						"last_visit": today,
					},
				},
			)

			if err != nil {
				log.Printf("Ошибка при обновлении last_visit: %v", err)
			}

			// Обновляем баланс в локальной копии пользователя для ответа
			user.Balance = user.Balance
		}

		// Фильтруем привычки для текущего дня
		todayHabits := []models.Habit{}
		today = time.Now().In(loc).Format("2006-01-02")

		for _, habit := range habits {
			log.Printf("Проверяем привычку: %+v", habit)
			if habit.IsOneTime {
				habitDate := habit.CreatedAt.Format("2006-01-02")
				log.Printf("Одноразовая привычка: %s vs %s", habitDate, today)
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
				log.Printf("Регулярная привычка, текущий день: %d, дни привычки: %v", weekday, habit.Days)
				// Если массив дней пустой или содержит текущий день
				if len(habit.Days) == 0 || contains(habit.Days, weekday) {
					todayHabits = append(todayHabits, habit)
				}
			}
		}
		log.Printf("Отфильтровано %d привычек для текущего дня", len(todayHabits))

		// Создаем структуру ответа
		type UserResponse struct {
			ID                   primitive.ObjectID `json:"_id,omitempty"`
			TelegramID           int64              `json:"telegram_id"`
			Username             string             `json:"username"`
			FirstName            string             `json:"first_name"`
			LanguageCode         string             `json:"language_code"`
			PhotoURL             string             `json:"photo_url"`
			CreatedAt            time.Time          `json:"created_at"`
			Balance              int                `json:"balance"`
			LastVisit            string             `json:"last_visit"`
			Timezone             string             `json:"timezone"`
			NotificationsEnabled bool               `json:"notifications_enabled"`
			NotificationTime     string             `json:"notification_time"`
			Habits               []models.Habit     `json:"habits"`
		}

		response := UserResponse{
			ID:                   user.ID,
			TelegramID:           user.TelegramID,
			Username:             user.Username,
			FirstName:            user.FirstName,
			LanguageCode:         user.LanguageCode,
			PhotoURL:             user.PhotoURL,
			CreatedAt:            user.CreatedAt,
			Balance:              user.Balance,
			LastVisit:            today,
			Timezone:             user.Timezone,
			NotificationsEnabled: user.NotificationsEnabled,
			NotificationTime:     user.NotificationTime,
			Habits:               todayHabits,
		}

		// Обновляем последний визит
		_, err = h.usersCollection.UpdateOne(
			context.Background(),
			bson.M{"telegram_id": id},
			bson.M{"$set": bson.M{"last_visit": today}},
		)

		if err != nil {
			log.Printf("Ошибка при обновлении last_visit: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
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

	// Обновляем last_visit
	update := bson.M{
		"$set": bson.M{
			"last_visit": today,
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

type UserSettingsRequest struct {
	TelegramID           int64  `json:"telegram_id"`
	NotificationsEnabled bool   `json:"notifications_enabled"`
	NotificationTime     string `json:"notification_time"`
}

// HandleSettings обрабатывает запросы для получения и обновления настроек пользователя
func (h *Handler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	log.Println("HandleSettings called")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		telegramID := r.URL.Query().Get("telegram_id")
		if telegramID == "" {
			http.Error(w, "telegram_id is required", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseInt(telegramID, 10, 64)
		if err != nil {
			http.Error(w, "invalid telegram_id", http.StatusBadRequest)
			return
		}

		var user models.User
		err = h.usersCollection.FindOne(context.Background(), bson.M{"telegram_id": id}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"notifications_enabled": user.NotificationsEnabled,
			"notification_time":     user.NotificationTime,
		})

	case "PUT":
		var req UserSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.TelegramID == 0 {
			http.Error(w, "telegram_id is required", http.StatusBadRequest)
			return
		}

		// Проверяем формат времени, если оно указано
		if req.NotificationTime != "" {
			_, err := time.Parse("15:04", req.NotificationTime)
			if err != nil {
				http.Error(w, "invalid time format", http.StatusBadRequest)
				return
			}
		}

		// Обновляем настройки в базе данных
		filter := bson.M{"telegram_id": req.TelegramID}
		update := bson.M{
			"$set": bson.M{
				"notifications_enabled": req.NotificationsEnabled,
				"notification_time":     req.NotificationTime,
			},
		}

		_, err := h.usersCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			log.Printf("Ошибка при обновлении настроек: %v", err)
			http.Error(w, "failed to update settings", http.StatusInternalServerError)
			return
		}

		log.Printf("Настройки успешно обновлены для пользователя %d", req.TelegramID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":               true,
			"notifications_enabled": req.NotificationsEnabled,
			"notification_time":     req.NotificationTime,
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleUserProfile обрабатывает GET запрос для получения публичного профиля пользователя
func (h *Handler) HandleUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("HandleUserProfile called")
	username := r.URL.Query().Get("username")
	log.Printf("Searching for username: %s", username)

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	filter := bson.M{"username": username}
	log.Printf("MongoDB filter: %+v", filter)

	var user models.User
	err := h.usersCollection.FindOne(r.Context(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("User not found for username: %s", username)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error finding user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Found user: %+v", user)

	// Получаем привычки пользователя
	habitsFilter := bson.M{"telegram_id": user.TelegramID}
	log.Printf("Looking for habits with filter: %+v", habitsFilter)
	cursor, err := h.habitsCollection.Find(r.Context(), habitsFilter)
	if err != nil {
		log.Printf("Error finding habits: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	var habits []models.Habit
	if err = cursor.All(r.Context(), &habits); err != nil {
		log.Printf("Error decoding habits: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Printf("Found %d habits for user", len(habits))

	response := struct {
		TelegramID string         `json:"telegram_id"`
		Username   string         `json:"username"`
		FirstName  string         `json:"first_name"`
		PhotoURL   string         `json:"photo_url"`
		Habits     []models.Habit `json:"habits"`
	}{
		TelegramID: strconv.FormatInt(user.TelegramID, 10),
		Username:   user.Username,
		FirstName:  user.FirstName,
		PhotoURL:   user.PhotoURL,
		Habits:     habits,
	}

	log.Printf("Sending response: %+v", response)
	json.NewEncoder(w).Encode(response)
}
