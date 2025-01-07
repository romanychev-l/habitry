package migrations

import (
	"backend/models"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OldUser struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	TelegramID int64              `bson:"telegram_id"`
	Habits     []struct {
		ID            string    `bson:"id"`
		Title         string    `bson:"title"`
		WantToBecome  string    `bson:"want_to_become"`
		Days          []int     `bson:"days"`
		IsOneTime     bool      `bson:"is_one_time"`
		LastClickDate string    `bson:"last_click_date"`
		Streak        int       `bson:"streak"`
		Score         int       `bson:"score"`
		CreatedAt     time.Time `bson:"created_at"`
	} `bson:"habits"`
}

func MigrateHabits(client *mongo.Client, dbName string) error {
	ctx := context.Background()

	// Получаем коллекции
	usersCollection := client.Database(dbName).Collection("users")
	habitsCollection := client.Database(dbName).Collection("habits")

	// Получаем всех пользователей со старой структурой
	cursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var oldUser OldUser
		if err := cursor.Decode(&oldUser); err != nil {
			log.Printf("Ошибка декодирования пользователя: %v", err)
			continue
		}

		var habitIDs []primitive.ObjectID

		// Для каждой привычки пользователя
		for _, oldHabit := range oldUser.Habits {
			// Создаем новый ID для привычки
			habitID := primitive.NewObjectID()

			// Создаем новую привычку
			habit := models.Habit{
				ID:           habitID,
				Title:        oldHabit.Title,
				WantToBecome: oldHabit.WantToBecome,
				Days:         oldHabit.Days,
				IsOneTime:    oldHabit.IsOneTime,
				CreatedAt:    oldHabit.CreatedAt,
				CreatorID:    oldUser.TelegramID,
				IsShared:     false,
				Participants: []models.HabitParticipant{
					{
						TelegramID:    oldUser.TelegramID,
						LastClickDate: oldHabit.LastClickDate,
						Streak:        oldHabit.Streak,
						Score:         oldHabit.Score,
					},
				},
			}

			// Сохраняем привычку
			result, err := habitsCollection.InsertOne(ctx, habit)
			if err != nil {
				log.Printf("Ошибка сохранения привычки: %v", err)
				continue
			}

			habitIDs = append(habitIDs, result.InsertedID.(primitive.ObjectID))
		}

		// Обновляем пользователя
		_, err = usersCollection.UpdateOne(
			ctx,
			bson.M{"_id": oldUser.ID},
			bson.M{
				"$set": bson.M{
					"habit_ids": habitIDs,
				},
				"$unset": bson.M{
					"habits": "",
				},
			},
		)
		if err != nil {
			log.Printf("Ошибка обновления пользователя: %v", err)
		}
	}

	return nil
}
