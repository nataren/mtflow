package main

import (
	"testing"
)

func TestParse_empty_command(t *testing.T) {
	_, err := ParseCommand("    ", "1")
	if err == nil {
		t.Error("Empty command should have error")
	}
}

func TestParse_unrecognized_command(t *testing.T) {
	comm, err := ParseCommand(" lsdjhfl rwiuhgirehg erhgiuerhg oehrgioerhg,!#$% ", "2")
	if err != nil {
		t.Error("No error expected")
	}
	if comm.Target != CommandTargetNone {
		t.Error("command target none expexted")
	}
	if comm.Type != CommandNone {
		t.Error("command type none expexted")
	}
}

func TestParse_command_type(t *testing.T) {
	comm, err := ParseCommand("   @nataren,    please    give me the status of pr   ", "3")
	if err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if comm.Type != CommandStatus {
		t.Error("Expected type status but was: ", comm.Type)
	}
	if comm.Mentions[0] != "@nataren" {
		t.Error("Expected prefix @nataren but was: ", comm.Mentions)
	}
	if comm.Target != CommandTargetPR {
		t.Error("Expected type pr but was: ", comm.Target)
	}
}

func TestParse_command_with_user_parsing(t *testing.T) {
	comm, err := ParseCommand("   @nataren,    @yurig please    !@manuel! give me the status of pr   ", "4")
	if err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if comm.Type != CommandStatus {
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
	if comm.Target != CommandTargetPR {
		t.Error("Expected type pr but was: ", comm.Target)
	}
}

func TestParse_command_with_search(t *testing.T) {
	comm, err := ParseCommand(" @HAL please search for term1 term2 term3  ", "5")
	if err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if comm.Type != CommandSearch {
		t.Error("Expected type 'search' but was: ", comm.Type)
	}
	if comm.Mentions[0] != "@hal" {
		t.Error("Expected prefix @HAL but was: ", comm.Mentions[0])
	}
	if len(comm.Trailing) != 4 {
		t.Error("Expected 4 trailing arguments")
	}
}
