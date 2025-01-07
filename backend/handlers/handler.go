package handlers

import (
	"github.com/go-telegram/bot"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	usersCollection   *mongo.Collection
	historyCollection *mongo.Collection
	habitsCollection  *mongo.Collection
	bot               *bot.Bot
}

func NewHandler(usersCollection, historyCollection, habitsCollection *mongo.Collection, b *bot.Bot) *Handler {
	return &Handler{
		usersCollection:   usersCollection,
		historyCollection: historyCollection,
		habitsCollection:  habitsCollection,
		bot:               b,
	}
}
