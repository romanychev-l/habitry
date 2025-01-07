package habit

import (
	"backend/models"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	habitsCollection  *mongo.Collection
	historyCollection *mongo.Collection
}

func NewHandler(habitsCollection, historyCollection *mongo.Collection) *Handler {
	return &Handler{
		habitsCollection:  habitsCollection,
		historyCollection: historyCollection,
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
	habit.IsShared = false
	habit.IsArchived = false
	habit.Participants = []models.HabitParticipant{
		{
			TelegramID: habitRequest.TelegramID,
			Streak:     0,
			Score:      0,
		},
	}

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

	// Возращаем созданную привычку
	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": createdHabit,
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

	// Находим участника
	participantIndex := -1
	for i, p := range habit.Participants {
		if p.TelegramID == habitRequest.TelegramID {
			participantIndex = i
			break
		}
	}

	if participantIndex == -1 {
		http.Error(w, "Участник не найден", http.StatusNotFound)
		return
	}

	// Обновляем статистику участника
	today := time.Now().Format("2006-01-02")
	lastClick := habit.Participants[participantIndex].LastClickDate

	if lastClick != today {
		// Если это первое выполнение или новый день
		habit.Participants[participantIndex].LastClickDate = today
		habit.Participants[participantIndex].Streak++
		habit.Participants[participantIndex].Score++

		// Обновляем привычку
		_, err = h.habitsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": habitID},
			bson.M{"$set": bson.M{"participants": habit.Participants}},
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Обновляем историю
		today := time.Now().Format("2006-01-02")
		update := bson.M{
			"$set": bson.M{
				"habits.$[habit].done": true,
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
			// Если записи в истории нет, создаем новую
			if err == mongo.ErrNoDocuments {
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
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	// Получаем обновленную привычку
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habit,
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

	// Находим участника
	participantIndex := -1
	for i, p := range habit.Participants {
		if p.TelegramID == habitRequest.TelegramID {
			participantIndex = i
			break
		}
	}

	if participantIndex == -1 {
		http.Error(w, "Участник не найден", http.StatusNotFound)
		return
	}

	// Обновляем статистику участника
	today := time.Now().Format("2006-01-02")
	if habit.Participants[participantIndex].LastClickDate == today {
		habit.Participants[participantIndex].LastClickDate = ""
		habit.Participants[participantIndex].Streak--
		habit.Participants[participantIndex].Score--

		// Обновляем привычку
		_, err = h.habitsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": habitID},
			bson.M{"$set": bson.M{"participants": habit.Participants}},
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

	// Получаем обновленную привычку
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habit,
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(request)

	// Получаем привычку
	habitID, err := primitive.ObjectIDFromHex(request.HabitID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(habitID)

	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Println(habit)

	// Проверяем, является ли пользователь создателем привычки
	if habit.CreatorID == request.TelegramID {
		// Если создатель - помечаем привычку как архивную
		_, err = h.habitsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": habitID},
			bson.M{"$set": bson.M{"is_archived": true}},
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Удаляем пользователя из списка участников
	_, err = h.habitsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": habitID},
		bson.M{
			"$pull": bson.M{
				"participants": bson.M{
					"telegram_id": request.TelegramID,
				},
			},
		},
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
		TelegramID int64  `json:"telegram_id"`
		HabitID    string `json:"habit_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем привычку
	habitID, err := primitive.ObjectIDFromHex(request.HabitID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Проверяем, не является ли пользователь уже участником
	for _, participant := range habit.Participants {
		if participant.TelegramID == request.TelegramID {
			// Если пользователь уже участник, просто возвращаем привычку
			json.NewEncoder(w).Encode(map[string]interface{}{
				"habit": habit,
			})
			return
		}
	}

	// Добавляем нового участника
	newParticipant := models.HabitParticipant{
		TelegramID:    request.TelegramID,
		LastClickDate: "",
		Streak:        0,
		Score:         0,
	}

	// Обновляем привычку
	_, err = h.habitsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": habitID},
		bson.M{
			"$push": bson.M{"participants": newParticipant},
			"$set":  bson.M{"is_shared": true},
		},
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем обновленную привычку
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitID}).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": habit,
	})
}
