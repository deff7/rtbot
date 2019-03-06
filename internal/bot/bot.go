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
		rtclient:       rtclient,
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
		reply  = ""
	)

	switch state {
	case StateStart:
		reply = "Привет! Напиши запрос"
		b.stateManager.Set(userID, StateWaitQuery)
	case StateWaitQuery:
		b.stateManager.Set(userID, StateWaitCommand)
		res, err := b.rtclient.NewCollection(msg.Text)
		if err != nil {
			reply = err.Error()
			break
		}
		b.sessionManager.Set(userID, Session{
			Results: res,
		})
		sep := ""
		for _, f := range res.ListNext() {
			reply += sep + f.Name + "\n/download" + strconv.Itoa(f.ID)
			sep = "\n\n"
		}
	case StateWaitCommand:
		reply = "Вообще я жду /next или /download1234"
		b.stateManager.Set(userID, StateWaitQuery)
	default:
	}

	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, reply)
	b.bot.Send(replyMsg)
}
