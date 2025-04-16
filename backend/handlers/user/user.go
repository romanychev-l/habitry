package user

import (
	"backend/models"
	"context"
	"log"
	"net/http"
	"time"

	"backend/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timezone"})
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
	lastVisit := time.Now().In(loc).Format("2006-01-02")

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

		_, err = h.usersCollection.InsertOne(context.Background(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		existingUser = user
		c.JSON(http.StatusOK, existingUser.ToResponseWithHabits([]models.Habit{}))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		// Обновляем существующего пользователя
		lastVisit = existingUser.LastVisit
		update := bson.M{
			"$set": bson.M{
				"first_name":    user.FirstName,
				"username":      user.Username,
				"language_code": user.LanguageCode,
				"timezone":      user.Timezone,
				"last_visit":    today,
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
	}

	// Загружаем привычки пользователя
	cursor, err := h.habitsCollection.Find(
		context.Background(),
		bson.M{"telegram_id": user.TelegramID},
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

	// Фильтруем привычки для текущего дня и обновляем streak
	todayHabits := []models.Habit{}
	loc, _ = time.LoadLocation(user.Timezone)
	if loc == nil {
		loc = time.UTC
	}

	// Для каждой привычки
	for _, habit := range habits {
		// Проверяем, запланирована ли привычка на сегодня
		if !contains(habit.Days, todayWeekday) {
			continue
		}

		if lastVisit == today {
			todayHabits = append(todayHabits, habit)
			continue
		}

		// Находим предыдущий день, когда привычка должна была быть выполнена
		daysAgo := findPrevHabitDay(habit.Days, todayWeekday)
		prevDate := now.AddDate(0, 0, -daysAgo)
		prevDateStr := prevDate.Format("2006-01-02")

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

		updatedHabit := habit
		newStreak := habit.Streak
		newScore := habit.Score
		wasDoneYesterday := err != mongo.ErrNoDocuments

		updateFields := bson.M{}

		// Обновляем стрик и скор в зависимости от типа привычки и выполнения
		if habit.IsAuto {
			if !wasDoneYesterday {
				// Случай 1: Автопривычка не была выполнена вчера
				newStreak = 1
				newScore += 1
			} else {
				// Случай 2: Автопривычка была выполнена вчера
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
			updatedHabit.LastClickDate = today
		} else if !wasDoneYesterday {
			// Случай 4: Не автопривычка и не была выполнена вчера
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

		updatedHabit.Streak = newStreak
		updatedHabit.Score = newScore
		todayHabits = append(todayHabits, updatedHabit)
	}

	c.JSON(http.StatusOK, existingUser.ToResponseWithHabits(todayHabits))
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
