package follower

import (
	"backend/models"
	"context"
	"net/http"
	"time"

	"backend/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	habitsCollection *mongo.Collection
	usersCollection  *mongo.Collection
}

func NewHandler(habitsCollection, usersCollection *mongo.Collection) *Handler {
	return &Handler{
		habitsCollection: habitsCollection,
		usersCollection:  usersCollection,
	}
}

func (h *Handler) HandleHabitProgress(c *gin.Context) {
	habitID := c.Query("habit_id")
	if habitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "habit_id is required"})
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

	// Получаем текущую дату с учетом таймзоны
	today := time.Now().In(loc).Format("2006-01-02")

	// Преобразуем habitID в ObjectID
	habitObjectID, err := primitive.ObjectIDFromHex(habitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit_id format"})
		return
	}

	// Находим привычку
	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{
		"_id": habitObjectID,
	}).Decode(&habit)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusOK, gin.H{
				"total_followers": 0,
				"completed_today": 0,
				"progress":        0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Проверяем, выполнил ли владелец привычку
	completedToday := 0
	if habit.LastClickDate == today {
		completedToday++
	}

	// Подсчитываем количество подписчиков, выполнивших привычку сегодня
	totalFollowers := len(habit.Followers) + 1 // +1 для владельца
	if len(habit.Followers) > 0 {
		// Преобразуем строковые ID в ObjectID
		var followerObjectIDs []primitive.ObjectID
		for _, followerID := range habit.Followers {
			objectID, err := primitive.ObjectIDFromHex(followerID)
			if err != nil {
				continue
			}
			followerObjectIDs = append(followerObjectIDs, objectID)
		}

		// Получаем привычки подписчиков
		cursor, err := h.habitsCollection.Find(context.Background(), bson.M{
			"_id": bson.M{"$in": followerObjectIDs},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		defer cursor.Close(context.Background())

		// Подсчитываем выполненные привычки
		for cursor.Next(context.Background()) {
			var followerHabit models.Habit
			if err := cursor.Decode(&followerHabit); err != nil {
				continue
			}
			if followerHabit.LastClickDate == today {
				completedToday++
			}
		}
	}

	progress := 0.0
	if totalFollowers > 0 {
		progress = float64(completedToday) / float64(totalFollowers)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_followers": totalFollowers,
		"completed_today": completedToday,
		"progress":        progress,
	})
}

func (h *Handler) HandleUnfollow(c *gin.Context) {
	var request struct {
		HabitID    string `json:"habit_id"`
		UnfollowID int64  `json:"unfollow_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Преобразуем ID привычки в ObjectID
	habitObjectID, err := primitive.ObjectIDFromHex(request.HabitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit_id"})
		return
	}

	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{
		"_id": habitObjectID,
	}).Decode(&habit)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	// Получаем все привычки из массива followers
	if len(habit.Followers) > 0 {
		var followerObjectIDs []primitive.ObjectID
		for _, followerID := range habit.Followers {
			objectID, err := primitive.ObjectIDFromHex(followerID)
			if err != nil {
				continue
			}
			followerObjectIDs = append(followerObjectIDs, objectID)
		}

		// Получаем привычки подписчиков
		cursor, err := h.habitsCollection.Find(context.Background(), bson.M{
			"_id": bson.M{"$in": followerObjectIDs},
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Создаем массив ID привычек, которые нужно удалить из followers
		// (если пользователь подписался на одну и туже привычку дважды - они все удалятся)
		var habitsToRemove []string
		for cursor.Next(context.Background()) {
			var followerHabit models.Habit
			if err := cursor.Decode(&followerHabit); err != nil {
				continue
			}

			// Если привычка принадлежит пользователю, который отписывается
			if followerHabit.TelegramID == request.UnfollowID {
				habitsToRemove = append(habitsToRemove, followerHabit.ID.Hex())
			}
		}

		// Удаляем найденные привычки из массива followers
		if len(habitsToRemove) > 0 {
			_, err = h.habitsCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id": habitObjectID,
				},
				bson.M{
					"$pull": bson.M{
						"followers": bson.M{
							"$in": habitsToRemove,
						},
					},
				},
			)

			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
		}
	}

	// Получаем обновленную привычку
	var updatedHabit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{
		"_id": habitObjectID,
	}).Decode(&updatedHabit)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, updatedHabit)
}
