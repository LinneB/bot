package commands

import (
	"bot/models"
	"cmp"
	"slices"
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
	Usage       string
	Examples    []example
}

type example struct {
	Description string
	Command     string
	Response    string
}

type handler struct {
	Commands []command
	// map of latest command executions:
	// {
	//     123: {
	//         "command": "last-execution-time"
	//     }
	// }
	Cooldowns map[int]map[string]time.Time
}

// Attempts to get a command based on an alias.
func (h *handler) GetCommandByAlias(alias string) (command command, found bool) {
	for _, c := range h.Commands {
		if slices.Contains(c.Metadata.Aliases, alias) {
			return c, true
		}
	}
	return command, false
}

func (h *handler) GetCommandByName(name string) (command command, found bool) {
	for _, c := range h.Commands {
		if c.Metadata.Name == name {
			return c, true
		}
	}
	return command, false
}

func (h *handler) GetAllCommands() []command {
	commands := h.Commands
	slices.SortFunc(commands,
		func(a, b command) int {
			return cmp.Compare(a.Metadata.Name, b.Metadata.Name)
		},
	)
	return commands
}

func (h *handler) IsOnCooldown(id int, name string, cooldown time.Duration) bool {
	if _, found := h.Cooldowns[id]; !found {
		return false
	}
	lastExecution, found := h.Cooldowns[id][name]
	if !found {
		return false
	}
	if lastExecution.Add(cooldown).After(time.Now()) {
		return true
	}
	return false
}

func (h *handler) SetCooldown(id int, name string) {
	if _, found := h.Cooldowns[id]; !found {
		h.Cooldowns[id] = make(map[string]time.Time)
	}
	h.Cooldowns[id][name] = time.Now()
}

var Handler *handler

func init() {
	Handler = &handler{
		Commands: []command{
			help,
			id,
			live,
			ping,
		},
		Cooldowns: make(map[int]map[string]time.Time),
	}
}
