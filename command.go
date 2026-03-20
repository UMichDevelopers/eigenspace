package main

import "strings"

type ParsedCommand struct {
	Command string
	Args    []string
}

func parseCommand(content string) (*ParsedCommand, bool) {
	if !strings.HasPrefix(content, "%") {
		return nil, false
	}

	body := strings.TrimPrefix(content, "%")
	if body == "" || strings.HasPrefix(body, " ") {
		return nil, false
	}

	head, tail, hasTail := strings.Cut(body, " :")
	fields := strings.Fields(head)
	if len(fields) == 0 {
		return nil, false
	}

	args := append([]string{}, fields[1:]...)
	if hasTail {
		args = append(args, tail)
	}

	return &ParsedCommand{
		Command: strings.ToUpper(fields[0]),
		Args:    args,
	}, true
}
