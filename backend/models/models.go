package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InvoiceResponse struct {
	URL string `json:"url"`
}

type Follower struct {
	TelegramID int64  `bson:"telegram_id" json:"telegram_id"`
	HabitID    string `bson:"habit_id" json:"habit_id"`
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

type Habit struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Title        string             `bson:"title" json:"title"`
	WantToBecome string             `bson:"want_to_become" json:"want_to_become"`
	Days         []int              `bson:"days" json:"days"`
	IsOneTime    bool               `bson:"is_one_time" json:"is_one_time"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	CreatorID    int64              `bson:"creator_id" json:"creator_id"`
}

type HabitRequest struct {
	TelegramID int64 `json:"telegram_id"`
	Habit      Habit `json:"habit"`
}

type HabitHistory struct {
	HabitID string `bson:"habit_id" json:"habit_id"`
	Title   string `bson:"title" json:"title"`
	Done    bool   `bson:"done" json:"done"`
}

type History struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	TelegramID int64              `bson:"telegram_id" json:"telegram_id"`
	Date       string             `bson:"date" json:"date"`
	Habits     []HabitHistory     `bson:"habits" json:"habits"`
}

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	TelegramID int64              `bson:"telegram_id" json:"telegram_id"`
	Username   string             `bson:"username" json:"username"`
	FirstName  string             `bson:"first_name" json:"first_name"`
	Language   string             `bson:"language_code" json:"language_code"`
	PhotoURL   string             `bson:"photo_url" json:"photo_url"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	Credit     int                `bson:"credit" json:"credit"`
	LastVisit  string             `bson:"last_visit" json:"last_visit"`
	Timezone   string             `bson:"timezone" json:"timezone"`
}

type HabitWithStats struct {
	Habit         `bson:"habit" json:"habit"`
	LastClickDate string `bson:"last_click_date" json:"last_click_date"`
	Streak        int    `bson:"streak" json:"streak"`
	Score         int    `bson:"score" json:"score"`
}
