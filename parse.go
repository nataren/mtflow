package main

import (
	"strings"
	"errors"
)

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
	Prefix          string
}

func ParseCommand(commandStr string) (*Command, error) {
	trimmedCommand := strings.Trim(strings.ToLower(string(commandStr)), " ")
	parts := strings.Split(trimmedCommand, " ");
	if len(parts) == 0 {
		return nil, errors.New("The command was empty")
	}
	
	// parse the prefix
	prefix := strings.Trim(parts[0], " ,")
	if len(prefix) < 2 || !strings.HasPrefix(prefix, "@") {
		return nil, errors.New("The command did not start with @username")
	}
	parts = parts[1:]

	// parse the command name
	var commandType CommandType = COMMAND_NONE
outerLoop:
	for i, part := range parts {
		switch part {
		case "start":
			commandType = COMMAND_START
			parts = parts[i+1:]
			break outerLoop
		case "stop":
			commandType = COMMAND_STOP
			parts = parts[i+1:]
			break outerLoop
		case "status":
			commandType = COMMAND_STATUS
			parts = parts[i+1:]
			break outerLoop
		default:
			continue
		}
	}
	if commandType == COMMAND_NONE {
		return nil, errors.New("Could not find command type")
	}

	// parse command target
	var commandTarget CommandTarget = COMMAND_TARGET_NONE
outerLoop2:
	for i, part := range parts {
		switch part {
		case "pr":
			commandTarget = COMMAND_TARGET_PR
			parts = parts[i+1:]
			break outerLoop2
		default:
			continue
		}
	}
	if commandTarget == COMMAND_TARGET_NONE {
		return nil, errors.New("Could not find command target")
	}
	return &Command{Prefix: prefix, Type: commandType, Target: commandTarget}, nil
}

