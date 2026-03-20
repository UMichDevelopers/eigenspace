package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/davecgh/go-spew/spew"
)

func run() error {
	if len(os.Args) != 2 {
		return errors.New("usage: eigenspace /path/to/eigenspace.conf")
	}

	cfg, err := loadConfig(os.Args[1])
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	bot, err := newBot(cfg)
	if err != nil {
		return fmt.Errorf("create discord bot: %w", err)
	}

	if err := bot.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}
	defer func() {
		if err := bot.Close(); err != nil {
			slog.Error("close discord session", "err", err)
		}
	}()

	slog.Info("discord bot is running")

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigch)

	sig := <-sigch
	slog.Info("shutdown signal received", "signal", sig.String())
	return nil
}

func main() {
	spew.Config.Indent = " "
	spew.Config.SortKeys = true
	spew.Config.DisableCapacities = true

	if err := run(); err != nil {
		slog.Error("bot exited with error", "err", err)
		os.Exit(1)
	}
}
