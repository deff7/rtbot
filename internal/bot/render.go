package bot

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func renderFiles(files RTCollection) string {
	var out, sep string
	for _, f := range files.ListNext() {
		out += sep + f.Name + "\n/download" + strconv.Itoa(f.ID)
		sep = "\n\n"
	}
	return out
}

func makeNextPageMarkup() *tgbotapi.InlineKeyboardMarkup {
	rows := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Следующая страница", "/next"),
	}
	m := tgbotapi.NewInlineKeyboardMarkup(rows)
	return &m
}
