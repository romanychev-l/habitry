package migrations

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func MigrateHabitsToFollowers(client *mongo.Client, dbName string) error {
	ctx := context.Background()

	// Получаем коллекции
	habitsCollection := client.Database(dbName).Collection("habits")
	followersCollection := client.Database(dbName).Collection("followers")

	// Получаем все записи из коллекции followers
	cursor, err := followersCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Ошибка при получении записей из followers: %v", err)
		return err
	}
	defer cursor.Close(ctx)

	processedCount := 0
	errorCount := 0

	for cursor.Next(ctx) {
		var follower struct {
			ID            string `bson:"_id"`
			TelegramID    int64  `bson:"telegram_id"`
			HabitID       string `bson:"habit_id"`
			LastClickDate string `bson:"last_click_date"`
			Streak        int    `bson:"streak"`
			Score         int    `bson:"score"`
			Followers     []struct {
				TelegramID int64  `bson:"telegram_id"`
				HabitID    string `bson:"habit_id"`
			} `bson:"followers"`
		}

		if err := cursor.Decode(&follower); err != nil {
			log.Printf("Ошибка при декодировании записи followers: %v", err)
			errorCount++
			continue
		}
		fmt.Println(follower)

		// Получаем соответствующую привычку из коллекции habits
		var habit struct {
			Title        string    `bson:"title"`
			WantToBecome string    `bson:"want_to_become"`
			Days         []int     `bson:"days"`
			IsOneTime    bool      `bson:"is_one_time"`
			CreatedAt    time.Time `bson:"created_at"`
		}

		habitObjectID, err := primitive.ObjectIDFromHex(follower.HabitID)
		if err != nil {
			log.Printf("Ошибка при преобразовании habit_id: %v", err)
			errorCount++
			continue
		}

		err = habitsCollection.FindOne(ctx, bson.M{"_id": habitObjectID}).Decode(&habit)
		if err != nil {
			log.Printf("Ошибка при получении привычки: %v", err)
			errorCount++
			continue
		}
		fmt.Println(habit)
		// Обновляем запись в followers, добавляя поля из habits и удаляя habit_id
		update := bson.M{
			"$set": bson.M{
				"title":          habit.Title,
				"want_to_become": habit.WantToBecome,
				"days":           habit.Days,
				"is_one_time":    habit.IsOneTime,
				"created_at":     habit.CreatedAt,
			},
		}
		// Преобразуем строковый ID в ObjectID
		followerObjectID, err := primitive.ObjectIDFromHex(follower.ID)
		if err != nil {
			log.Printf("Ошибка при преобразовании follower ID: %v", err)
			errorCount++
			continue
		}

		_, err = followersCollection.UpdateOne(
			ctx,
			bson.M{"_id": followerObjectID},
			update,
		)

		if err != nil {
			log.Printf("Ошибка при обновлении записи followers: %v", err)
			errorCount++
			continue
		}

		processedCount++
		if processedCount%100 == 0 {
			log.Printf("Обработано записей: %d, ошибок: %d", processedCount, errorCount)
		}
	}

	log.Printf("Миграция завершена. Всего обработано: %d, ошибок: %d", processedCount, errorCount)

	// Обработка массива followers
	cursor, err = followersCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Ошибка при получении записей для обновления массива followers: %v", err)
		return err
	}
	defer cursor.Close(ctx)

	processedCount = 0
	errorCount = 0

	for cursor.Next(ctx) {
		var doc struct {
			ID        string `bson:"_id"`
			Followers []struct {
				TelegramID int64  `bson:"telegram_id"`
				HabitID    string `bson:"habit_id"`
			} `bson:"followers"`
		}

		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Ошибка при декодировании документа: %v", err)
			errorCount++
			continue
		}

		var updatedFollowers []string
		for _, follower := range doc.Followers {
			// Ищем соответствующий документ по telegram_id и habit_id
			var matchingDoc struct {
				ID primitive.ObjectID `bson:"_id"`
			}
			err := followersCollection.FindOne(ctx, bson.M{
				"telegram_id": follower.TelegramID,
				"habit_id":    follower.HabitID,
			}).Decode(&matchingDoc)

			if err != nil {
				log.Printf("Не найден документ для telegram_id: %d и habit_id: %s: %v",
					follower.TelegramID, follower.HabitID, err)
				continue
			}
			updatedFollowers = append(updatedFollowers, matchingDoc.ID.Hex())
		}

		// Обновляем массив followers только если он не пустой
		if len(updatedFollowers) > 0 {
			// Обновляем массив followers и удаляем habit_id
			docID, err := primitive.ObjectIDFromHex(doc.ID)
			if err != nil {
				log.Printf("Ошибка при преобразовании ID документа: %v", err)
				errorCount++
				continue
			}

			_, err = followersCollection.UpdateOne(
				ctx,
				bson.M{"_id": docID},
				bson.M{
					"$set": bson.M{"followers": updatedFollowers},
				},
			)

			if err != nil {
				log.Printf("Ошибка при обновлении массива followers: %v", err)
				errorCount++
				continue
			}
		}

		processedCount++
		if processedCount%100 == 0 {
			log.Printf("Обработано записей (обновление followers): %d, ошибок: %d",
				processedCount, errorCount)
		}
	}

	log.Printf("Обновление массива followers завершено. Всего обработано: %d, ошибок: %d",
		processedCount, errorCount)

	// Удаляем поле habit_id из всех документов
	result, err := followersCollection.UpdateMany(
		ctx,
		bson.M{}, // пустой фильтр для обновления всех документов
		bson.M{
			"$unset": bson.M{"habit_id": ""},
		},
	)
	if err != nil {
		log.Printf("Ошибка при удалении поля habit_id: %v", err)
		return err
	}
	log.Printf("Поле habit_id удалено из %d документов", result.ModifiedCount)

	// Удаляем коллекцию habits, так как она больше не нужна
	// if err := habitsCollection.Drop(ctx); err != nil {
	// 	log.Printf("Ошибка при удалении коллекции habits: %v", err)
	// 	return err
	// }
	log.Printf("Коллекция habits успешно удалена")

	return nil
}
