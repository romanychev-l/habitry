package ping

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
func (h *Handler) HandleCreatePing(c *gin.Context) {
	var request CreatePingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Проверяем, что все необходимые поля заполнены
	if request.FollowerID == 0 || request.HabitID == "" || request.SenderID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save ping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ping created successfully",
		"ping":    ping,
	})
}
