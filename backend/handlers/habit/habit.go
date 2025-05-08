package habit

import (
	"backend/models"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"backend/middleware"
	// "backend/services" // <--- ЗАКОММЕНТИРОВАТЬ ИЛИ УДАЛИТЬ
	"backend/services" // <--- ИСПОЛЬЗУЕМ ПРАВИЛЬНЫЙ ПУТЬ МОДУЛЯ

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
	habitsCollection  *mongo.Collection
	historyCollection *mongo.Collection
	usersCollection   *mongo.Collection
}

func NewHandler(habitsCollection, historyCollection, usersCollection *mongo.Collection) *Handler {
	return &Handler{
		habitsCollection:  habitsCollection,
		historyCollection: historyCollection,
		usersCollection:   usersCollection,
	}
}

func (h *Handler) HandleCreate(c *gin.Context) {
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

	var habit models.Habit
	if err := c.ShouldBindJSON(&habit); err != nil {
		log.Printf("Ошибка при декодировании JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат JSON"})
		return
	}

	log.Printf("Получен запрос на создание привычки: %+v", habit)

	// Проверяем обязательные поля
	if habit.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title обязателен"})
		return
	}

	// Создаем новую привычку
	habit.ID = primitive.NewObjectID()
	habit.TelegramID = initData.User.ID
	habit.CreatedAt = time.Now().In(loc)
	habit.LastClickDate = ""
	habit.Streak = 0
	habit.Score = 0
	habit.Followers = []string{} // пустой массив подписчиков

	// Сначала сохраняем привычку
	result, err := h.habitsCollection.InsertOne(context.Background(), habit)
	if err != nil {
		log.Printf("Ошибка при сохранении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении привычки"})
		return
	}
	log.Printf("Привычка создана с ID: %v", result.InsertedID)

	// Если это автопривычка и она должна быть выполнена сегодня, отмечаем её
	if habit.IsAuto {
		today := time.Now().In(loc).Format("2006-01-02")
		shouldComplete := false

		weekday := int(time.Now().In(loc).Weekday())
		if weekday == 0 {
			weekday = 6
		} else {
			weekday--
		}
		shouldComplete = contains(habit.Days, weekday)

		if shouldComplete {
			// Обновляем привычку
			_, err = h.habitsCollection.UpdateOne(
				context.Background(),
				bson.M{"_id": habit.ID},
				bson.M{
					"$set": bson.M{
						"last_click_date": today,
						"streak":          1,
						"score":           1,
					},
				},
			)
			if err != nil {
				log.Printf("Ошибка при обновлении автопривычки: %v", err)
			}

			// Создаем запись в истории
			history := models.History{
				TelegramID: habit.TelegramID,
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
					"telegram_id": habit.TelegramID,
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
						"telegram_id": habit.TelegramID,
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

	// Получаем обновленную привычку
	var createdHabit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habit.ID}).Decode(&createdHabit)
	if err != nil {
		log.Printf("Ошибка при получении созданной привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении созданной привычки"})
		return
	}

	// Обогащаем привычку информацией о подписчиках
	enrichedHabit, err := h.enrichHabitWithFollowers(context.Background(), createdHabit)
	if err != nil {
		log.Printf("Ошибка при обогащении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обогащении привычки"})
		return
	}

	c.JSON(http.StatusOK, enrichedHabit)
}

