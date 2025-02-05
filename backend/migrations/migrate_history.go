package migrations

import (
	"context"
	"log"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func MigrateHistory(client *mongo.Client, dbName string) error {
	ctx := context.Background()

	// Получаем коллекции
	historyCollection := client.Database(dbName).Collection("history")
	habitsCollection := client.Database(dbName).Collection("habits")

	// Получаем все записи из коллекции history
	cursor, err := historyCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Ошибка при получении записей из history: %v", err)
		return err
	}
	defer cursor.Close(ctx)

	processedCount := 0
	errorCount := 0

	for cursor.Next(ctx) {
		var history struct {
			ID         primitive.ObjectID `bson:"_id"`
			TelegramID int64              `bson:"telegram_id"`
			Date       string             `bson:"date"`
			Habits     []struct {
				HabitID primitive.ObjectID `bson:"habit_id"`
				Title   string             `bson:"title"`
				Done    bool               `bson:"done"`
			} `bson:"habits"`
		}

		if err := cursor.Decode(&history); err != nil {
			log.Printf("Ошибка при декодировании записи history: %v", err)
			errorCount++
			continue
		}

		// Обрабатываем каждую привычку в массиве habits
		for i, habit := range history.Habits {
			// Ищем привычку по telegram_id и title
			var foundHabit struct {
				ID primitive.ObjectID `bson:"_id"`
			}
			err := habitsCollection.FindOne(ctx, bson.M{
				"telegram_id": history.TelegramID,
				"title":       habit.Title,
			}).Decode(&foundHabit)

			if err != nil {
				log.Printf("Не найдена привычка для telegram_id: %d и title: %s: %v",
					history.TelegramID, habit.Title, err)
				errorCount++
				continue
			}

			// Обновляем habit_id в массиве habits
			update := bson.M{
				"$set": bson.M{
					"habits." + strconv.Itoa(i) + ".habit_id": foundHabit.ID,
				},
			}

			_, err = historyCollection.UpdateOne(
				ctx,
				bson.M{"_id": history.ID},
				update,
			)

			if err != nil {
				log.Printf("Ошибка при обновлении записи history: %v", err)
				errorCount++
				continue
			}
		}

		processedCount++
		if processedCount%100 == 0 {
			log.Printf("Обработано записей: %d, ошибок: %d", processedCount, errorCount)
		}
	}

	log.Printf("Миграция истории завершена. Всего обработано: %d, ошибок: %d", processedCount, errorCount)
	return nil
}
