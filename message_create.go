package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func (b *bot) handleMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author == nil || event.Author.Bot {
		return
	}

	command, ok := parseCommand(event.Content)
	if !ok {
		return
	}

	switch command.Command {
	case "PING":
		if _, err := session.ChannelMessageSend(event.ChannelID, "PONG"); err != nil {
			log.Printf("send message to channel %s: %v", event.ChannelID, err)
		}
	}
}
