package main

import (
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

func (b *bot) handleShowMessageCommand(session *discordgo.Session, event *discordgo.MessageCreate, command *ParsedCommand) error {
	channelID, messageID, err := resolveShowMessageTarget(event, command, "usage: %show-message [channel-id] <message-id>")
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
