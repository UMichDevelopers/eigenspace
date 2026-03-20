package main

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

func resolveShowMessageTarget(event *discordgo.MessageCreate, command *ParsedCommand, usage string) (string, string, error) {
	switch len(command.Args) {
	case 1:
		return event.ChannelID, command.Args[0], nil
	case 2:
		return command.Args[0], command.Args[1], nil
	default:
		return "", "", errors.New(usage)
	}
}
