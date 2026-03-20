package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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
			log.Printf("close discord session: %v", err)
		}
	}()

	log.Printf("discord bot is running")

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigch)

	<-sigch
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
