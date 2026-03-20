package main

import (
	"log/slog"

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
		_, err := session.ChannelMessageSend(
			event.ChannelID,
			"usage: %show-message [channel-id] <message-id>",
		)
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

	_, err = session.ChannelMessageSend(
		event.ChannelID,
		"logged details for message "+messageID+" in channel "+channelID,
	)
	return err
}
