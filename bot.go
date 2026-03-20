package main

import (
	"github.com/bwmarrin/discordgo"
)

type bot struct {
	session *discordgo.Session
}

func newBot(cfg *Config) (*bot, error) {
	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, err
	}

	session.Identify.Intents = discordgo.IntentGuildMessages |
		discordgo.IntentDirectMessages |
		discordgo.IntentMessageContent

	bot := &bot{session: session}
	session.AddHandler(bot.handleMessageCreate)
	return bot, nil
}

func (b *bot) Open() error {
	return b.session.Open()
}

func (b *bot) Close() error {
	return b.session.Close()
}
