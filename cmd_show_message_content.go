package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (b *bot) handleShowMessageContentCommand(session *discordgo.Session, event *discordgo.MessageCreate, command *ParsedCommand) error {
	channelID, messageID, err := resolveShowMessageTarget(event, command, "usage: %show-message-content [channel-id] <message-id>")
	if err != nil {
		return err
	}

	message, err := session.ChannelMessage(channelID, messageID)
	if err != nil {
		return err
	}

	return replyFile(
		session,
		event,
		"",
		"message-content-"+messageID+".txt",
		strings.NewReader(message.Content),
	)
}
