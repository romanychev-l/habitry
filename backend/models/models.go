package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InvoiceResponse struct {
	URL string `json:"url"`
}

type Follower struct {
	TelegramID    int64  `bson:"telegram_id" json:"telegram_id"`
	HabitID       string `bson:"habit_id" json:"habit_id"`
	Streak        int    `bson:"streak" json:"streak"`
	Score         int    `bson:"score" json:"score"`
	LastClickDate string `bson:"last_click_date" json:"last_click_date"`
}

type HabitFollowers struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	TelegramID    int64              `bson:"telegram_id" json:"telegram_id"`
	HabitID       string             `bson:"habit_id" json:"habit_id"`
	LastClickDate string             `bson:"last_click_date" json:"last_click_date"`
	Streak        int                `bson:"streak" json:"streak"`
	Score         int                `bson:"score" json:"score"`
	Followers     []Follower         `bson:"followers" json:"followers"`
}

// Habit - основная структура для хранения привычки в БД
type Habit struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	TelegramID    int64              `bson:"telegram_id" json:"telegram_id"`
	Title         string             `bson:"title" json:"title"`
	WantToBecome  string             `bson:"want_to_become" json:"want_to_become"`
	Days          []int              `bson:"days" json:"days"`
	IsOneTime     bool               `bson:"is_one_time" json:"is_one_time"`
	IsAuto        bool               `bson:"is_auto" json:"is_auto"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	LastClickDate string             `bson:"last_click_date" json:"last_click_date"`
	Streak        int                `bson:"streak" json:"streak"`
	Score         int                `bson:"score" json:"score"`
	Stake         int                `bson:"stake" json:"stake"`
	Followers     []string           `bson:"followers" json:"followers,omitempty"` // ID других привычек
}

// HabitResponse - структура для отправки данных на фронтенд
type HabitResponse struct {
	ID            primitive.ObjectID `json:"_id"`
	TelegramID    int64              `json:"telegram_id"`
	Title         string             `json:"title"`
	WantToBecome  string             `json:"want_to_become"`
	Days          []int              `json:"days"`
	IsOneTime     bool               `json:"is_one_time"`
	IsAuto        bool               `json:"is_auto"`
	CreatedAt     time.Time          `json:"created_at"`
	LastClickDate string             `json:"last_click_date"`
	Streak        int                `json:"streak"`
	Score         int                `json:"score"`
	Stake         int                `json:"stake"`
	Followers     []FollowerInfo     `json:"followers"` // Обогащенная информация о подписчиках
	Progress      float64            `json:"progress"`
}

// FollowerInfo - информация о подписчике для отправки на фронтенд
type FollowerInfo struct {
	ID             primitive.ObjectID `json:"_id"`
	TelegramID     int64              `json:"telegram_id"`
	Title          string             `json:"title"`
	LastClickDate  string             `json:"last_click_date"`
	Streak         int                `json:"streak"`
	Score          int                `json:"score"`
	Username       string             `json:"username"`
	FirstName      string             `json:"first_name"`
	PhotoURL       string             `json:"photo_url"`
	IsMutual       bool               `json:"is_mutual"`
	CompletedToday bool               `json:"completed_today"`
}

type HabitRequest struct {
	TelegramID int64 `json:"telegram_id"`
	Habit      Habit `json:"habit"`
}

type HabitHistory struct {
	HabitID primitive.ObjectID `bson:"habit_id" json:"habit_id"`
	Title   string             `bson:"title" json:"title"`
	Done    bool               `bson:"done" json:"done"`
}

type History struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	TelegramID int64              `bson:"telegram_id" json:"telegram_id"`
	Date       string             `bson:"date" json:"date"`
	Habits     []HabitHistory     `bson:"habits" json:"habits"`
}

type User struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	TelegramID           int64              `bson:"telegram_id" json:"telegram_id"`
	FirstName            string             `bson:"first_name" json:"first_name"`
	Username             string             `bson:"username" json:"username"`
	LanguageCode         string             `bson:"language_code" json:"language_code"`
	PhotoURL             string             `bson:"photo_url" json:"photo_url"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	Balance              int                `bson:"balance" json:"balance"`
	LastVisit            string             `bson:"last_visit" json:"last_visit"`
	Timezone             string             `bson:"timezone" json:"timezone"`
	NotificationsEnabled bool               `bson:"notifications_enabled" json:"notifications_enabled"`
	NotificationTime     string             `bson:"notification_time" json:"notification_time"`
}

type UserResponseWithHabits struct {
	ID                   primitive.ObjectID `json:"_id,omitempty"`
	TelegramID           int64              `json:"telegram_id"`
	Username             string             `json:"username"`
	FirstName            string             `json:"first_name"`
	LanguageCode         string             `json:"language_code"`
	PhotoURL             string             `json:"photo_url"`
	CreatedAt            time.Time          `json:"created_at"`
	Balance              int                `json:"balance"`
	LastVisit            string             `json:"last_visit"`
	Timezone             string             `json:"timezone"`
	NotificationsEnabled bool               `json:"notifications_enabled"`
	NotificationTime     string             `json:"notification_time"`
	Habits               []HabitResponse    `json:"habits"`
}

func (u *User) ToResponseWithHabits(habitResponses []HabitResponse) UserResponseWithHabits {
	return UserResponseWithHabits{
		ID:                   u.ID,
		TelegramID:           u.TelegramID,
		Username:             u.Username,
		FirstName:            u.FirstName,
		LanguageCode:         u.LanguageCode,
		PhotoURL:             u.PhotoURL,
		CreatedAt:            u.CreatedAt,
		Balance:              u.Balance,
		LastVisit:            u.LastVisit,
		Timezone:             u.Timezone,
		NotificationsEnabled: u.NotificationsEnabled,
		NotificationTime:     u.NotificationTime,
		Habits:               habitResponses, // Просто передаем готовый слайс
	}
}
