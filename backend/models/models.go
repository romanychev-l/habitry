package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InvoiceResponse struct {
	URL string `json:"url"`
}

type Habit struct {
	ID            string    `bson:"id" json:"id"`
	Title         string    `bson:"title" json:"title"`
	Score         int       `bson:"score" json:"score"`
	Streak        int       `bson:"streak" json:"streak"`
	Days          []int     `bson:"days" json:"days"`
	LastClickDate string    `bson:"last_click_date" json:"last_click_date"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	IsOneTime     bool      `bson:"is_one_time" json:"is_one_time"`
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

type DayHistory struct {
	Date   string         `bson:"date" json:"date"`
	Habits []HabitHistory `bson:"habits" json:"habits"`
}

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	TelegramID int64              `bson:"telegram_id" json:"telegram_id"`
	Username   string             `bson:"username" json:"username"`
	FirstName  string             `bson:"first_name" json:"first_name"`
	Language   string             `bson:"language_code" json:"language_code"`
	PhotoURL   string             `bson:"photo_url" json:"photo_url"`
	Habits     []Habit            `bson:"habits" json:"habits"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	Credit     int                `bson:"credit" json:"credit"`
	LastVisit  string             `bson:"last_visit" json:"last_visit"`
	History    []DayHistory       `bson:"history" json:"history"`
}
