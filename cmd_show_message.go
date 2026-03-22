package main

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

func (b *bot) handleShowMessageCommand(session *discordgo.Session, event *discordgo.MessageCreate, command *ParsedCommand) error {
	if len(command.Args) != 1 {
		return errors.New("usage: %show-message <discord-message-url>")
	}

	channelID, messageID, err := parseDiscordMessageURL(command.Args[0], "usage: %show-message <discord-message-url>")
	if err != nil {
		return err
	}

	message, err := session.ChannelMessage(channelID, messageID)
	if err != nil {
		return err
	}

	slog.Info(
		"discord historical message fetched",
		"requested_by_message_id", event.ID,
		"requested_by_channel_id", event.ChannelID,
		"requested_by_author_id", event.Author.ID,
		"resolved_channel_id", channelID,
		"resolved_message_id", messageID,
		"data", spew.Sdump(message),
	)

	dump := spew.Sdump(message)
	err = replyFile(
		session,
		event,
		"",
		"message-"+messageID+".txt",
		strings.NewReader(dump),
	)
	return err
}
