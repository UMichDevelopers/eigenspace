package main

import (
	"errors"
	"strings"
)

func parseDiscordMessageURL(s string, usage string) (string, string, error) {
	const prefix = "https://discord.com/channels/"

	if !strings.HasPrefix(s, prefix) {
		return "", "", errors.New(usage)
	}

	parts := strings.Split(strings.TrimPrefix(s, prefix), "/")
	if len(parts) != 3 || parts[1] == "" || parts[2] == "" {
		return "", "", errors.New(usage)
	}

	return parts[1], parts[2], nil
}
