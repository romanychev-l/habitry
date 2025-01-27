package habit

import (
	"backend/models"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	habitsCollection    *mongo.Collection
	historyCollection   *mongo.Collection
	followersCollection *mongo.Collection
}

func NewHandler(habitsCollection, historyCollection, followersCollection *mongo.Collection) *Handler {
	return &Handler{
		habitsCollection:    habitsCollection,
		historyCollection:   historyCollection,
		followersCollection: followersCollection,
	}
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHabit", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var habitRequest models.HabitRequest
	if err := json.NewDecoder(r.Body).Decode(&habitRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем новую привычку
	habit := habitRequest.Habit
	habit.ID = primitive.NewObjectID()
	habit.CreatedAt = time.Now()
	habit.CreatorID = habitRequest.TelegramID

	log.Printf("Создаем привычку для пользователя %d", habitRequest.TelegramID)
	// Сохраняем привычку
	result, err := h.habitsCollection.InsertOne(context.Background(), habit)
	if err != nil {
		log.Printf("Ошибка при сохранении привычки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Привычка создана с ID: %v", result.InsertedID)

	// Создаем запись в коллекции followers
	follower := models.HabitFollowers{
		TelegramID:    habitRequest.TelegramID,
		HabitID:       habit.ID.Hex(),
		LastClickDate: "",
		Streak:        0,
		Score:         0,
		Followers:     []models.Follower{},
	}

	_, err = h.followersCollection.InsertOne(context.Background(), follower)
	if err != nil {
		log.Printf("Ошибка при сохранении followers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем созданную привычку
	var createdHabit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": result.InsertedID}).Decode(&createdHabit)
	if err != nil {
		log.Printf("Ошибка при получении созданной привычки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Получена созданная привычка: %+v", createdHabit)

	// Получаем созданную привычку с статистикой
	habitWithStats := models.HabitWithStats{
		Habit:         createdHabit,
		LastClickDate: "",
		Streak:        0,
		Score:         0,
	}

	// Возращаем созданную привычку
	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habitWithStats,
	})
}

func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHabitUpdate", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var habitRequest models.HabitRequest
	if err := json.NewDecoder(r.Body).Decode(&habitRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем привычку
	var habit models.Habit
	habitID, err := primitive.ObjectIDFromHex(habitRequest.Habit.ID.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Находим запись в followers
	var follower models.HabitFollowers
	err = h.followersCollection.FindOne(
		context.Background(),
		bson.M{
			"telegram_id": habitRequest.TelegramID,
			"habit_id":    habit.ID.Hex(),
		},
	).Decode(&follower)

	if err != nil {
		http.Error(w, "Участник не найден", http.StatusNotFound)
		return
	}

	log.Printf("follower: %+v", follower)
	// Обновляем статистику участника
	today := time.Now().Format("2006-01-02")
	if follower.LastClickDate != today {
		log.Printf("follower.LastClickDate != today")
		// Если это первое выполнение или новый день
		follower.LastClickDate = today
		follower.Streak++
		follower.Score++

		// Обновляем followers
		_, err = h.followersCollection.UpdateOne(
			context.Background(),
			bson.M{
				"telegram_id": habitRequest.TelegramID,
				"habit_id":    habit.ID.Hex(),
			},
			bson.M{"$set": bson.M{
				"last_click_date": follower.LastClickDate,
				"streak":          follower.Streak,
				"score":           follower.Score,
			}},
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Обновляем историю
		// Сначала проверяем существование записи
		var existingHistory models.History
		err = h.historyCollection.FindOne(
			context.Background(),
			bson.M{
				"telegram_id": habitRequest.TelegramID,
				"date":        today,
			},
		).Decode(&existingHistory)

		if err == mongo.ErrNoDocuments {
			// Если записи нет, создаем новую
			history := models.History{
				TelegramID: habitRequest.TelegramID,
				Date:       today,
				Habits: []models.HabitHistory{
					{
						HabitID: habit.ID.Hex(),
						Title:   habit.Title,
						Done:    true,
					},
				},
			}
			_, err = h.historyCollection.InsertOne(context.Background(), history)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			// Если запись существует, проверяем есть ли привычка в массиве
			habitExists := false
			for _, h := range existingHistory.Habits {
				if h.HabitID == habit.ID.Hex() {
					habitExists = true
					break
				}
			}

			update := bson.M{
				"$set": bson.M{
					"habits.$[habit].done": true,
				},
			}
			opts := options.Update().SetArrayFilters(options.ArrayFilters{
				Filters: []interface{}{
					bson.M{"habit.habit_id": habit.ID.Hex()},
				},
			})

			if !habitExists {
				// Если привычки нет в массиве, добавляем её
				update = bson.M{
					"$push": bson.M{
						"habits": models.HabitHistory{
							HabitID: habit.ID.Hex(),
							Title:   habit.Title,
							Done:    true,
						},
					},
				}
				opts = options.Update()
			}

			_, err = h.historyCollection.UpdateOne(
				context.Background(),
				bson.M{
					"telegram_id": habitRequest.TelegramID,
					"date":        today,
				},
				update,
				opts,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Создаем ответ с обновленной статистикой
		habitWithStats := models.HabitWithStats{
			Habit:         habit,
			LastClickDate: follower.LastClickDate,
			Streak:        follower.Streak,
			Score:         follower.Score,
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"habit": habitWithStats,
		})
		return
	}

	// Если привычка уже была выполнена сегодня, возвращаем её без изменений
	habitWithStats := models.HabitWithStats{
		Habit:         habit,
		LastClickDate: follower.LastClickDate,
		Streak:        follower.Streak,
		Score:         follower.Score,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habitWithStats,
	})
}

func (h *Handler) HandleEdit(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHabitEdit", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var habitRequest models.HabitRequest
	if err := json.NewDecoder(r.Body).Decode(&habitRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем привычку
	var habit models.Habit
	habitID, err := primitive.ObjectIDFromHex(habitRequest.Habit.ID.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Проверяем, является ли пользователь создателем привычки
	if habit.CreatorID != habitRequest.TelegramID {
		http.Error(w, "Только создатель может изменять привычку", http.StatusForbidden)
		return
	}

	// Обновляем поля привычки
	if habitRequest.Habit.Title != "" {
		habit.Title = habitRequest.Habit.Title
	}
	if habitRequest.Habit.WantToBecome != "" {
		habit.WantToBecome = habitRequest.Habit.WantToBecome
	}
	if len(habitRequest.Habit.Days) > 0 {
		habit.Days = habitRequest.Habit.Days
	}
	if habitRequest.Habit.IsOneTime {
		habit.IsOneTime = habitRequest.Habit.IsOneTime
	}

	// Сохраняем обновленную привычку
	_, err = h.habitsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": habitID},
		bson.M{"$set": habit},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Находим запись в followers для получения статистики
	var follower models.HabitFollowers
	err = h.followersCollection.FindOne(
		context.Background(),
		bson.M{
			"telegram_id": habitRequest.TelegramID,
			"habit_id":    habit.ID.Hex(),
		},
	).Decode(&follower)

	if err != nil {
		http.Error(w, "Участник не найден", http.StatusNotFound)
		return
	}

	// Возвращаем обновленную привычку со статистикой
	habitWithStats := models.HabitWithStats{
		Habit:         habit,
		LastClickDate: follower.LastClickDate,
		Streak:        follower.Streak,
		Score:         follower.Score,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habitWithStats,
	})
}

func (h *Handler) HandleUndo(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHabitUndo", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var habitRequest models.HabitRequest
	if err := json.NewDecoder(r.Body).Decode(&habitRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем привычку
	var habit models.Habit
	habitID, err := primitive.ObjectIDFromHex(habitRequest.Habit.ID.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Находим запись в followers
	var follower models.HabitFollowers
	err = h.followersCollection.FindOne(
		context.Background(),
		bson.M{
			"telegram_id": habitRequest.TelegramID,
			"habit_id":    habit.ID.Hex(),
		},
	).Decode(&follower)

	if err != nil {
		http.Error(w, "Участник не найден", http.StatusNotFound)
		return
	}

	// Обновляем статистику участника
	today := time.Now().Format("2006-01-02")
	if follower.LastClickDate == today {
		follower.LastClickDate = ""
		follower.Streak--
		follower.Score--

		// Обновляем followers
		_, err = h.followersCollection.UpdateOne(
			context.Background(),
			bson.M{
				"telegram_id": habitRequest.TelegramID,
				"habit_id":    habit.ID.Hex(),
			},
			bson.M{"$set": bson.M{
				"last_click_date": follower.LastClickDate,
				"streak":          follower.Streak,
				"score":           follower.Score,
			}},
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
				bson.M{"habit.habit_id": habit.ID.Hex()},
			},
		}
		opts := options.Update().SetArrayFilters(arrayFilters)

		_, err = h.historyCollection.UpdateOne(
			context.Background(),
			bson.M{
				"telegram_id": habitRequest.TelegramID,
				"date":        today,
			},
			update,
			opts,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Возвращаем привычку с обновленной статистикой
	habitWithStats := models.HabitWithStats{
		Habit:         habit,
		LastClickDate: follower.LastClickDate,
		Streak:        follower.Streak,
		Score:         follower.Score,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habitWithStats,
	})
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHabitDelete", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TelegramID int64  `json:"telegram_id"`
		HabitID    string `json:"habit_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Ошибка при декодировании запроса: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Удаляем привычку: telegram_id=%d, habit_id=%s", request.TelegramID, request.HabitID)

	// Удаляем запись из followers
	result, err := h.followersCollection.DeleteOne(
		context.Background(),
		bson.M{
			"telegram_id": request.TelegramID,
			"habit_id":    request.HabitID,
		},
	)

	if err != nil {
		log.Printf("Ошибка при удалении из followers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Удалено записей: %d", result.DeletedCount)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Пользователь успешно удален из привычки",
	})
}

func (h *Handler) HandleJoin(w http.ResponseWriter, r *http.Request) {
	log.Println("handleHabitJoin", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TelegramID         int64  `json:"telegram_id"`
		HabitID            string `json:"habit_id"`
		SharedByTelegramID string `json:"shared_by_telegram_id"`
		SharedByHabitID    string `json:"shared_by_habit_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Ошибка при декодировании запроса: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sharedByTelegramID, err := strconv.ParseInt(request.SharedByTelegramID, 10, 64)
	if err != nil {
		log.Printf("Ошибка при преобразовании shared_by_telegram_id: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Присоединяемся к привычке: telegram_id=%d, habit_id=%s, shared_by=%d",
		request.TelegramID, request.HabitID, sharedByTelegramID)

	// Проверяем, существует ли запись в followers
	var existingFollower models.HabitFollowers
	err = h.followersCollection.FindOne(
		context.Background(),
		bson.M{
			"telegram_id": request.TelegramID,
			"habit_id":    request.HabitID,
		},
	).Decode(&existingFollower)

	if err == mongo.ErrNoDocuments {
		log.Printf("Создаем новую запись в followers")
		// Если записи нет, создаем новую с массивом followers из одного элемента
		follower := models.HabitFollowers{
			TelegramID:    request.TelegramID,
			HabitID:       request.HabitID,
			LastClickDate: "",
			Streak:        0,
			Score:         0,
			Followers: []models.Follower{{
				TelegramID: sharedByTelegramID,
				HabitID:    request.SharedByHabitID,
			}},
		}

		result, err := h.followersCollection.InsertOne(context.Background(), follower)
		if err != nil {
			log.Printf("Ошибка при создании записи в followers: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Создана новая запись с ID: %v", result.InsertedID)
	} else if err != nil {
		log.Printf("Ошибка при поиске записи в followers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		log.Printf("Найдена существующая запись, обновляем followers")
		// Если запись существует, добавляем в массив followers
		result, err := h.followersCollection.UpdateOne(
			context.Background(),
			bson.M{
				"telegram_id": request.TelegramID,
				"habit_id":    request.HabitID,
			},
			bson.M{
				"$addToSet": bson.M{
					"followers": models.Follower{
						TelegramID: sharedByTelegramID,
						HabitID:    request.SharedByHabitID,
					},
				},
			},
		)
		if err != nil {
			log.Printf("Ошибка при обновлении followers: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Обновлено записей: %d", result.ModifiedCount)
	}

	// Получаем обновленный список привычек
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"telegram_id": request.TelegramID,
				"habit_id": bson.M{
					"$regex": "^[0-9a-fA-F]{24}$",
				},
			},
		},
		{
			"$addFields": bson.M{
				"objectId": bson.M{"$toObjectId": "$habit_id"},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "habits",
				"localField":   "objectId",
				"foreignField": "_id",
				"as":           "habit",
			},
		},
		{
			"$unwind": "$habit",
		},
		{
			"$project": bson.M{
				"habit": bson.M{
					"_id":            "$habit._id",
					"title":          "$habit.title",
					"want_to_become": "$habit.want_to_become",
					"days":           "$habit.days",
					"is_one_time":    "$habit.is_one_time",
					"created_at":     "$habit.created_at",
					"creator_id":     "$habit.creator_id",
				},
				"last_click_date": 1,
				"streak":          1,
				"score":           1,
			},
		},
	}

	cursor, err := h.followersCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Printf("Ошибка при получении привычек: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var habits []models.HabitWithStats
	if err = cursor.All(context.Background(), &habits); err != nil {
		log.Printf("Ошибка при декодировании привычек: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Успешно присоединился к привычке",
		"habits":  habits,
	})
}
