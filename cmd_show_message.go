package main

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

func (b *bot) handleShowMessageCommand(session *discordgo.Session, event *discordgo.MessageCreate, command *ParsedCommand) error {
	var channelID string
	var messageID string

	switch len(command.Args) {
	case 1:
		channelID = event.ChannelID
		messageID = command.Args[0]
	case 2:
		channelID = command.Args[0]
		messageID = command.Args[1]
	default:
		return errors.New("usage: %show-message [channel-id] <message-id>")
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
		"message "+messageID+" from channel "+channelID,
		"message-"+messageID+".txt",
		strings.NewReader(dump),
	)
	return err
}
