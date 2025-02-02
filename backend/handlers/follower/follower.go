package follower

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	followersCollection *mongo.Collection
}

func NewHandler(followersCollection *mongo.Collection) *Handler {
	return &Handler{
		followersCollection: followersCollection,
	}
}

func (h *Handler) HandleFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	telegramIDStr := r.URL.Query().Get("telegram_id")
	if telegramIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "telegram_id is required"})
		return
	}

	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid telegram_id"})
		return
	}

	// Находим все записи followers для данного пользователя
	cursor, err := h.followersCollection.Find(context.Background(), bson.M{
		"telegram_id": telegramID,
	})
	if err != nil {
		log.Printf("Error finding followers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var followers []bson.M
	if err = cursor.All(context.Background(), &followers); err != nil {
		log.Printf("Error decoding followers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"followers": followers,
	})
}

func (h *Handler) HandleHabitProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	habitID := r.URL.Query().Get("habit_id")
	telegramIDStr := r.URL.Query().Get("telegram_id")

	if habitID == "" || telegramIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "habit_id and telegram_id are required"})
		return
	}

	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid telegram_id"})
		return
	}

	// Получаем текущую дату
	today := time.Now().Format("2006-01-02")

	// Находим запись с фолловерами для данной привычки
	var result bson.M
	err = h.followersCollection.FindOne(context.Background(), bson.M{
		"telegram_id": telegramID,
		"habit_id":    habitID,
	}).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"total_followers": 0,
				"completed_today": 0,
				"progress":        0,
			})
			return
		}
		log.Printf("Error finding followers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	followers, ok := result["followers"].(primitive.A)
	if !ok || len(followers) == 0 {
		// Проверяем только владельца привычки
		completedToday := 0
		if lastClickDate, ok := result["last_click_date"].(string); ok && lastClickDate != "" {
			if lastClickDate[:10] == today {
				completedToday = 1
			}
		}

		progress := float64(completedToday)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_followers": 1, // Только владелец
			"completed_today": completedToday,
			"progress":        progress,
		})
		return
	}

	log.Printf("Найдено %d фолловеров", len(followers))

	// Создаем массив условий для поиска
	conditions := make([]bson.M, 0, len(followers))
	for _, follower := range followers {
		if f, ok := follower.(bson.M); ok {
			if telegramID, ok := f["telegram_id"].(int64); ok {
				if habitID, ok := f["habit_id"].(string); ok {
					conditions = append(conditions, bson.M{
						"telegram_id": telegramID,
						"habit_id":    habitID,
					})
				}
			}
		}
	}

	log.Printf("Поиск записей по условиям: %+v", conditions)

	// Получаем все записи фолловеров
	cursor, err := h.followersCollection.Find(context.Background(), bson.M{
		"$or": conditions,
	})
	if err != nil {
		log.Printf("Error finding follower records: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var followerRecords []bson.M
	if err = cursor.All(context.Background(), &followerRecords); err != nil {
		log.Printf("Error decoding follower records: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Найдено %d записей фолловеров", len(followerRecords))

	// Проверяем, выполнил ли владелец привычку
	ownerCompleted := false
	completedToday := 0
	if lastClickDate, ok := result["last_click_date"].(string); ok && lastClickDate != "" {
		if lastClickDate[:10] == today {
			ownerCompleted = true
			completedToday++
		}
	}

	log.Printf("Владелец %s привычку", map[bool]string{true: "выполнил", false: "не выполнил"}[ownerCompleted])

	// Подсчитываем количество выполнивших привычку сегодня
	totalFollowers := len(followers) + 1 // +1 для владельца
	for _, record := range followerRecords {
		if lastClickDate, ok := record["last_click_date"].(string); ok && lastClickDate != "" {
			if lastClickDate[:10] == today {
				completedToday++
			}
		}
	}

	progress := 0.0
	if totalFollowers > 0 {
		progress = float64(completedToday) / float64(totalFollowers)
	}

	log.Printf("Прогресс: %d из %d выполнили привычку (%.2f%%)",
		completedToday, totalFollowers, progress*100)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_followers": totalFollowers,
		"completed_today": completedToday,
		"progress":        progress,
	})
}

func (h *Handler) HandleUnfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TelegramID int64  `json:"telegram_id"`
		HabitID    string `json:"habit_id"`
		UnfollowID int64  `json:"unfollow_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// Находим запись с фолловерами для данной привычки
	var result bson.M
	err := h.followersCollection.FindOne(context.Background(), bson.M{
		"telegram_id": request.TelegramID,
		"habit_id":    request.HabitID,
	}).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "habit not found"})
			return
		}
		log.Printf("Error finding followers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Получаем текущий список фолловеров
	followers, ok := result["followers"].(primitive.A)
	if !ok {
		followers = primitive.A{}
	}

	// Создаем новый список фолловеров без отписавшегося пользователя
	newFollowers := make(primitive.A, 0, len(followers))
	for _, follower := range followers {
		if f, ok := follower.(bson.M); ok {
			if telegramID, ok := f["telegram_id"].(int64); ok {
				if telegramID != request.UnfollowID {
					newFollowers = append(newFollowers, follower)
				}
			}
		}
	}

	// Обновляем запись в базе данных
	_, err = h.followersCollection.UpdateOne(
		context.Background(),
		bson.M{
			"telegram_id": request.TelegramID,
			"habit_id":    request.HabitID,
		},
		bson.M{
			"$set": bson.M{
				"followers": newFollowers,
			},
		},
	)

	if err != nil {
		log.Printf("Error updating followers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
