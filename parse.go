package main

import (
	"strings"
	"errors"
)

const trimSymbols = " !#$%^&*()~<>?,"

type CommandType uint
const (
	COMMAND_NONE  CommandType = iota
	COMMAND_START             = iota
	COMMAND_STOP              = iota
	COMMAND_STATUS            = iota
)

func (self CommandType) String() string {
	switch self {
	case COMMAND_NONE:
		return "none"
	case COMMAND_START:
		return "start"
	case COMMAND_STOP:
		return "stop"
	case COMMAND_STATUS:
		return "status"
	default:
		panic("CommandType switch not considering all cases!")
	}
}

type CommandTarget uint
const (
	COMMAND_TARGET_NONE CommandTarget = iota
	COMMAND_TARGET_PR                 = iota
)

func (self CommandTarget) String() string {
	switch self {
	case COMMAND_TARGET_NONE:
		return "none"
	case COMMAND_TARGET_PR:
		return "pr"
	default:
		panic("CommandTarget switch not considering all cases!")
	}
}

type Command struct {
	Type            CommandType
	Target          CommandTarget
	Mentions        []string
}

func ParseCommand(commandStr string) (*Command, error) {
	trimmedCommand := strings.Trim(strings.ToLower(string(commandStr)), trimSymbols)
	parts := strings.Split(trimmedCommand, " ");

	// filter empty entries
	filteredParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) > 0 {
			filteredParts = append(filteredParts, part)
		}
	}
	if len(filteredParts) == 0 {
		return nil, errors.New("The command was empty")
	}
	mentions := make([]string, 0, 10)
	
	// parse the command name
	var commandType CommandType = COMMAND_NONE
	var commandTarget CommandTarget = COMMAND_TARGET_NONE
	for _, part := range filteredParts {
		part = strings.Trim(part, trimSymbols)
		switch part {
		case "start":
			commandType = COMMAND_START
		case "stop":
			commandType = COMMAND_STOP
		case "status":
			commandType = COMMAND_STATUS

			// command targets
		case "pr":
			commandTarget = COMMAND_TARGET_PR
		default:
			if len(part) > 1 && part[0] == '@' {
				mentions = append(mentions, part)
			}
		}
	}
	return &Command{Mentions: mentions, Type: commandType, Target: commandTarget}, nil
}

