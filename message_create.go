package main

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type commandHandler func(*discordgo.Session, *discordgo.MessageCreate, *ParsedCommand) error

func (b *bot) handleMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) error {
	if event.Author == nil || event.Author.Bot {
		return nil
	}

	command, ok := parseCommand(event.Content)
	if !ok {
		return nil
	}

	slog.Info(
		"discord command parsed",
		"channel_id", event.ChannelID,
		"guild_id", event.GuildID,
		"message_id", event.ID,
		"author_id", event.Author.ID,
		"command", command.Command,
		"arg_count", len(command.Args),
	)

	handlers := map[string]commandHandler{
		"PING": b.handlePingCommand,
	}

	handler, ok := handlers[command.Command]
	if !ok {
		return nil
	}

	return handler(session, event, command)
}
