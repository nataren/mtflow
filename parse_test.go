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

func TestParse_unrecognized_command(t *testing.T) {
	comm, err := ParseCommand(" lsdjhfl rwiuhgirehg erhgiuerhg oehrgioerhg,!#$% ")
	if err != nil {
		t.Error("No error expected")
	}
	if comm.Target != COMMAND_TARGET_NONE {
		t.Error("command target none expexted")
	}
	if comm.Type != COMMAND_NONE {
		t.Error("command type none expexted")
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
	if comm.Mentions[0] != "@nataren" {
		t.Error("Expected prefix @nataren but was: ", comm.Mentions)
	}
	if comm.Target != COMMAND_TARGET_PR {
		t.Error("Expected type pr but was: ", comm.Target)
	}
}

func TestParse_command_with_user_parsing(t *testing.T) {
	comm, err := ParseCommand("   @nataren,    @yurig please    !@manuel! give me the status of pr   ")
	if err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if comm.Type != COMMAND_STATUS {
		t.Error("Expected type status but was: ", comm.Type)
	}
	if comm.Mentions[0] != "@nataren" {
		t.Error("Expected prefix @nataren but was: ", comm.Mentions[0])
	}
	if comm.Mentions[1] != "@yurig" {
		t.Error("Expected prefix @nataren but was: ", comm.Mentions[0])
	}
	if comm.Mentions[2] != "@manuel" {
		t.Error("Expected prefix @nataren but was: ", comm.Mentions[0])
	}
	if comm.Target != COMMAND_TARGET_PR {
		t.Error("Expected type pr but was: ", comm.Target)
	}
}
