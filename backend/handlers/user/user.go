package user

import (
	"backend/models"
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"backend/middleware"
	"backend/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func findPrevHabitDay(days []int, currentWeekday int) (daysAgo int) {
	prevWeekdayIndex := 0
	for i := 0; i < len(days); i++ {
		if days[i] == currentWeekday {
			prevWeekdayIndex = i - 1
			break
		}
	}

	if prevWeekdayIndex == -1 {
		prevWeekdayIndex = len(days) - 1
	}

	prevWeekday := days[prevWeekdayIndex]
	daysAgo = currentWeekday - prevWeekday
	if daysAgo < 0 {
		daysAgo += 7
	}
	return daysAgo
}

// HandleUser обрабатывает запросы на создание и обновление пользователя
func (h *Handler) HandleUser(c *gin.Context) {
	// Получаем данные из контекста Telegram
	initData, exists := middleware.CtxInitData(c.Request.Context())
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: no user data in context"})
		return
	}

	// --- Начало: Обработка реферальной ссылки из start_param ---
	startParam := initData.StartParam
	var referrerUsername string

	if strings.HasPrefix(startParam, "ref_") {
		referrerUsername = strings.TrimPrefix(startParam, "ref_")
	} else if strings.HasPrefix(startParam, "profile_") {
		referrerUsername = strings.TrimPrefix(startParam, "profile_")
	}

	if referrerUsername != "" {
		// Пытаемся найти пользователя-реферера
		var referrerUser models.User
		err := h.usersCollection.FindOne(context.Background(), bson.M{"username": referrerUsername}).Decode(&referrerUser)
		if err != nil {
			log.Printf("Referrer user with username '%s' not found: %v", referrerUsername, err)
		} else {
			// Реферер найден, теперь работаем с текущим пользователем
			// Устанавливаем реферера только если текущий пользователь - не тот же самый человек
			if referrerUser.TelegramID != initData.User.ID {
				// Пытаемся найти существующего пользователя, чтобы проверить, есть ли у него уже реферер
				var currentUser models.User
				err := h.usersCollection.FindOne(context.Background(), bson.M{"telegram_id": initData.User.ID}).Decode(&currentUser)

				if err == nil {
					// Пользователь существует, проверяем referrer_id
					if currentUser.ReferrerID == 0 {
						// Реферер еще не установлен, устанавливаем
						_, updateErr := h.usersCollection.UpdateOne(
							context.Background(),
							bson.M{"_id": currentUser.ID},
							bson.M{"$set": bson.M{"referrer_id": referrerUser.TelegramID}},
						)
						if updateErr != nil {
							log.Printf("Failed to set referrer for existing user %d: %v", currentUser.TelegramID, updateErr)
						} else {
							log.Printf("Referrer %d set for existing user %d", referrerUser.TelegramID, currentUser.TelegramID)
						}
					}
				} else if err == mongo.ErrNoDocuments {
					// Пользователь новый. Реферер будет добавлен при создании.
					// Мы не можем здесь просто создать пользователя, так как остальная часть функции HandleUser это делает.
					// Вместо этого, передадим ID реферера дальше.
					// Однако, текущая структура кода создает пользователя ниже. Мы можем просто добавить referrer_id в объект user.
					// Важно: эта логика сработает до блока "if err == mongo.ErrNoDocuments"
				}
			}
		}
	}
	// --- Конец: Обработка реферальной ссылки ---

	// Получаем timezone из контекста
	timezone, exists := middleware.CtxTimezone(c.Request.Context())
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Timezone not provided in context"})
		return
	}

	// Получаем данные из тела запроса
	var req struct {
		PhotoURL *string `json:"photo_url,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Формируем данные пользователя
	user := models.User{
		TelegramID:   initData.User.ID,
		FirstName:    initData.User.FirstName,
		Username:     initData.User.Username,
		LanguageCode: initData.User.LanguageCode,
		Timezone:     timezone,
	}

	// Если передан URL фото, используем его
	if req.PhotoURL != nil {
		user.PhotoURL = *req.PhotoURL
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		log.Printf("ERROR: Failed to load timezone '%s' for user %d: %v", user.Timezone, user.TelegramID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timezone provided"})
		return
	}
	today := time.Now().In(loc).Format("2006-01-02")
	now := time.Now().In(loc)
	todayWeekday := int(now.Weekday())
	if todayWeekday == 0 {
		todayWeekday = 6
	} else {
		todayWeekday--
	}
	originalLastVisitDate := ""

	// Пытаемся найти существующего пользователя
	var existingUser models.User
	err = h.usersCollection.FindOne(
		context.Background(),
		bson.M{"telegram_id": user.TelegramID},
	).Decode(&existingUser)

	if err == mongo.ErrNoDocuments {
		// Создаем нового пользователя
		user.CreatedAt = time.Now()
		user.LastVisit = today
		user.NotificationsEnabled = false
		user.NotificationTime = "09:00"
		user.Balance = 100 // Начисляем 100 WILL за регистрацию

		// Добавляем реферера, если он был определен выше
		if referrerUsername != "" {
			var referrerUser models.User
			err := h.usersCollection.FindOne(context.Background(), bson.M{"username": referrerUsername}).Decode(&referrerUser)
			if err == nil && referrerUser.TelegramID != user.TelegramID {
				user.ReferrerID = referrerUser.TelegramID
				log.Printf("Referrer %d set for new user %d", referrerUser.TelegramID, user.TelegramID)
			}
		}

		_, err = h.usersCollection.InsertOne(context.Background(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		existingUser = user

		// Для нового пользователя привычек нет, возвращаем пустой []HabitResponse
		response := existingUser.ToResponseWithHabits([]models.HabitResponse{}) // Используем обновленный метод
		c.JSON(http.StatusOK, response)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		// Обновляем существующего пользователя
		originalLastVisitDate = existingUser.LastVisit // Сохраняем исходный LAST VISIT
		// lastVisit = existingUser.LastVisit // Закомментируем, так как lastVisit нужен перед циклом привычек
		update := bson.M{
			"$set": bson.M{
				"first_name":    user.FirstName,
				"username":      user.Username,
				"language_code": user.LanguageCode,
				"timezone":      user.Timezone,
				"last_visit":    today, // Обновляем last_visit здесь
			},
		}

		// Добавляем photo_url в обновление только если он передан
		if req.PhotoURL != nil {
			update["$set"].(bson.M)["photo_url"] = *req.PhotoURL
		}

		_, err = h.usersCollection.UpdateOne(
			context.Background(),
			bson.M{"telegram_id": user.TelegramID},
			update,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		// Обновляем existingUser после апдейта, чтобы иметь актуальные данные
		existingUser.FirstName = user.FirstName
		existingUser.Username = user.Username
		existingUser.LanguageCode = user.LanguageCode
		existingUser.Timezone = user.Timezone
		existingUser.LastVisit = today
		if req.PhotoURL != nil {
			existingUser.PhotoURL = *req.PhotoURL
		}
	}

	// Загружаем привычки пользователя
	cursor, err := h.habitsCollection.Find(
		context.Background(),
		bson.M{"telegram_id": user.TelegramID, "archived": bson.M{"$ne": true}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load habits"})
		return
	}
	defer cursor.Close(context.Background())

	var habits []models.Habit
	if err = cursor.All(context.Background(), &habits); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode habits"})
		return
	}

	// Фильтруем привычки для текущего дня, обновляем streak и ГОТОВИМ ОТВЕТ С ПРОГРЕССОМ
	todayHabitResponses := []models.HabitResponse{}
	log.Printf("originalLastVisitDate: %v", originalLastVisitDate)
	log.Printf("today: %v", today)

	// Для каждой привычки
	for _, habit := range habits {
		// Проверяем, запланирована ли привычка на сегодня
		if !contains(habit.Days, todayWeekday) {
			continue
		}

		updatedHabit := habit // Копируем привычку для модификаций

		// Обновляем стрик, если нужно (логика обновления стрейка остается прежней)
		if originalLastVisitDate != today {
			daysAgo := findPrevHabitDay(habit.Days, todayWeekday)
			prevDate := now.AddDate(0, 0, -daysAgo)
			prevDateStr := prevDate.Format("2006-01-02")
			log.Printf("prevDateStr: %v", prevDateStr)

			// Проверяем, была ли привычка выполнена в предыдущий день
			var prevHistory models.History
			err = h.historyCollection.FindOne(
				context.Background(),
				bson.M{
					"telegram_id": user.TelegramID,
					"date":        prevDateStr,
					"habits": bson.M{
						"$elemMatch": bson.M{
							"habit_id": habit.ID,
							"done":     true,
						},
					},
				},
			).Decode(&prevHistory)

			newStreak := habit.Streak
			newScore := habit.Score
			wasDoneYesterday := err != mongo.ErrNoDocuments

			updateFields := bson.M{}

			// Обновляем стрик и скор в зависимости от типа привычки и выполнения
			if habit.IsAuto {
				if !wasDoneYesterday {
					newStreak = 1
					newScore += 1
				} else {
					newStreak += 1
					newScore += 1
				}
				// Для автопривычек всегда обновляем историю
				history := models.History{
					TelegramID: user.TelegramID,
					Date:       today,
					Habits: []models.HabitHistory{{
						HabitID: habit.ID,
						Title:   habit.Title,
						Done:    true,
					}},
				}
				err = h.upsertHistory(user.TelegramID, today, history)
				if err != nil {
					log.Printf("Ошибка при обновлении истории для автопривычки: %v", err)
				}

				updateFields["last_click_date"] = today
				updatedHabit.LastClickDate = today // Обновляем копию
			} else if !wasDoneYesterday {
				newStreak = 0
			}
			// Случай 3: Не автопривычка и была выполнена вчера - ничего не делаем

			// Обновляем привычку в базе
			updateFields["streak"] = newStreak
			updateFields["score"] = newScore

			_, err = h.habitsCollection.UpdateOne(
				context.Background(),
				bson.M{"_id": habit.ID},
				bson.M{"$set": updateFields},
			)
			if err != nil {
				log.Printf("Ошибка при обновлении привычки %s: %v", habit.ID.Hex(), err)
			}

			updatedHabit.Streak = newStreak // Обновляем копию
			updatedHabit.Score = newScore   // Обновляем копию
		}

		// Вызываем функцию из сервиса
		progress, err := services.CalculateHabitCompletionProgress(c.Request.Context(), updatedHabit, timezone, h.habitsCollection)
		if err != nil {
			log.Printf("Ошибка расчета прогресса для привычки %s в HandleUser: %v. Установлен прогресс 0.", updatedHabit.ID.Hex(), err)
			progress = 0.0 // Устанавливаем 0 в случае ошибки
		}

		// Добавляем HabitResponse в результат
		todayHabitResponses = append(todayHabitResponses, models.HabitResponse{
			ID:            updatedHabit.ID,
			TelegramID:    updatedHabit.TelegramID,
			Title:         updatedHabit.Title,
			WantToBecome:  updatedHabit.WantToBecome,
			Days:          updatedHabit.Days,
			IsOneTime:     updatedHabit.IsOneTime,
			IsAuto:        updatedHabit.IsAuto,
			CreatedAt:     updatedHabit.CreatedAt,
			LastClickDate: updatedHabit.LastClickDate,
			Streak:        updatedHabit.Streak,
			Score:         updatedHabit.Score,
			Stake:         updatedHabit.Stake,
			Followers:     []models.FollowerInfo{}, // Подписчиков здесь не обогащаем
			Progress:      progress,
		})
	}

	// Формируем и отправляем ответ
	response := existingUser.ToResponseWithHabits(todayHabitResponses) // Используем обновленный метод
	c.JSON(http.StatusOK, response)
}

// HandleSettings обрабатывает запросы на получение и обновление настроек пользователя
func (h *Handler) HandleSettings(c *gin.Context) {
	// Получаем данные из контекста Telegram
	initData, exists := middleware.CtxInitData(c.Request.Context())
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: no user data in context"})
		return
	}

	switch c.Request.Method {
	case http.MethodGet:
		var user models.User
		err := h.usersCollection.FindOne(context.Background(), bson.M{"telegram_id": initData.User.ID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"notifications_enabled": user.NotificationsEnabled,
			"notification_time":     user.NotificationTime,
		})

	case http.MethodPut:
		var req struct {
			NotificationsEnabled bool   `json:"notifications_enabled"`
			NotificationTime     string `json:"notification_time" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем формат времени
		_, err := time.Parse("15:04", req.NotificationTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time format"})
			return
		}

		// Обновляем настройки
		filter := bson.M{"telegram_id": initData.User.ID}
		update := bson.M{
			"$set": bson.M{
				"notifications_enabled": req.NotificationsEnabled,
				"notification_time":     req.NotificationTime,
			},
		}

		result, err := h.usersCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":               true,
			"notifications_enabled": req.NotificationsEnabled,
			"notification_time":     req.NotificationTime,
		})

	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

