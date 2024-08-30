package bot

import (
	"context"
	"fmt"
	"net/http"
	"pipe/internal/config"
	"pipe/internal/services"
	"time"

	"golang.org/x/net/proxy"
	"gopkg.in/telebot.v3"
)

type Telegram struct {
	App *services.App
	Bot *telebot.Bot

	ctx    context.Context
	cancel context.CancelFunc
}

func NewTelegram(ctx context.Context, app *services.App) (*Telegram, error) {
	ctx, cancel := context.WithCancel(ctx)
	t := &Telegram{
		App:    app,
		ctx:    ctx,
		cancel: cancel,
	}

	client, err := buildClientWithProxy(config.AppConfig.ProxyAddr)
	if err != nil {
		return nil, err
	}

	pref := telebot.Settings{
		Token:  config.AppConfig.Token,
		Poller: &telebot.LongPoller{Timeout: 30 * time.Second},
		Client: client,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	t.Bot = bot

	t.setupHandlers()
	return t, nil
}

func buildClientWithProxy(addr string) (*http.Client, error) {
	if addr != "" {
		var auth *proxy.Auth

		dialer, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		httpTransport := &http.Transport{Dial: dialer.Dial}
		hc := &http.Client{Transport: httpTransport}

		return hc, nil
	}

	return &http.Client{}, nil
}

func (t *Telegram) setupHandlers() {
	// middlewares
	// t.Bot.Use()

	// handlers
	t.Bot.Handle("/start", t.start)
}

func (t *Telegram) start(c telebot.Context) error {
	args := c.Message().Payload

	if args != "" {
		return c.Send(fmt.Sprintf("الان داری به %s پیام میدی", args), &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{
					{
						Text:   "Open",
						WebApp: &telebot.WebApp{URL: fmt.Sprintf("%s/sendMessage/%s", config.AppConfig.ClientURL, args)},
					},
				},
			},
		})
	}

	return c.Send(&telebot.Photo{Caption: `Welcome to Pipe.
Pipe is a Telegram Mini App with E2EE, Users can send hidden message to each other.`, File: telebot.FromDisk("assets/img/banner.png")}, &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{
					Text:   "Open",
					WebApp: &telebot.WebApp{URL: config.AppConfig.ClientURL},
				},
			},
			{
				{
					Text: "Community",
					URL:  "t.me/PipeChatCommunity",
				},
			},
		},
	})
}

func (t *Telegram) Start() {
	t.Bot.Start()
}

func (t *Telegram) Shutdown() {
	t.cancel()
	t.Bot.Stop()
}
