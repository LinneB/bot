package commands

import (
	"os"
	"testing"
)

func TestLoadedCommands(t *testing.T) {
	files, err := os.ReadDir(".")
	if err != nil {
		t.Errorf("Could not read command directory: %s", err)
	}
	cmdFiles := len(files) - 2 // Subtract 2 for "main.go" and "main_test.go"
	if len(Handler.Commands) != cmdFiles {
		t.Errorf("Missing command file: Expected %d; Loaded %d", cmdFiles, len(Handler.Commands))
	}
}