func (h *Handler) HandleUpdate(c *gin.Context) {
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

	var req struct {
		ID string `json:"_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Ошибка при декодировании JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат JSON"})
		return
	}

	// Преобразуем ID привычки
	habitID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		log.Printf("Ошибка при преобразовании ID привычки: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID привычки"})
		return
	}

	// Получаем привычку
	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Привычка не найдена"})
			return
		}
		log.Printf("Ошибка при получении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении привычки"})
		return
	}

	// Проверяем, что пользователь является владельцем привычки
	if habit.TelegramID != initData.User.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к привычке"})
		return
	}

	// Получаем текущее время в часовом поясе пользователя
	today := time.Now().In(loc).Format("2006-01-02")

	// Проверяем, не была ли привычка уже выполнена сегодня
	if habit.LastClickDate == today {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Привычка уже выполнена сегодня"})
		return
	}

	// Обновляем привычку
	update := bson.M{
		"$set": bson.M{
			"last_click_date": today,
			"streak":          habit.Streak + 1,
			"score":           habit.Score + 1,
		},
	}

	_, err = h.habitsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": habitID},
		update,
	)
	if err != nil {
		log.Printf("Ошибка при обновлении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении привычки"})
		return
	}

	// Обновляем историю
	history := models.History{
		TelegramID: initData.User.ID,
		Date:       today,
		Habits: []models.HabitHistory{
			{
				HabitID: habitID,
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
			"telegram_id": initData.User.ID,
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
				"telegram_id": initData.User.ID,
				"date":        today,
			},
			bson.M{
				"$push": bson.M{
					"habits": models.HabitHistory{
						HabitID: habitID,
						Title:   habit.Title,
						Done:    true,
					},
				},
			},
		)
	}

	if err != nil {
		log.Printf("Ошибка при обновлении истории: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении истории"})
		return
	}

	// Получаем обновленную привычку
	var updatedHabit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&updatedHabit)
	if err != nil {
		log.Printf("Ошибка при получении обновленной привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении обновленной привычки"})
		return
	}

	// Обогащаем привычку информацией о подписчиках
	enrichedHabit, err := h.enrichHabitWithFollowers(context.Background(), updatedHabit)
	if err != nil {
		log.Printf("Ошибка при обогащении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обогащении привычки"})
		return
	}

	c.JSON(http.StatusOK, enrichedHabit)
}

func (h *Handler) HandleEdit(c *gin.Context) {
	// Получаем данные из контекста Telegram
	initData, exists := middleware.CtxInitData(c.Request.Context())
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: no user data in context"})
		return
	}

	var habit models.Habit
	if err := c.ShouldBindJSON(&habit); err != nil {
		log.Printf("Ошибка при декодировании JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат JSON"})
		return
	}

	log.Printf("Edit habit: %v", habit)

	// Проверяем, что пользователь является владельцем привычки
	var existingHabit models.Habit
	err := h.habitsCollection.FindOne(
		context.Background(),
		bson.M{"_id": habit.ID},
	).Decode(&existingHabit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Привычка не найдена"})
			return
		}
		log.Printf("Ошибка при получении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении привычки"})
		return
	}

	if existingHabit.TelegramID != initData.User.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к привычке"})
		return
	}

	// Обновляем привычку
	update := bson.M{
		"$set": bson.M{
			"title":          habit.Title,
			"want_to_become": habit.WantToBecome,
			"days":           habit.Days,
			"is_one_time":    habit.IsOneTime,
			"is_auto":        habit.IsAuto,
			"stake":          habit.Stake,
		},
	}

	_, err = h.habitsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": habit.ID},
		update,
	)
	if err != nil {
		log.Printf("Ошибка при обновлении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении привычки"})
		return
	}

	// Получаем обновленную привычку
	var updatedHabit models.Habit
	err = h.habitsCollection.FindOne(
		context.Background(),
		bson.M{"_id": habit.ID},
	).Decode(&updatedHabit)
	if err != nil {
		log.Printf("Ошибка при получении обновленной привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении обновленной привычки"})
		return
	}

	// Обогащаем привычку информацией о подписчиках
	enrichedHabit, err := h.enrichHabitWithFollowers(context.Background(), updatedHabit)
	if err != nil {
		log.Printf("Ошибка при обогащении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обогащении привычки"})
		return
	}

	c.JSON(http.StatusOK, enrichedHabit)
}

func (h *Handler) HandleDelete(c *gin.Context) {
	var req struct {
		TelegramID int64  `json:"telegram_id" binding:"required"`
		HabitID    string `json:"habit_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Ошибка при декодировании JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат JSON"})
		return
	}

	log.Printf("Удаляем привычку: telegram_id=%d, habit_id=%s", req.TelegramID, req.HabitID)

	habitObjectID, err := primitive.ObjectIDFromHex(req.HabitID)
	if err != nil {
		log.Printf("Ошибка при преобразовании habit_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID привычки"})
		return
	}

	// Проверяем, что пользователь является владельцем привычки
	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitObjectID}).Decode(&habit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Привычка не найдена"})
			return
		}
		log.Printf("Ошибка при получении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении привычки"})
		return
	}

	if habit.TelegramID != req.TelegramID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к привычке"})
		return
	}

	// Удаляем привычку
	result, err := h.habitsCollection.DeleteOne(
		context.Background(),
		bson.M{
			"_id": habitObjectID,
		},
	)

	if err != nil {
		log.Printf("Ошибка при удалении привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении привычки"})
		return
	}

	log.Printf("Удалено записей: %d", result.DeletedCount)

	// Удаляем ID удаленной привычки из массива followers других привычек
	_, err = h.habitsCollection.UpdateMany(
		context.Background(),
		bson.M{},
		bson.M{
			"$pull": bson.M{
				"followers": req.HabitID,
			},
		},
	)

	if err != nil {
		log.Printf("Ошибка при удалении ID из followers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении ID из followers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Привычка успешно удалена",
	})
}

