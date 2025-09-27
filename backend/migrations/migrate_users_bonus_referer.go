package migrations

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MigrateUsersBonusAndReferer выполняет следующие действия для всех пользователей:
// 1) Начисляет +100 WILL к текущему балансу
// 2) Приводит баланс по модулю 1000
// 3) Удаляет старое поле referer_id
// 4) Устанавливает referrer_id = 248603604
func MigrateUsersBonusAndReferer(client *mongo.Client, dbName string) error {
	ctx := context.Background()
	usersCollection := client.Database(dbName).Collection("users")

	cursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Ошибка при получении пользователей: %v", err)
		return err
	}
	defer cursor.Close(ctx)

	processedCount := 0
	errorCount := 0

	for cursor.Next(ctx) {
		var user struct {
			ID      primitive.ObjectID `bson:"_id"`
			Balance int64              `bson:"balance"`
		}

		if err := cursor.Decode(&user); err != nil {
			log.Printf("Ошибка декодирования пользователя: %v", err)
			errorCount++
			continue
		}

		// Начисляем +100 и берём по модулю 1000
		newBalance := user.Balance%100 + 100

		update := bson.M{
			// "$unset": bson.M{"referer_id": ""}, // удаляем старое поле
			"$set": bson.M{
				"balance": newBalance,
				// "referrer_id": int64(248603604),
			},
		}

		_, err := usersCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
		if err != nil {
			log.Printf("Ошибка при обновлении пользователя %s: %v", user.ID.Hex(), err)
			errorCount++
			continue
		}

		processedCount++
		if processedCount%100 == 0 {
			log.Printf("Обработано пользователей: %d, ошибок: %d", processedCount, errorCount)
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Ошибка курсора при обходе пользователей: %v", err)
		return err
	}

	log.Printf("Миграция пользователей завершена. Всего обработано: %d, ошибок: %d", processedCount, errorCount)
	return nil
}
