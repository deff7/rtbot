package bot

import (
	"log"
	"strconv"

	"github.com/deff7/rutracker/internal/rutracker"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

type Bot struct {
	bot            *tgbotapi.BotAPI
	updatesCh      tgbotapi.UpdatesChannel
	rtclient       *rutracker.Client
	stateManager   StateManager
	sessionManager SessionManager
}

func NewBot(token string, rtclient *rutracker.Client) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.Wrap(err, "initializing bot api")
	}
	//bot.Debug = true

	_, err = bot.RemoveWebhook()
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updatesCh, err := bot.GetUpdatesChan(u)
	return &Bot{
		bot:            bot,
		updatesCh:      updatesCh,
		stateManager:   newStateManager(),
		sessionManager: newSessionManager(),
		rtclient:       rtclient,
	}, err
}

func (b *Bot) Run() {
	for update := range b.updatesCh {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		b.handleUpdate(update)
	}
}

func (b *Bot) handleStartState(u tgbotapi.Update, userID int) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Привет, напиши запрос")
	b.stateManager.Set(userID, StateWaitQuery)
	b.bot.Send(msg)
}

func (b *Bot) handleWaitQueryState(u tgbotapi.Update, userID int) {
	if u.Message == nil {
		return
	}

	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Нет результатов")
	defer func() {
		b.bot.Send(msg)
	}()

	b.stateManager.Set(userID, StateWaitCommand)
	res, err := b.rtclient.NewCollection(u.Message.Text)
	if err != nil {
		msg.Text = err.Error()
		return
	}
	b.sessionManager.Set(userID, Session{
		Results: res,
	})

	sep := ""
	msg.Text = ""
	for _, f := range res.ListNext() {
		msg.Text += sep + f.Name + "\n/download" + strconv.Itoa(f.ID)
		sep = "\n\n"
	}

	if !res.HasNext() {
		return
	}
	rows := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Следующая страница", "/next"),
	}
	markup := tgbotapi.NewInlineKeyboardMarkup(rows)
	msg.ReplyMarkup = markup
}

func (b *Bot) handleWaitCommandState(u tgbotapi.Update, userID int) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Вообще я жду /next или /download1234")
	b.stateManager.Set(userID, StateWaitQuery)
	b.bot.Send(msg)
}

func (b *Bot) handleUpdate(u tgbotapi.Update) {
	userID, err := getUserID(u)
	if err != nil {
		log.Print(err)
		return
	}
	state := b.stateManager.Get(userID)

	switch state {
	case StateStart:
		b.handleStartState(u, userID)
	case StateWaitQuery:
		b.handleWaitQueryState(u, userID)
	case StateWaitCommand:
	default:
	}
}