func (h *Handler) HandleJoin(c *gin.Context) {
	var request struct {
		TelegramID         int64  `json:"telegram_id" binding:"required"`
		HabitID            string `json:"habit_id" binding:"required"`
		SharedByTelegramID string `json:"shared_by_telegram_id" binding:"required"`
		SharedByHabitID    string `json:"shared_by_habit_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Ошибка при декодировании запроса: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат JSON"})
		return
	}

	sharedByTelegramID, err := strconv.ParseInt(request.SharedByTelegramID, 10, 64)
	if err != nil {
		log.Printf("Ошибка при преобразовании shared_by_telegram_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат shared_by_telegram_id"})
		return
	}

	log.Printf("Присоединяемся к привычке: telegram_id=%d, habit_id=%s, shared_by=%d",
		request.TelegramID, request.HabitID, sharedByTelegramID)

	// Если HabitID равен SharedByHabitID, создаем новую привычку
	if request.HabitID == request.SharedByHabitID {
		// Получаем оригинальную привычку
		originalHabitID, err := primitive.ObjectIDFromHex(request.SharedByHabitID)
		if err != nil {
			log.Printf("Ошибка при преобразовании habit_id: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат habit_id"})
			return
		}

		var originalHabit models.Habit
		err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": originalHabitID}).Decode(&originalHabit)
		if err != nil {
			log.Printf("Ошибка при получении оригинальной привычки: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении привычки"})
			return
		}

		// Создаем новую привычку
		newHabit := models.Habit{
			ID:            primitive.NewObjectID(),
			TelegramID:    request.TelegramID,
			Title:         originalHabit.Title,
			WantToBecome:  originalHabit.WantToBecome,
			Days:          originalHabit.Days,
			IsOneTime:     originalHabit.IsOneTime,
			CreatedAt:     time.Now(),
			LastClickDate: "",
			Streak:        0,
			Score:         0,
			Followers:     []string{request.SharedByHabitID},
		}

		// Сохраняем новую привычку
		_, err = h.habitsCollection.InsertOne(context.Background(), newHabit)
		if err != nil {
			log.Printf("Ошибка при создании новой привычки: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании привычки"})
			return
		}

		// Обновляем request.HabitID на ID новой привычки
		request.HabitID = newHabit.ID.Hex()
		log.Printf("Создана новая привычка с ID: %s", request.HabitID)
	} else {
		// Если присоединяемся к существующей привычке
		habitID, err := primitive.ObjectIDFromHex(request.HabitID)
		if err != nil {
			log.Printf("Ошибка при преобразовании habit_id: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат habit_id"})
			return
		}

		// Добавляем ID привычки пользователя в followers
		_, err = h.habitsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": habitID},
			bson.M{
				"$addToSet": bson.M{
					"followers": request.SharedByHabitID,
				},
			},
		)
		if err != nil {
			log.Printf("Ошибка при обновлении followers: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении привычки"})
			return
		}
	}

	// Получаем обновленный список привычек
	cursor, err := h.habitsCollection.Find(
		context.Background(),
		bson.M{"telegram_id": request.TelegramID},
	)
	if err != nil {
		log.Printf("Ошибка при получении привычек: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении привычек"})
		return
	}
	defer cursor.Close(context.Background())

	var habits []models.Habit
	if err = cursor.All(context.Background(), &habits); err != nil {
		log.Printf("Ошибка при декодировании привычек: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обработке привычек"})
		return
	}

	// Преобразуем привычки в HabitResponse
	var habitResponses []models.HabitResponse
	for _, habit := range habits {
		habitResponse, err := h.enrichHabitWithFollowers(c.Request.Context(), habit)
		if err != nil {
			log.Printf("Ошибка при обогащении данных привычки: %v", err)
			continue
		}
		habitResponses = append(habitResponses, habitResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Успешно присоединился к привычке",
		"habits":  habitResponses,
	})
}

func (h *Handler) HandleGetFollowers(c *gin.Context) {
	habitID := c.Query("habit_id")
	if habitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "habit_id is required"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(habitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit_id format"})
		return
	}

	// Получаем привычку
	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&habit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Если привычка не найдена, возвращаем пустой массив
			c.JSON(http.StatusOK, []models.FollowerInfo{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Получаем timezone из контекста
	timezone, exists := middleware.CtxTimezone(c.Request.Context())
	if !exists {
		// Если таймзона не передана, пытаемся получить ее из данных пользователя
		var requestingUser models.User
		err := h.usersCollection.FindOne(context.Background(), bson.M{"telegram_id": habit.TelegramID}).Decode(&requestingUser)
		if err == nil && requestingUser.Timezone != "" {
			timezone = requestingUser.Timezone
		} else {
			timezone = "UTC" // Фоллбэк на UTC
			log.Printf("Таймзона не найдена для пользователя %d в запросе GetFollowers, используется UTC", habit.TelegramID)
		}
	}

	// Получаем текущую дату в нужном часовом поясе
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Ошибка загрузки таймзоны %s: %v. Используется UTC.", timezone, err)
		loc = time.UTC
	}
	today := time.Now().In(loc).Format("2006-01-02")

	// Получаем информацию о подписчиках
	followerInfosMap := make(map[string]models.FollowerInfo) // Ключ - ID привычки другого пользователя (строка)

	// 1. Пользователи, на которых подписан currentUserActualHabit
	for _, followedUserHabitIDStr := range habit.Followers {
		followedUserHabitObjectID, err := primitive.ObjectIDFromHex(followedUserHabitIDStr)
		if err != nil {
			log.Printf("GetFollowers: Ошибка преобразования followedUserHabitIDStr '%s': %v", followedUserHabitIDStr, err)
			continue
		}

		var followedUserHabit models.Habit
		err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": followedUserHabitObjectID}).Decode(&followedUserHabit)
		if err != nil {
			log.Printf("GetFollowers: Ошибка получения followedUserHabit с ID '%s': %v", followedUserHabitIDStr, err)
			continue // Если привычка удалена или не найдена, пропускаем
		}

		var followedUser models.User
		err = h.usersCollection.FindOne(context.Background(), bson.M{"telegram_id": followedUserHabit.TelegramID}).Decode(&followedUser)
		if err != nil {
			log.Printf("GetFollowers: Ошибка получения followedUser с TelegramID '%d': %v", followedUserHabit.TelegramID, err)
			// Можно пропустить или использовать значения по умолчанию для информации о пользователе
			continue
		}

		isFollowingBack := false
		for _, idInFollowedHabitFollowers := range followedUserHabit.Followers {
			if idInFollowedHabitFollowers == habit.ID.Hex() { // habit.ID это ID текущей обрабатываемой привычки (currentUserActualHabit)
				isFollowingBack = true
				break
			}
		}
		completedToday := followedUserHabit.LastClickDate == today

		info := models.FollowerInfo{
			ID:                         followedUserHabit.ID, // ID привычки пользователя, на которого подписаны
			TelegramID:                 followedUserHabit.TelegramID,
			Title:                      followedUserHabit.Title,
			LastClickDate:              followedUserHabit.LastClickDate,
			Streak:                     followedUserHabit.Streak,
			Score:                      followedUserHabit.Score,
			Username:                   followedUser.Username,
			FirstName:                  followedUser.FirstName,
			PhotoURL:                   followedUser.PhotoURL,
			CompletedToday:             completedToday,
			CurrentUserFollowsThisUser: true, // Текущий пользователь подписан на этого пользователя (по определению этого цикла)
			ThisUserFollowsCurrentUser: isFollowingBack,
		}
		followerInfosMap[followedUserHabitIDStr] = info
	}

	// 2. Пользователи, которые подписаны на currentUserActualHabit (обновляем или добавляем в карту)
	followersCursor, err := h.habitsCollection.Find(context.Background(), bson.M{"followers": habit.ID.Hex()})
	if err != nil {
		log.Printf("GetFollowers: Ошибка поиска привычек, которые подписаны на '%s': %v", habit.ID.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error finding followers"})
		return
	}
	defer followersCursor.Close(context.Background())

	for followersCursor.Next(context.Background()) {
		var habitThatFollowsCurrentUser models.Habit
		if err := followersCursor.Decode(&habitThatFollowsCurrentUser); err != nil {
			log.Printf("GetFollowers: Ошибка декодирования habitThatFollowsCurrentUser: %v", err)
			continue
		}

		otherUserHabitIDStr := habitThatFollowsCurrentUser.ID.Hex()

		// Пропускаем, если это собственная привычка пользователя (на всякий случай, хотя логика "followers" должна это исключать)
		if habitThatFollowsCurrentUser.ID == habit.ID {
			continue
		}

		var otherUser models.User
		err = h.usersCollection.FindOne(context.Background(), bson.M{"telegram_id": habitThatFollowsCurrentUser.TelegramID}).Decode(&otherUser)
		if err != nil {
			log.Printf("GetFollowers: Ошибка получения otherUser с TelegramID '%d' для привычки '%s': %v", habitThatFollowsCurrentUser.TelegramID, otherUserHabitIDStr, err)
			continue
		}

		completedToday := habitThatFollowsCurrentUser.LastClickDate == today

		if existingInfo, ok := followerInfosMap[otherUserHabitIDStr]; ok {
			// Этот пользователь уже в карте (т.к. currentUserActualHabit на него подписан).
			// Просто обновляем флаг, что он также подписан на currentUserActualHabit.
			existingInfo.ThisUserFollowsCurrentUser = true
			// Убедимся, что CompletedToday обновлено, если вдруг привычка та же, но другой путь её получения
			if existingInfo.ID == habitThatFollowsCurrentUser.ID { // Должно быть всегда так, если ключ ID привычки
				existingInfo.CompletedToday = completedToday
			}
			followerInfosMap[otherUserHabitIDStr] = existingInfo
		} else {
			// Этот пользователь подписан на currentUserActualHabit, но currentUserActualHabit (пока) не подписан на него.
			currentUserFollowsThisOtherUser := false // По определению этого блока `else`
			// Можно дополнительно проверить, есть ли otherUserHabitIDStr в habit.Followers, но должно быть false

			info := models.FollowerInfo{
				ID:                         habitThatFollowsCurrentUser.ID,
				TelegramID:                 habitThatFollowsCurrentUser.TelegramID,
				Title:                      habitThatFollowsCurrentUser.Title,
				LastClickDate:              habitThatFollowsCurrentUser.LastClickDate,
				Streak:                     habitThatFollowsCurrentUser.Streak,
				Score:                      habitThatFollowsCurrentUser.Score,
				Username:                   otherUser.Username,
				FirstName:                  otherUser.FirstName,
				PhotoURL:                   otherUser.PhotoURL,
				CompletedToday:             completedToday,
				CurrentUserFollowsThisUser: currentUserFollowsThisOtherUser,
				ThisUserFollowsCurrentUser: true, // По определению этого цикла (habitThatFollowsCurrentUser имеет habit.ID в своих followers)
			}
			followerInfosMap[otherUserHabitIDStr] = info
		}
	}

	// Преобразуем карту в слайс
	resultFollowers := make([]models.FollowerInfo, 0, len(followerInfosMap))
	for _, info := range followerInfosMap {
		resultFollowers = append(resultFollowers, info)
	}

	c.JSON(http.StatusOK, resultFollowers)
}

func (h *Handler) HandleGetActivity(c *gin.Context) {
	habitID := c.Query("habit_id")
	if habitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "habit_id is required"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(habitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit_id format"})
		return
	}

	// Получаем привычку
	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&habit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Получаем историю активности
	cursor, err := h.historyCollection.Find(
		context.Background(),
		bson.M{
			"telegram_id":     habit.TelegramID,
			"habits.habit_id": objectID,
		},
		options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(30),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer cursor.Close(context.Background())

	var activity []map[string]interface{}
	for cursor.Next(context.Background()) {
		var history models.History
		if err := cursor.Decode(&history); err != nil {
			continue
		}

		// Находим информацию о привычке в истории
		for _, h := range history.Habits {
			if h.HabitID == objectID {
				activity = append(activity, map[string]interface{}{
					"date": history.Date,
					"done": h.Done,
				})
				break
			}
		}
	}

	c.JSON(http.StatusOK, activity)
}

func (h *Handler) HandleUndo(c *gin.Context) {
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

	var req struct {
		ID string `json:"_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	habitID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit_id format"})
		return
	}

	// Получаем привычку
	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Обновляем статистику привычки
	today := time.Now().In(loc).Format("2006-01-02")
	if habit.LastClickDate == today {
		habit.LastClickDate = ""
		if habit.Streak > 0 {
			habit.Streak--
			habit.Score--
		}

		// Обновляем привычку
		_, err = h.habitsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": habitID},
			bson.M{"$set": bson.M{
				"last_click_date": habit.LastClickDate,
				"streak":          habit.Streak,
				"score":           habit.Score,
			}},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update habit"})
			return
		}

		// Обновляем историю
		update := bson.M{
			"$pull": bson.M{
				"habits": bson.M{"habit_id": habit.ID},
			},
		}

		_, err = h.historyCollection.UpdateOne(
			context.Background(),
			bson.M{
				"telegram_id": initData.User.ID,
				"date":        today,
			},
			update,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update history"})
			return
		}

		// Получаем обновленную версию привычки
		err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get updated habit"})
			return
		}
	}

	// Обогащаем привычку информацией о подписчиках
	enrichedHabit, err := h.enrichHabitWithFollowers(context.Background(), habit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enrich habit"})
		return
	}

	c.JSON(http.StatusOK, enrichedHabit)
}

