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
	habitsCollection  *mongo.Collection
	historyCollection *mongo.Collection
	usersCollection   *mongo.Collection
}

func NewHandler(habitsCollection, historyCollection, usersCollection *mongo.Collection) *Handler {
	return &Handler{
		habitsCollection:  habitsCollection,
		historyCollection: historyCollection,
		usersCollection:   usersCollection,
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
	habit.TelegramID = habitRequest.TelegramID
	habit.LastClickDate = ""
	habit.Streak = 0
	habit.Score = 0
	habit.Followers = []string{} // пустой массив подписчиков

	log.Printf("Создаем привычку для пользователя %d", habitRequest.TelegramID)
	// Сохраняем привычку
	result, err := h.habitsCollection.InsertOne(context.Background(), habit)
	if err != nil {
		log.Printf("Ошибка при сохранении привычки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Привычка создана с ID: %v", result.InsertedID)

	// Получаем созданную привычку
	var createdHabit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": result.InsertedID}).Decode(&createdHabit)
	if err != nil {
		log.Printf("Ошибка при получении созданной привычки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Получена созданная привычка: %+v", createdHabit)

	// Преобразуем в HabitResponse
	habitResponse, err := h.enrichHabitWithFollowers(r.Context(), createdHabit)
	if err != nil {
		log.Printf("Ошибка при обогащении данных привычки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем созданную привычку
	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habitResponse,
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

	// Обновляем статистику привычки
	today := time.Now().Format("2006-01-02")
	if habit.LastClickDate != today {
		log.Printf("habit.LastClickDate != today")
		// Если это первое выполнение или новый день
		habit.LastClickDate = today
		habit.Streak++
		habit.Score++

		// Обновляем привычку
		_, err = h.habitsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": habitID},
			bson.M{"$set": bson.M{
				"last_click_date": habit.LastClickDate,
				"streak":          habit.Streak,
				"score":           habit.Score,
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
						HabitID: habit.ID,
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
				if h.HabitID == habit.ID {
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
					bson.M{"habit.habit_id": habit.ID},
				},
			})

			if !habitExists {
				// Если привычки нет в массиве, добавляем её
				update = bson.M{
					"$push": bson.M{
						"habits": models.HabitHistory{
							HabitID: habit.ID,
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

		// Преобразуем в HabitResponse
		habitResponse, err := h.enrichHabitWithFollowers(r.Context(), habit)
		if err != nil {
			log.Printf("Ошибка при обогащении данных привычки: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"habit": habitResponse,
		})
		return
	}

	// Если привычка уже была выполнена сегодня, возвращаем её без изменений
	habitResponse, err := h.enrichHabitWithFollowers(r.Context(), habit)
	if err != nil {
		log.Printf("Ошибка при обогащении данных привычки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habitResponse,
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
	if habit.TelegramID != habitRequest.TelegramID {
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

	// Преобразуем в HabitResponse
	habitResponse, err := h.enrichHabitWithFollowers(r.Context(), habit)
	if err != nil {
		log.Printf("Ошибка при обогащении данных привычки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habitResponse,
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

	// Обновляем статистику привычки
	today := time.Now().Format("2006-01-02")
	if habit.LastClickDate == today {
		habit.LastClickDate = ""
		habit.Streak--
		habit.Score--

		// Обновляем привычку
		_, err = h.habitsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": habitID},
			bson.M{"$set": bson.M{
				"last_click_date": habit.LastClickDate,
				"streak":          habit.Streak,
				"score":           habit.Score,
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
				bson.M{"habit.habit_id": habit.ID},
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

	// Преобразуем в HabitResponse
	habitResponse, err := h.enrichHabitWithFollowers(r.Context(), habit)
	if err != nil {
		log.Printf("Ошибка при обогащении данных привычки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habitResponse,
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

	habitObjectID, err := primitive.ObjectIDFromHex(request.HabitID)
	if err != nil {
		log.Printf("Ошибка при преобразовании habit_id: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Удаляем привычку
	result, err := h.habitsCollection.DeleteOne(
		context.Background(),
		bson.M{
			"_id":         habitObjectID,
		},
	)

	if err != nil {
		log.Printf("Ошибка при удалении привычки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Удалено записей: %d", result.DeletedCount)

	// Удаляем ID удаленной привычки из массива followers других привычек
	_, err = h.habitsCollection.UpdateMany(
		context.Background(),
		bson.M{},
		bson.M{
			"$pull": bson.M{
				"followers": request.HabitID,
			},
		},
	)

	if err != nil {
		log.Printf("Ошибка при удалении ID из followers: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Привычка успешно удалена",
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

	// Если HabitID равен SharedByHabitID, создаем новую привычку
	if request.HabitID == request.SharedByHabitID {
		// Получаем оригинальную привычку
		originalHabitID, err := primitive.ObjectIDFromHex(request.SharedByHabitID)
		if err != nil {
			log.Printf("Ошибка при преобразовании habit_id: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var originalHabit models.Habit
		err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": originalHabitID}).Decode(&originalHabit)
		if err != nil {
			log.Printf("Ошибка при получении оригинальной привычки: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Создаем новую привычку
		newHabit := models.Habit{
			ID:            primitive.NewObjectID(),
			TelegramID:    request.TelegramID,
			Title:         originalHabit.Title,
			WantToBecome:  originalHabit.WantToBecome,
			Days:          originalHabit.Days,
			IsOneTime:     originalHabit.IsOneTime,
			CreatedAt:     time.Now(),
			LastClickDate: "",
			Streak:        0,
			Score:         0,
			Followers:     []string{request.SharedByHabitID},
		}

		// Сохраняем новую привычку
		_, err = h.habitsCollection.InsertOne(context.Background(), newHabit)
		if err != nil {
			log.Printf("Ошибка при создании новой привычки: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Обновляем request.HabitID на ID новой привычки
		request.HabitID = newHabit.ID.Hex()
		log.Printf("Создана новая привычка с ID: %s", request.HabitID)
	} else {
		// Если присоединяемся к существующей привычке
		habitID, err := primitive.ObjectIDFromHex(request.HabitID)
		if err != nil {
			log.Printf("Ошибка при преобразовании habit_id: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Добавляем ID привычки пользователя в followers
		_, err = h.habitsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": habitID},
			bson.M{
				"$addToSet": bson.M{
					"followers": request.SharedByHabitID,
				},
			},
		)
		if err != nil {
			log.Printf("Ошибка при обновлении followers: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Получаем обновленный список привычек
	cursor, err := h.habitsCollection.Find(
		context.Background(),
		bson.M{"telegram_id": request.TelegramID},
	)
	if err != nil {
		log.Printf("Ошибка при получении привычек: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var habits []models.Habit
	if err = cursor.All(context.Background(), &habits); err != nil {
		log.Printf("Ошибка при декодировании привычек: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Преобразуем привычки в HabitResponse
	var habitResponses []models.HabitResponse
	for _, habit := range habits {
		habitResponse, err := h.enrichHabitWithFollowers(r.Context(), habit)
		if err != nil {
			log.Printf("Ошибка при обогащении данных привычки: %v", err)
			continue
		}
		habitResponses = append(habitResponses, habitResponse)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Успешно присоединился к привычке",
		"habits":  habitResponses,
	})
}

func (h *Handler) HandleGetFollowers(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleGetFollowers", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID привычки и telegram_id из параметров запроса
	habitID := r.URL.Query().Get("habit_id")
	telegramIDStr := r.URL.Query().Get("telegram_id")

	if habitID == "" || telegramIDStr == "" {
		http.Error(w, "ID привычки и telegram_id должны быть указаны", http.StatusBadRequest)
		return
	}

	// Проверяем формат telegram_id
	_, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный формат telegram_id", http.StatusBadRequest)
		return
	}

	// Получаем привычку
	habitObjectID, err := primitive.ObjectIDFromHex(habitID)
	if err != nil {
		http.Error(w, "Некорректный формат habit_id", http.StatusBadRequest)
		return
	}

	var habit models.Habit
	err = h.habitsCollection.FindOne(
		context.Background(),
		bson.M{"_id": habitObjectID},
	).Decode(&habit)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем информацию о подписчиках
	var users []map[string]interface{}
	if len(habit.Followers) > 0 {
		// Получаем все привычки подписчиков
		followerHabitIDs := make([]primitive.ObjectID, 0)
		for _, followerID := range habit.Followers {
			objectID, err := primitive.ObjectIDFromHex(followerID)
			if err != nil {
				log.Printf("Ошибка при преобразовании ID подписчика: %v", err)
				continue
			}
			followerHabitIDs = append(followerHabitIDs, objectID)
		}

		cursor, err := h.habitsCollection.Find(
			context.Background(),
			bson.M{"_id": bson.M{"$in": followerHabitIDs}},
		)
		if err != nil {
			log.Printf("Ошибка при получении привычек подписчиков: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Для каждой привычки получаем информацию о пользователе
		for cursor.Next(context.Background()) {
			var followerHabit models.Habit
			if err := cursor.Decode(&followerHabit); err != nil {
				log.Printf("Ошибка при декодировании привычки подписчика: %v", err)
				continue
			}

			var user models.User
			err := h.usersCollection.FindOne(
				context.Background(),
				bson.M{"telegram_id": followerHabit.TelegramID},
			).Decode(&user)

			if err == nil {
				users = append(users, map[string]interface{}{
					"username":    user.Username,
					"telegram_id": user.TelegramID,
				})
			}
		}
	}

	json.NewEncoder(w).Encode(users)
}

// Вспомогательная функция для обогащения данных привычки информацией о подписчиках
func (h *Handler) enrichHabitWithFollowers(ctx context.Context, habit models.Habit) (models.HabitResponse, error) {
	response := models.HabitResponse{
		ID:            habit.ID,
		TelegramID:    habit.TelegramID,
		Title:         habit.Title,
		WantToBecome:  habit.WantToBecome,
		Days:          habit.Days,
		IsOneTime:     habit.IsOneTime,
		CreatedAt:     habit.CreatedAt,
		LastClickDate: habit.LastClickDate,
		Streak:        habit.Streak,
		Score:         habit.Score,
		Followers:     []models.FollowerInfo{},
	}

	// Если есть подписчики, получаем их данные
	if len(habit.Followers) > 0 {
		// Преобразуем строковые ID в ObjectID
		var followerObjectIDs []primitive.ObjectID
		for _, followerID := range habit.Followers {
			objectID, err := primitive.ObjectIDFromHex(followerID)
			if err != nil {
				log.Printf("Ошибка при преобразовании ID подписчика: %v", err)
				continue
			}
			followerObjectIDs = append(followerObjectIDs, objectID)
		}

		// Получаем информацию о всех подписчиках одним запросом
		cursor, err := h.habitsCollection.Find(ctx, bson.M{
			"_id": bson.M{"$in": followerObjectIDs},
		})
		if err != nil {
			return response, err
		}
		defer cursor.Close(ctx)

		// Собираем информацию о подписчиках
		for cursor.Next(ctx) {
			var followerHabit models.Habit
			if err := cursor.Decode(&followerHabit); err != nil {
				log.Printf("Ошибка при декодировании привычки подписчика: %v", err)
				continue
			}

			response.Followers = append(response.Followers, models.FollowerInfo{
				ID:            followerHabit.ID,
				TelegramID:    followerHabit.TelegramID,
				Title:         followerHabit.Title,
				LastClickDate: followerHabit.LastClickDate,
				Streak:        followerHabit.Streak,
				Score:         followerHabit.Score,
			})
		}
	}

	return response, nil
}
