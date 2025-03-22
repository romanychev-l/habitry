package ping

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Структура для хранения пинга
type Ping struct {
	FollowerID       int64     `json:"follower_id" bson:"follower_id"`
	FollowerUsername string    `json:"follower_username" bson:"follower_username"`
	HabitID          string    `json:"habit_id" bson:"habit_id"`
	HabitTitle       string    `json:"habit_title" bson:"habit_title"`
	SenderID         int64     `json:"sender_id" bson:"sender_id"`
	SenderUsername   string    `json:"sender_username" bson:"sender_username"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
	Status           string    `json:"status" bson:"status"` // "pending", "sent", "error"
}

// Структура запроса для создания пинга
type CreatePingRequest struct {
	FollowerID       int64  `json:"follower_id"`
	FollowerUsername string `json:"follower_username"`
	HabitID          string `json:"habit_id"`
	HabitTitle       string `json:"habit_title"`
	SenderID         int64  `json:"sender_id"`
	SenderUsername   string `json:"sender_username"`
}

type Handler struct {
	pingsCollection *mongo.Collection
}

func NewHandler(pingsCollection *mongo.Collection) *Handler {
	return &Handler{
		pingsCollection: pingsCollection,
	}
}

// HandleCreatePing обрабатывает запрос на создание пинга
func (h *Handler) HandleCreatePing(w http.ResponseWriter, r *http.Request) {
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

	var request CreatePingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Ошибка при декодировании запроса", http.StatusBadRequest)
		return
	}

	// Проверяем, что все необходимые поля заполнены
	if request.FollowerID == 0 || request.HabitID == "" || request.SenderID == 0 {
		http.Error(w, "Не все обязательные поля заполнены", http.StatusBadRequest)
		return
	}

	// Создаем новый пинг
	ping := Ping{
		FollowerID:       request.FollowerID,
		FollowerUsername: request.FollowerUsername,
		HabitID:          request.HabitID,
		HabitTitle:       request.HabitTitle,
		SenderID:         request.SenderID,
		SenderUsername:   request.SenderUsername,
		CreatedAt:        time.Now(),
		Status:           "pending",
	}

	// Записываем пинг в базу данных
	_, err := h.pingsCollection.InsertOne(context.Background(), ping)
	if err != nil {
		log.Printf("Ошибка при сохранении пинга: %v", err)
		http.Error(w, "Ошибка при сохранении пинга", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Пинг успешно создан",
	})
}

// GetPendingPings возвращает список ожидающих отправки пингов
func (h *Handler) GetPendingPings(ctx context.Context) ([]Ping, error) {
	var pings []Ping

	cursor, err := h.pingsCollection.Find(ctx, bson.M{"status": "pending"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &pings); err != nil {
		return nil, err
	}

	return pings, nil
}

// UpdatePingStatus обновляет статус пинга
func (h *Handler) UpdatePingStatus(ctx context.Context, followerID int64, habitID string, senderID int64, status string) error {
	_, err := h.pingsCollection.UpdateOne(
		ctx,
		bson.M{
			"follower_id": followerID,
			"habit_id":    habitID,
			"sender_id":   senderID,
			"status":      "pending",
		},
		bson.M{
			"$set": bson.M{
				"status": status,
			},
		},
	)

	return err
}

// DeletePing удаляет пинг из базы данных
func (h *Handler) DeletePing(ctx context.Context, followerID int64, habitID string, senderID int64) error {
	_, err := h.pingsCollection.DeleteOne(
		ctx,
		bson.M{
			"follower_id": followerID,
			"habit_id":    habitID,
			"sender_id":   senderID,
		},
	)

	return err
}
