package main

import (
	"errors"
	"log"
	"strings"
)

const trimSymbols = " !#$%^&*()~<>?,"

// CommandType is a top level command that needs to have a target
// associated with it to be applied
type CommandType uint

const (

	// CommandNone is a wild card to detect unknown command
	CommandNone CommandType = iota

	// CommandStart command starts 'something'
	CommandStart

	// CommandStop command stops 'something'
	CommandStop

	// CommandStatus command gets the status of 'something'
	CommandStatus
)

func (commandType CommandType) String() string {
	switch commandType {
	case CommandNone:
		return "none"
	case CommandStart:
		return "start"
	case CommandStop:
		return "stop"
	case CommandStatus:
		return "status"
	default:
		log.Println("CommandType switch not considering all cases!")
		return ""
	}
}

// CommandTarget is the set of possible targets for a command
type CommandTarget uint

const (

	// CommandTargetNone is a wild card used to detect an unhandled target
	CommandTargetNone CommandTarget = iota

	// CommandTargetPR is the PullRequestService target name
	CommandTargetPR

	// CommandTargetMtFlow is this program :)
	CommandTargetMtFlow
)

func (commandTarget CommandTarget) String() string {
	switch commandTarget {
	case CommandTargetNone:
		return "none"
	case CommandTargetPR:
		return "pr"
	case CommandTargetMtFlow:
		return "mtflow"
	default:
		log.Println("CommandTarget switch not considering all cases!")
		return ""
	}
}

// Command is the metadata about the type, target, and whom the
// command was issued to
type Command struct {
	Type     CommandType
	Target   CommandTarget
	Mentions []string
}

// ParseCommand takes care of turning a  text string into a proper command
func ParseCommand(commandStr string) (*Command, error) {
	trimmedCommand := strings.Trim(strings.ToLower(string(commandStr)), trimSymbols)
	parts := strings.Split(trimmedCommand, " ")

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
	var commandType = CommandNone
	var commandTarget = CommandTargetNone
	for _, part := range filteredParts {
		part = strings.Trim(part, trimSymbols)
		switch part {
		case "start":
			commandType = CommandStart
		case "stop":
			commandType = CommandStop
		case "status":
			commandType = CommandStatus

			// command targets
		case "pr":
			commandTarget = CommandTargetPR
		case "mtflow":
			commandTarget = CommandTargetMtFlow
		default:
			if len(part) > 1 && part[0] == '@' {
				mentions = append(mentions, part)
			}
		}
	}
	return &Command{Mentions: mentions, Type: commandType, Target: commandTarget}, nil
}
