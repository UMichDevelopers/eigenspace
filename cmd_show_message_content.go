package main

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (b *bot) handleShowMessageContentCommand(session *discordgo.Session, event *discordgo.MessageCreate, command *ParsedCommand) error {
	if len(command.Args) != 1 {
		return errors.New("usage: %show-message-content <discord-message-url>")
	}

	channelID, messageID, err := parseDiscordMessageURL(command.Args[0], "usage: %show-message-content <discord-message-url>")
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
