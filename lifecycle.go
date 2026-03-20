package main

import (
	"github.com/bwmarrin/discordgo"
)

func (b *bot) handleConnect(session *discordgo.Session, event *discordgo.Connect) error {
	return nil
}

func (b *bot) handleDisconnect(session *discordgo.Session, event *discordgo.Disconnect) error {
	return nil
}

func (b *bot) handleReady(session *discordgo.Session, event *discordgo.Ready) error {
	return nil
}