// Вспомогательная функция для обогащения данных привычки информацией о подписчиках
func (h *Handler) enrichHabitWithFollowers(ctx context.Context, habit models.Habit) (models.HabitResponse, error) {
	// Получаем timezone
	timezone := "UTC" // Значение по умолчанию
	userTimezone, userExists := middleware.CtxTimezone(ctx)
	if userExists {
		timezone = userTimezone
	} else {
		var user models.User
		err := h.usersCollection.FindOne(ctx, bson.M{"telegram_id": habit.TelegramID}).Decode(&user)
		if err == nil && user.Timezone != "" {
			timezone = user.Timezone
		} else {
			log.Printf("Не удалось получить таймзону для пользователя %d в enrichHabitWithFollowers, используется UTC", habit.TelegramID)
		}
	}

	// Рассчитываем прогресс выполнения подписчиками через сервис
	progress, err := services.CalculateHabitCompletionProgress(ctx, habit, timezone, h.habitsCollection)
	if err != nil {
		log.Printf("Ошибка расчета прогресса для привычки %s: %v. Установлен прогресс 0.", habit.ID.Hex(), err)
		progress = 0.0
	}

	response := models.HabitResponse{
		ID:            habit.ID,
		TelegramID:    habit.TelegramID,
		Title:         habit.Title,
		WantToBecome:  habit.WantToBecome,
		Days:          habit.Days,
		IsOneTime:     habit.IsOneTime,
		IsAuto:        habit.IsAuto,
		CreatedAt:     habit.CreatedAt,
		LastClickDate: habit.LastClickDate,
		Streak:        habit.Streak,
		Score:         habit.Score,
		Stake:         habit.Stake,
		Followers:     []models.FollowerInfo{}, // Это поле теперь будет заполняться отдельным запросом getHabitFollowers на фронте
		Progress:      progress,
	}

	// Старая логика заполнения Followers здесь больше не нужна,
	// так как HandleGetFollowers теперь отвечает за полный список связанных пользователей.
	// Поле HabitResponse.Followers можно либо оставить пустым, либо удалить из HabitResponse,
	// если фронтенд всегда будет получать их через /api/habit/followers.
	// Пока оставлю пустым, чтобы не ломать структуру ответа HabitResponse кардинально.

	return response, nil
}

