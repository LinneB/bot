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

func TestVerifyMetadata(t *testing.T) {
	for _, c := range Handler.GetAllCommands() {
		valid := []struct {
			fieldName string
			valid     bool
		}{
			{"description", c.Metadata.Description != ""},
			{"cooldown", c.Metadata.Cooldown != 0},
			{"aliases", len(c.Metadata.Aliases) > 0},
			{"usage", c.Metadata.Usage != ""},
			{"minimumRole", c.Metadata.MinimumRole != 0},
			{"examples", len(c.Metadata.Examples) > 0},
		}
		for _, v := range valid {
			if !v.valid {
				t.Errorf("Missing metadata field %s in command %s", v.fieldName, c.Metadata.Name)
			}
		}
	}
}
