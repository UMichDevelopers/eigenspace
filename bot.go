package main

import (
	"github.com/bwmarrin/discordgo"
)

type bot struct {
	session *discordgo.Session
	cfg     *Config
}

func newBot(cfg *Config) (*bot, error) {
	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, err
	}

	session.Identify.Intents = discordgo.IntentGuildMessages |
		discordgo.IntentGuildMessagePolls |
		discordgo.IntentDirectMessages |
		discordgo.IntentMessageContent

	bot := &bot{
		session: session,
		cfg:     cfg,
	}
	session.AddHandler(eventMiddleware("connect", bot.handleConnect))
	session.AddHandler(eventMiddleware("disconnect", bot.handleDisconnect))
	session.AddHandler(eventMiddleware("ready", bot.handleReady))
	session.AddHandler(eventMiddleware("message_create", bot.handleMessageCreate))
	session.AddHandler(eventMiddleware("message_poll_vote_add", bot.handleMessagePollVoteAdd))
	session.AddHandler(eventMiddleware("message_poll_vote_remove", bot.handleMessagePollVoteRemove))
	return bot, nil
}

func (b *bot) Open() error {
	return b.session.Open()
}

func (b *bot) Close() error {
	return b.session.Close()
}
