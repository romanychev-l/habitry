package services

import (
	"backend/models"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CalculateHabitCompletionProgress вычисляет прогресс выполнения привычки подписчиками на сегодня.
// Требует доступ к коллекции привычек для получения данных подписчиков.
func CalculateHabitCompletionProgress(ctx context.Context, habit models.Habit, timezone string, habitsCollection *mongo.Collection) (float64, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Ошибка загрузки таймзоны %s: %v", timezone, err)
		return 0.0, err // Возвращаем ошибку, если таймзона невалидна
	}
	today := time.Now().In(loc).Format("2006-01-02")

	completedCount := 0
	// Проверяем владельца
	if habit.LastClickDate == today {
		completedCount++
	}

	totalParticipants := 1 // Начинаем с владельца (всегда 1)

	// Проверяем подписчиков (дедуп по пользователю, учитываем только реально найденные привычки)
	if len(habit.Followers) > 0 {
		// Преобразуем строковые ID в ObjectID c дедупликацией
		idSet := make(map[primitive.ObjectID]struct{})
		for _, followerID := range habit.Followers {
			objectID, err := primitive.ObjectIDFromHex(followerID)
			if err != nil {
				log.Printf("Неверный формат ID подписчика %s: %v", followerID, err)
				continue // Пропускаем невалидные ID
			}
			idSet[objectID] = struct{}{}
		}

		if len(idSet) > 0 {
			followerObjectIDs := make([]primitive.ObjectID, 0, len(idSet))
			for id := range idSet {
				followerObjectIDs = append(followerObjectIDs, id)
			}

			cursor, err := habitsCollection.Find(ctx, bson.M{
				"_id": bson.M{"$in": followerObjectIDs},
			})
			if err != nil {
				log.Printf("Ошибка получения привычек подписчиков для %s: %v", habit.ID.Hex(), err)
				return 0.0, err
			}

			func() {
				defer cursor.Close(ctx)
				// Дедуп по пользователю и агрегирование выполнения «сегодня»
				userSeen := make(map[int64]bool)
				userCompleted := make(map[int64]bool)
				for cursor.Next(ctx) {
					var followerHabit models.Habit
					if err := cursor.Decode(&followerHabit); err != nil {
						log.Printf("Ошибка декодирования привычки подписчика для %s: %v", habit.ID.Hex(), err)
						continue
					}
					userSeen[followerHabit.TelegramID] = true
					if followerHabit.LastClickDate == today {
						userCompleted[followerHabit.TelegramID] = true
					}
				}

				// Всего участников = владелец + уникальные пользователи среди подписок
				totalParticipants = 1 + len(userSeen)
				// Добавляем выполнивших среди уникальных пользователей
				for _, done := range userCompleted {
					if done {
						completedCount++
					}
				}
			}()
		}
	}

	progress := 0.0
	if totalParticipants > 0 {
		progress = float64(completedCount) / float64(totalParticipants)
	}

	return progress, nil
}
