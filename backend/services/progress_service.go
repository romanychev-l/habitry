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

	totalParticipants := 1 // Начинаем с владельца

	// Проверяем подписчиков
	if len(habit.Followers) > 0 {
		totalParticipants += len(habit.Followers)

		// Преобразуем строковые ID в ObjectID
		var followerObjectIDs []primitive.ObjectID
		for _, followerID := range habit.Followers {
			objectID, err := primitive.ObjectIDFromHex(followerID)
			if err != nil {
				log.Printf("Неверный формат ID подписчика %s: %v", followerID, err)
				continue // Пропускаем невалидные ID
			}
			followerObjectIDs = append(followerObjectIDs, objectID)
		}

		// Получаем привычки подписчиков одним запросом, если есть валидные ID
		if len(followerObjectIDs) > 0 {
			cursor, err := habitsCollection.Find(ctx, bson.M{
				"_id": bson.M{"$in": followerObjectIDs},
			})
			if err != nil {
				log.Printf("Ошибка получения привычек подписчиков для %s: %v", habit.ID.Hex(), err)
				// Возвращаем 0.0 и ошибку, т.к. не можем рассчитать точно
				return 0.0, err
			}
			// Важно закрывать курсор сразу после использования
			// Используем func() для defer внутри блока if
			func() {
				defer cursor.Close(ctx)
				for cursor.Next(ctx) {
					var followerHabit models.Habit
					if err := cursor.Decode(&followerHabit); err != nil {
						log.Printf("Ошибка декодирования привычки подписчика для %s: %v", habit.ID.Hex(), err)
						continue
					}
					if followerHabit.LastClickDate == today {
						completedCount++
					}
				}
			}() // Немедленный вызов func
		}
	}

	progress := 0.0
	if totalParticipants > 0 {
		progress = float64(completedCount) / float64(totalParticipants)
	}

	return progress, nil
}
