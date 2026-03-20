package main

import "github.com/bwmarrin/discordgo"

func (b *bot) handlePingCommand(session *discordgo.Session, event *discordgo.MessageCreate, command *ParsedCommand) error {
	return reply(session, event, "PONG")
}
