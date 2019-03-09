package bot

import (
	"log"

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
	// bot.Debug = true

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
		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

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

	res, err := b.rtclient.NewCollection(u.Message.Text)
	if err != nil {
		msg.Text = err.Error()
		return
	}
	if res.Len() == 0 {
		return
	}
	b.stateManager.Set(userID, StateWaitCommand)

	b.sessionManager.Set(userID, Session{
		Results: res,
	})
	msg.Text = renderFiles(res)
	if !res.HasNext() {
		return
	}
	msg.ReplyMarkup = makeNextPageMarkup()
}

func (b *Bot) handleNextPageCallback(u tgbotapi.Update, userID int) {
	msg := u.CallbackQuery.Message

	s := b.sessionManager.Get(userID)
	if !s.Results.HasNext() {
		return
	}
	text := renderFiles(s.Results)

	editMsg := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, text)
	defer func() {
		b.bot.Send(editMsg)
	}()
	if !s.Results.HasNext() {
		return
	}
	editMsg.ReplyMarkup = makeNextPageMarkup()
}

func (b *Bot) handleWaitCommandState(u tgbotapi.Update, userID int) {
	if u.CallbackQuery != nil {
		b.handleNextPageCallback(u, userID)
		return
	}

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
		b.handleWaitCommandState(u, userID)
	default:
	}
}
