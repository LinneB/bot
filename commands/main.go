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
	//     "ping": {
	//         123: "last-execution-time"
	//     }
	// }
	Cooldowns map[string]map[int]time.Time
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
	lastExecution, found := h.Cooldowns[command.Metadata.Name][id]
	if !found {
		return false
	}
	if lastExecution.Add(command.Metadata.Cooldown).After(time.Now()) {
		return true
	}
	return false
}

func (h *handler) SetCooldown(id int, command *command) {
	h.Cooldowns[command.Metadata.Name][id] = time.Now()
}

var Handler *handler

func init() {
	Handler = &handler{
		Commands: []command{
			ping,
		},
		Cooldowns: make(map[string]map[int]time.Time),
	}
	for _, command := range Handler.Commands {
		Handler.Cooldowns[command.Metadata.Name] = make(map[int]time.Time)
	}
}
