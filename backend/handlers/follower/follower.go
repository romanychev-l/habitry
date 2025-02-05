package follower

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
)

type Handler struct {
	habitsCollection *mongo.Collection
	usersCollection  *mongo.Collection
}

func NewHandler(habitsCollection, usersCollection *mongo.Collection) *Handler {
	return &Handler{
		habitsCollection: habitsCollection,
		usersCollection:  usersCollection,
	}
}

// HandleGetFollowers возвращает список подписчиков для привычки
func (h *Handler) HandleGetFollowers(w http.ResponseWriter, r *http.Request) {
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

	habitID := r.URL.Query().Get("habit_id")
	if habitID == "" {
		http.Error(w, "habit_id обязателен", http.StatusBadRequest)
		return
	}

	// Получаем привычку
	habitObjectID, err := primitive.ObjectIDFromHex(habitID)
	if err != nil {
		http.Error(w, "Неверный формат habit_id", http.StatusBadRequest)
		return
	}

	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{"_id": habitObjectID}).Decode(&habit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Привычка не найдена", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем информацию о подписчиках
	var followers []models.FollowerInfo
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

		// Получаем привычки подписчиков
		cursor, err := h.habitsCollection.Find(context.Background(), bson.M{
			"_id": bson.M{"$in": followerObjectIDs},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Собираем информацию о подписчиках
		for cursor.Next(context.Background()) {
			var followerHabit models.Habit
			if err := cursor.Decode(&followerHabit); err != nil {
				log.Printf("Ошибка при декодировании привычки подписчика: %v", err)
				continue
			}

			// Получаем информацию о пользователе
			var user models.User
			err := h.usersCollection.FindOne(
				context.Background(),
				bson.M{"telegram_id": followerHabit.TelegramID},
			).Decode(&user)

			if err == nil {
				followers = append(followers, models.FollowerInfo{
					ID:            followerHabit.ID,
					TelegramID:    followerHabit.TelegramID,
					Title:         followerHabit.Title,
					LastClickDate: followerHabit.LastClickDate,
					Streak:        followerHabit.Streak,
					Score:         followerHabit.Score,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

// HandleGetFollowing возвращает список привычек, на которые подписан пользователь
func (h *Handler) HandleGetFollowing(w http.ResponseWriter, r *http.Request) {
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

	telegramIDStr := r.URL.Query().Get("telegram_id")
	if telegramIDStr == "" {
		http.Error(w, "telegram_id обязателен", http.StatusBadRequest)
		return
	}

	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Неверный формат telegram_id", http.StatusBadRequest)
		return
	}

	// Получаем все привычки пользователя
	cursor, err := h.habitsCollection.Find(context.Background(), bson.M{
		"telegram_id": telegramID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var userHabits []models.Habit
	if err = cursor.All(context.Background(), &userHabits); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Собираем все ID привычек, на которые подписан пользователь
	var followingHabits []models.HabitResponse
	for _, habit := range userHabits {
		if len(habit.Followers) > 0 {
			// Получаем информацию о привычках, на которые подписан пользователь
			for _, followerID := range habit.Followers {
				followerObjectID, err := primitive.ObjectIDFromHex(followerID)
				if err != nil {
					continue
				}

				var followedHabit models.Habit
				err = h.habitsCollection.FindOne(
					context.Background(),
					bson.M{"_id": followerObjectID},
				).Decode(&followedHabit)

				if err == nil {
					// Получаем информацию о создателе привычки
					var creator models.User
					err = h.usersCollection.FindOne(
						context.Background(),
						bson.M{"telegram_id": followedHabit.TelegramID},
					).Decode(&creator)

					if err == nil {
						habitResponse := models.HabitResponse{
							ID:            followedHabit.ID,
							TelegramID:    followedHabit.TelegramID,
							Title:         followedHabit.Title,
							WantToBecome:  followedHabit.WantToBecome,
							Days:          followedHabit.Days,
							IsOneTime:     followedHabit.IsOneTime,
							CreatedAt:     followedHabit.CreatedAt,
							LastClickDate: followedHabit.LastClickDate,
							Streak:        followedHabit.Streak,
							Score:         followedHabit.Score,
							Followers:     []models.FollowerInfo{}, // Можно добавить при необходимости
						}
						followingHabits = append(followingHabits, habitResponse)
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followingHabits)
}

func (h *Handler) HandleHabitProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	habitID := r.URL.Query().Get("habit_id")
	log.Printf("habitID: %s", habitID)

	if habitID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "habit_id and telegram_id are required"})
		return
	}

	// Получаем текущую дату
	today := time.Now().Format("2006-01-02")

	// Преобразуем habitID в ObjectID
	habitObjectID, err := primitive.ObjectIDFromHex(habitID)
	if err != nil {
		log.Printf("Error converting habit_id to ObjectID: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Находим привычку
	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{
		"_id": habitObjectID,
	}).Decode(&habit)
	log.Printf("Found habit: %+v", habit)

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
		log.Printf("Error finding habit: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Проверяем, выполнил ли владелец привычку
	ownerCompleted := false
	completedToday := 0
	if habit.LastClickDate == today {
		ownerCompleted = true
		completedToday++
	}

	log.Printf("Owner completed: %v", ownerCompleted)

	// Подсчитываем количество подписчиков, выполнивших привычку сегодня
	totalFollowers := len(habit.Followers) + 1 // +1 для владельца
	if len(habit.Followers) > 0 {
		// Преобразуем строковые ID в ObjectID
		var followerObjectIDs []primitive.ObjectID
		for _, followerID := range habit.Followers {
			objectID, err := primitive.ObjectIDFromHex(followerID)
			if err != nil {
				log.Printf("Error converting follower ID: %v", err)
				continue
			}
			followerObjectIDs = append(followerObjectIDs, objectID)
		}

		// Получаем привычки подписчиков
		cursor, err := h.habitsCollection.Find(context.Background(), bson.M{
			"_id": bson.M{"$in": followerObjectIDs},
		})
		if err != nil {
			log.Printf("Error finding follower habits: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Подсчитываем выполненные привычки
		for cursor.Next(context.Background()) {
			var followerHabit models.Habit
			if err := cursor.Decode(&followerHabit); err != nil {
				log.Printf("Error decoding follower habit: %v", err)
				continue
			}
			if followerHabit.LastClickDate == today {
				completedToday++
			}
		}
	}

	progress := 0.0
	if totalFollowers > 0 {
		progress = float64(completedToday) / float64(totalFollowers)
	}

	log.Printf("Progress: %d out of %d completed (%f%%)", completedToday, totalFollowers, progress*100)

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
		HabitID    string `json:"habit_id"`
		UnfollowID int64  `json:"unfollow_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Ошибка при декодировании запроса: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	log.Printf("Получен запрос на отписку: %+v", request)

	// Преобразуем ID привычки в ObjectID
	habitObjectID, err := primitive.ObjectIDFromHex(request.HabitID)
	if err != nil {
		log.Printf("Ошибка при преобразовании ID привычки: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid habit_id"})
		return
	}

	var habit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{
		"_id": habitObjectID,
	}).Decode(&habit)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Привычка не найдена: %v", err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "habit not found"})
			return
		}
		log.Printf("Ошибка при поиске привычки: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Получаем все привычки из массива followers
	if len(habit.Followers) > 0 {
		var followerObjectIDs []primitive.ObjectID
		for _, followerID := range habit.Followers {
			objectID, err := primitive.ObjectIDFromHex(followerID)
			if err != nil {
				log.Printf("Ошибка при преобразовании ID подписчика: %v", err)
				continue
			}
			followerObjectIDs = append(followerObjectIDs, objectID)
		}

		// Получаем привычки подписчиков
		cursor, err := h.habitsCollection.Find(context.Background(), bson.M{
			"_id": bson.M{"$in": followerObjectIDs},
		})
		if err != nil {
			log.Printf("Ошибка при поиске привычек подписчиков: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Создаем массив ID привычек, которые нужно удалить из followers
		var habitsToRemove []string
		for cursor.Next(context.Background()) {
			var followerHabit models.Habit
			if err := cursor.Decode(&followerHabit); err != nil {
				log.Printf("Ошибка при декодировании привычки подписчика: %v", err)
				continue
			}

			// Если привычка принадлежит пользователю, который отписывается
			if followerHabit.TelegramID == request.UnfollowID {
				habitsToRemove = append(habitsToRemove, followerHabit.ID.Hex())
			}
		}

		// Удаляем найденные привычки из массива followers
		if len(habitsToRemove) > 0 {
			_, err = h.habitsCollection.UpdateOne(
				context.Background(),
				bson.M{
					"_id": habitObjectID,
				},
				bson.M{
					"$pull": bson.M{
						"followers": bson.M{
							"$in": habitsToRemove,
						},
					},
				},
			)

			if err != nil {
				log.Printf("Ошибка при обновлении списка подписчиков: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			log.Printf("Успешно удалили привычки %v из followers привычки %s", habitsToRemove, request.HabitID)
		}
	}

	// Получаем обновленную привычку
	var updatedHabit models.Habit
	err = h.habitsCollection.FindOne(context.Background(), bson.M{
		"_id": habitObjectID,
	}).Decode(&updatedHabit)

	if err != nil {
		log.Printf("Ошибка при получении обновленной привычки: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"habit": updatedHabit,
	})
}
