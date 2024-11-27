package bot

import (
	"github.com/go-telegram/bot"
)

func New(token string) (*bot.Bot, error) {
	return bot.New(token)
}