// HandleUpdateLastVisit обновляет дату последнего визита пользователя
func (h *Handler) HandleUpdateLastVisit(c *gin.Context) {
	// Получаем данные из контекста Telegram
	initData, exists := middleware.CtxInitData(c.Request.Context())
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: no user data in context"})
		return
	}

	// Получаем timezone из контекста
	timezone, exists := middleware.CtxTimezone(c.Request.Context())
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Timezone not provided in context"})
		return
	}

	// Получаем текущее время в часовом поясе пользователя
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timezone"})
		return
	}
	today := time.Now().In(loc).Format("2006-01-02")

	// Обновляем last_visit
	filter := bson.M{"telegram_id": initData.User.ID}
	update := bson.M{
		"$set": bson.M{
			"last_visit": today,
		},
	}

	result, err := h.usersCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"last_visit": today,
	})
}

// HandleUserProfile обрабатывает запрос на получение публичного профиля пользователя
func (h *Handler) HandleUserProfile(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	var user models.User
	err := h.usersCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Получаем привычки пользователя
	cursor, err := h.habitsCollection.Find(context.Background(), bson.M{"telegram_id": user.TelegramID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var habits []models.Habit
	if err = cursor.All(context.Background(), &habits); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"telegram_id": user.TelegramID,
		"username":    user.Username,
		"first_name":  user.FirstName,
		"photo_url":   user.PhotoURL,
		"habits":      habits,
	})
}

