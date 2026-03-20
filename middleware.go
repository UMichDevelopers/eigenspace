package main

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

func eventMiddleware[T any](eventName string, handler func(*discordgo.Session, *T) error) func(*discordgo.Session, *T) {
	return func(session *discordgo.Session, event *T) {
		slog.Info("discord event received", "event", eventName, "data", spew.Sdump(event))

		if err := handler(session, event); err != nil {
			slog.Error("discord event handler failed", "event", eventName, "err", err)
		}
	}
}