// HandleSubscribeToFollower обрабатывает подписку текущего пользователя на привычку другого пользователя
func (h *Handler) HandleSubscribeToFollower(c *gin.Context) {
	initData, exists := middleware.CtxInitData(c.Request.Context())
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: no user data in context"})
		return
	}
	currentUserTelegramID := initData.User.ID

	var req struct {
		CurrentUserHabitID string `json:"current_user_habit_id" binding:"required"`
		TargetUserHabitID  string `json:"target_user_habit_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HandleSubscribeToFollower: Ошибка при декодировании JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат JSON"})
		return
	}

	currentUserHabitObjectID, err := primitive.ObjectIDFromHex(req.CurrentUserHabitID)
	if err != nil {
		log.Printf("HandleSubscribeToFollower: Ошибка преобразования CurrentUserHabitID '%s': %v", req.CurrentUserHabitID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID текущей привычки"})
		return
	}

	targetUserHabitObjectID, err := primitive.ObjectIDFromHex(req.TargetUserHabitID)
	if err != nil {
		log.Printf("HandleSubscribeToFollower: Ошибка преобразования TargetUserHabitID '%s': %v", req.TargetUserHabitID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID целевой привычки"})
		return
	}

	// Проверяем, что CurrentUserHabitID принадлежит текущему пользователю
	var currentUserHabit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{
		"_id":         currentUserHabitObjectID,
		"telegram_id": currentUserTelegramID,
	}).Decode(&currentUserHabit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Привычка текущего пользователя не найдена или не принадлежит ему"})
			return
		}
		log.Printf("HandleSubscribeToFollower: Ошибка получения currentUserHabit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при проверке привычки"})
		return
	}

	// Проверяем, существует ли TargetUserHabitID (целевая привычка)
	var targetUserHabit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": targetUserHabitObjectID}).Decode(&targetUserHabit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Целевая привычка для подписки не найдена"})
			return
		}
		log.Printf("HandleSubscribeToFollower: Ошибка получения targetUserHabit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при проверке целевой привычки"})
		return
	}

	// Добавляем TargetUserHabitID (строку) в массив followers текущей привычки пользователя
	// Используем $addToSet для предотвращения дубликатов
	_, err = h.habitsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": currentUserHabitObjectID},
		bson.M{"$addToSet": bson.M{"followers": req.TargetUserHabitID}},
	)
	if err != nil {
		log.Printf("HandleSubscribeToFollower: Ошибка обновления привычки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при подписке на привычку"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Успешно подписан на привычку"})
}
