package main

import (
	"testing"
)

func TestParse_empty_command(t *testing.T) {
	_, err := ParseCommand("    ")
	if err == nil {
		t.Error("Empty command should have error")
	}
}

func TestParse_without_prefix(t *testing.T) {
	_, err := ParseCommand("pr start")
	if err == nil {
		t.Error("Command with no prefix should have error")
	}
}

func TestParse_command_type(t *testing.T) {
	comm, err := ParseCommand("   @nataren,    please    give me the status of pr   ")
	if err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if comm.Type != COMMAND_STATUS {
		t.Error("Expected type status but was: ", comm.Type)
	}
	if comm.Prefix != "@nataren" {
		t.Error("Expected prefix @nataren but was: ", comm.Prefix)
	}
	if comm.Target != COMMAND_TARGET_PR {
		t.Error("Expected type pr but was: ", comm.Target)
	}
}
