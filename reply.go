package main

import "github.com/bwmarrin/discordgo"

func reply(session *discordgo.Session, event *discordgo.MessageCreate, content string) error {
	_, err := session.ChannelMessageSendComplex(
		event.ChannelID,
		&discordgo.MessageSend{
			Content:   content,
			Reference: event.Reference(),
		},
	)
	return err
}
