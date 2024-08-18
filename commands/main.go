package commands

import (
	"bot/models"
	"slices"
	"strings"
	"time"
)

// Command execution context
type Context struct {
	// Sender information
	SenderUserID      int
	SenderUsername    string
	SenderDisplayname string
	// Channel information
	ChannelID   int
	ChannelName string
	// Full message
	Message string
	// Message split into words
	Arguments []string
	// Message split into words, with command removed
	Parameters []string
	// Command alias used including prefix
	Command string
	// Command alias used excluding prefix
	Invocation string
}

type command struct {
	Run      func(state *models.State, ctx Context) (reply string, err error)
	Metadata metadata
}

type metadata struct {
	Name        string
	Description string
	Cooldown    time.Duration
	Aliases     []string
}

type handler struct {
	Commands []command
	// map of latest command executions:
	// {
	//     123: {
	//         "ping": "some-time"
	//     }
	// }
	Cooldowns map[int]map[string]time.Time
}

func (h *handler) GetCommandByAlias(prefix string, alias string) (command command, found bool) {
	if !strings.HasPrefix(alias, prefix) {
		return command, false
	}

	// Strip prefix
	commandName := strings.TrimPrefix(alias, prefix)
	for _, c := range h.Commands {
		if slices.Contains(c.Metadata.Aliases, commandName) {
			return c, true
		}
	}
	return command, false
}

func (h *handler) IsOnCooldown(id int, command *command) bool {
	_, found := h.Cooldowns[id]
	if !found {
		return false
	}
	cooldown, found := h.Cooldowns[id][command.Metadata.Name]
	if !found {
		return false
	}
	if cooldown.Add(command.Metadata.Cooldown).After(time.Now()) {
		return true
	}
	return false
}

func (h *handler) SetCooldown(id int, command *command) {
	_, found := h.Cooldowns[id]
	if !found {
		h.Cooldowns[id] = make(map[string]time.Time)
	}
	h.Cooldowns[id][command.Metadata.Name] = time.Now()
}

var Handler *handler

func init() {
	Handler = &handler{
		Commands: []command{
			ping,
		},
		Cooldowns: make(map[int]map[string]time.Time),
	}
}
