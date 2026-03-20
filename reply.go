package main

import (
	"io"

	"github.com/bwmarrin/discordgo"
)

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

func replyFile(session *discordgo.Session, event *discordgo.MessageCreate, content string, name string, reader io.Reader) error {
	_, err := session.ChannelMessageSendComplex(
		event.ChannelID,
		&discordgo.MessageSend{
			Content:   content,
			Reference: event.Reference(),
			Files: []*discordgo.File{
				{
					Name:   name,
					Reader: reader,
				},
			},
		},
	)
	return err
}
