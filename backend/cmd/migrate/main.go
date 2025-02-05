package main

import (
	"backend/migrations"
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Подключаемся к MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Проверяем подключение
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Запускаем миграцию
	if err := migrations.MigrateHabitsToFollowers(client, "ht_db"); err != nil {
		log.Printf("Ошибка при выполнении миграции: %v", err)
		os.Exit(1)
	}

	log.Println("Миграция успешно завершена")
}