// GetLeaderboard возвращает список лидеров по балансу
func (h *Handler) GetLeaderboard(c *gin.Context) {
	// Устанавливаем параметры для поиска и сортировки
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "balance", Value: -1}}) // Сортировка по убыванию баланса
	findOptions.SetLimit(50)                                 // Ограничиваем до 50 результатов

	// Выполняем поиск
	cursor, err := h.usersCollection.Find(context.Background(), bson.D{}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	defer cursor.Close(context.Background())

	// Декодируем результаты
	var users []models.User
	if err = cursor.All(context.Background(), &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
		return
	}

	// Формируем ответ
	type LeaderboardUser struct {
		Rank      int    `json:"rank"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		PhotoURL  string `json:"photo_url"`
		Balance   int    `json:"balance"`
	}

	leaderboard := make([]LeaderboardUser, 0, len(users))
	for i, user := range users {
		leaderboard = append(leaderboard, LeaderboardUser{
			Rank:      i + 1,
			Username:  user.Username,
			FirstName: user.FirstName,
			PhotoURL:  user.PhotoURL,
			Balance:   user.Balance,
		})
	}

	c.JSON(http.StatusOK, leaderboard)
}

func (h *Handler) upsertHistory(telegramID int64, date string, history models.History) error {
	var existingHistory models.History
	err := h.historyCollection.FindOne(
		context.Background(),
		bson.M{
			"telegram_id": telegramID,
			"date":        date,
		},
	).Decode(&existingHistory)

	if err == mongo.ErrNoDocuments {
		_, err = h.historyCollection.InsertOne(context.Background(), history)
	} else if err == nil {
		_, err = h.historyCollection.UpdateOne(
			context.Background(),
			bson.M{
				"telegram_id": telegramID,
				"date":        date,
			},
			bson.M{
				"$push": bson.M{
					"habits": bson.M{
						"$each": history.Habits,
					},
				},
			},
		)
	}
	return err
}
