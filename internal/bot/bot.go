package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

type Bot struct {
	bot            *tgbotapi.BotAPI
	updatesCh      tgbotapi.UpdatesChannel
	stateManager   StateManager
	sessionManager SessionManager
}

func NewBot(token string) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.Wrap(err, "initializing bot api")
	}
	bot.Debug = true

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
	}, err
}

func (b *Bot) Run() {
	for update := range b.updatesCh {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		b.handleMessage(update.Message)
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	var (
		userID = msg.From.ID
		state  = b.stateManager.Get(userID)
		reply  = "Фигня какая-то"
	)

	switch state {
	case StateStart:
		reply = "Привет! Напиши запрос"
		b.stateManager.Set(userID, StateWaitQuery)
	case StateWaitQuery:
		reply = "Я типа ответ прислал"
		b.stateManager.Set(userID, StateWaitCommand)
	case StateWaitCommand:
		reply = "Вообще я жду /next или /download1234"
		b.stateManager.Set(userID, StateWaitQuery)
	default:
	}

	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, reply)
	b.bot.Send(replyMsg)
}
