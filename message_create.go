package main

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

type commandHandler func(*discordgo.Session, *discordgo.MessageCreate, *ParsedCommand) error

func (b *bot) handleMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) error {
	if event.Author == nil || event.Author.Bot {
		return nil
	}

	command, ok := parseCommand(event.Content)
	if !ok {
		return nil
	}

	slog.Info(
		"discord command parsed",
		"channel_id", event.ChannelID,
		"guild_id", event.GuildID,
		"message_id", event.ID,
		"author_id", event.Author.ID,
		"command", command.Command,
		"arg_count", len(command.Args),
	)

	handlers := map[string]commandHandler{
		"PING": b.handlePingCommand,
		"SHOW-MESSAGE": b.requireRole(
			b.cfg.Discord.AdminRoleID,
			b.handleShowMessageCommand,
		),
		"SHOW-MESSAGE-CONTENT": b.requireRole(
			b.cfg.Discord.AdminRoleID,
			b.handleShowMessageContentCommand,
		),
	}

	handler, ok := handlers[command.Command]
	if !ok {
		return nil
	}

	err := handler(session, event, command)
	if err == nil {
		return nil
	}

	replyErr := reply(session, event, err.Error())
	if replyErr != nil {
		return errors.Join(err, replyErr)
	}

	return err
}

func (b *bot) requireRole(roleID uint64, handler commandHandler) commandHandler {
	requiredRoleID := strconv.FormatUint(roleID, 10)

	return func(session *discordgo.Session, event *discordgo.MessageCreate, command *ParsedCommand) error {
		if event.Member == nil {
			return errors.New("this command may only be used in a guild")
		}

		for _, roleID := range event.Member.Roles {
			if roleID == requiredRoleID {
				return handler(session, event, command)
			}
		}

		return errors.New("you do not have the required role for this command")
	}
}
