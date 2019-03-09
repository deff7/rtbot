package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

func getUserID(u tgbotapi.Update) (int, error) {
	if u.Message != nil {
		return u.Message.From.ID, nil
	}

	if u.CallbackQuery != nil {
		return u.CallbackQuery.From.ID, nil
	}

	return -1, errors.New("can't get User ID")
}
