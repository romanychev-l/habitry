package migrations

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func MigrateCreditToBalance(client *mongo.Client, dbName string) error {
	ctx := context.Background()
	usersCollection := client.Database(dbName).Collection("users")

	// Обновляем все документы: удаляем поле credit и устанавливаем balance = 1000
	result, err := usersCollection.UpdateMany(
		ctx,
		bson.M{}, // пустой фильтр для обновления всех документов
		bson.M{
			"$unset": bson.M{"credit": ""},    // удаляем поле credit
			"$set":   bson.M{"balance": 1000}, // устанавливаем balance = 1000
		},
	)

	if err != nil {
		log.Printf("Ошибка при обновлении документов: %v", err)
		return err
	}

	log.Printf("Обновлено документов: %d", result.ModifiedCount)
	return nil
}
