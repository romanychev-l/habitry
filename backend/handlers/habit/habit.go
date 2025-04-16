package habit

import (
	"backend/models"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"backend/middleware"

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

	// Инициализируем массив с начальной емкостью равной количеству подписчиков
	followers := make([]models.FollowerInfo, 0, len(habit.Followers))

	// Получаем информацию о подписчиках
	for _, followerID := range habit.Followers {
		followerObjectID, err := primitive.ObjectIDFromHex(followerID)
		if err != nil {
			continue
		}

		var followerHabit models.Habit
		err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": followerObjectID}).Decode(&followerHabit)
		if err != nil {
			continue
		}

		// Получаем информацию о пользователе
		var user models.User
		err = h.usersCollection.FindOne(context.Background(), bson.M{"telegram_id": followerHabit.TelegramID}).Decode(&user)
		if err != nil {
			log.Printf("Ошибка при получении информации о пользователе: %v", err)
			continue
		}

		// Проверяем взаимную подписку
		isMutual := false
		// Получаем все привычки подписчика
		cursor, err := h.habitsCollection.Find(context.Background(), bson.M{
			"telegram_id": followerHabit.TelegramID,
		})
		if err == nil {
			defer cursor.Close(context.Background())
			var followerHabits []models.Habit
			if err = cursor.All(context.Background(), &followerHabits); err == nil {
				// Проверяем, есть ли среди привычек подписчика те, которые подписаны на текущую привычку
				for _, fh := range followerHabits {
					for _, fFollowerID := range fh.Followers {
						if fFollowerID == habitID {
							isMutual = true
							break
						}
					}
					if isMutual {
						break
					}
				}
			}
		}

		// Добавляем полную информацию о подписчике
		followers = append(followers, models.FollowerInfo{
			ID:            followerHabit.ID,
			TelegramID:    followerHabit.TelegramID,
			Title:         followerHabit.Title,
			LastClickDate: followerHabit.LastClickDate,
			Streak:        followerHabit.Streak,
			Score:         followerHabit.Score,
			Username:      user.Username,
			FirstName:     user.FirstName,
			PhotoURL:      user.PhotoURL,
			IsMutual:      isMutual,
		})
	}

	// Всегда возвращаем массив, даже если он пустой
	c.JSON(http.StatusOK, followers)
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
			"$set": bson.M{
				"habits.$[habit].done": false,
			},
		}
		arrayFilters := options.ArrayFilters{
			Filters: []interface{}{
				bson.M{"habit.habit_id": habit.ID},
			},
		}
		opts := options.Update().SetArrayFilters(arrayFilters)

		_, err = h.historyCollection.UpdateOne(
			context.Background(),
			bson.M{
				"telegram_id": initData.User.ID,
				"date":        today,
			},
			update,
			opts,
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
		Followers:     []models.FollowerInfo{},
	}

	// Если есть подписчики, получаем их данные
	if len(habit.Followers) > 0 {
		// Преобразуем строковые ID в ObjectID
		var followerObjectIDs []primitive.ObjectID
		for _, followerID := range habit.Followers {
			objectID, err := primitive.ObjectIDFromHex(followerID)
			if err != nil {
				log.Printf("Ошибка при преобразовании ID подписчика: %v", err)
				continue
			}
			followerObjectIDs = append(followerObjectIDs, objectID)
		}

		// Получаем информацию о всех подписчиках одним запросом
		cursor, err := h.habitsCollection.Find(ctx, bson.M{
			"_id": bson.M{"$in": followerObjectIDs},
		})
		if err != nil {
			return response, err
		}
		defer cursor.Close(ctx)

		// Собираем информацию о подписчиках
		for cursor.Next(ctx) {
			var followerHabit models.Habit
			if err := cursor.Decode(&followerHabit); err != nil {
				log.Printf("Ошибка при декодировании привычки подписчика: %v", err)
				continue
			}

			response.Followers = append(response.Followers, models.FollowerInfo{
				ID:            followerHabit.ID,
				TelegramID:    followerHabit.TelegramID,
				Title:         followerHabit.Title,
				LastClickDate: followerHabit.LastClickDate,
				Streak:        followerHabit.Streak,
				Score:         followerHabit.Score,
			})
		}
	}

	return response, nil
}
