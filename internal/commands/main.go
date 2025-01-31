package commands

import (
	"bot/internal/models"
	"cmp"
	"slices"
	"strconv"
	"strings"
	"time"

	irc "github.com/gempir/go-twitch-irc/v4"
)

const (
	RGeneric = iota + 1
	RMod
	RBroadcaster
	RAdmin
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
	// Moderator
	IsMod bool
	// Broadcaster
	IsBroadcaster bool
	// Admin
	IsAdmin bool
	// Role
	Role int
}

func NewContext(state *models.State, msg irc.PrivateMessage) (context Context, err error) {
	SenderUserID, err := strconv.Atoi(msg.User.ID)
	if err != nil {
		return context, err
	}
	ChannelID, err := strconv.Atoi(msg.RoomID)
	if err != nil {
		return context, err
	}
	Arguments := strings.Fields(msg.Message)
	Invocation := strings.TrimPrefix(Arguments[0], state.Config.Prefix)
	isMod := msg.Tags["mod"] == "1"
	isBroadcaster := SenderUserID == ChannelID
	isAdmin := slices.Contains(state.Config.Admins, msg.User.Name)
	role := RGeneric
	if isMod {
		role = RMod
	}
	if isBroadcaster {
		role = RBroadcaster
	}
	if isAdmin {
		role = RAdmin
	}

	return Context{
		SenderUserID:      SenderUserID,
		SenderUsername:    msg.User.Name,
		SenderDisplayname: msg.User.DisplayName,
		ChannelID:         ChannelID,
		ChannelName:       msg.Channel,
		Message:           msg.Message,
		Arguments:         Arguments,
		Parameters:        Arguments[1:],
		Command:           strings.ToLower(Arguments[0]),
		Invocation:        strings.ToLower(Invocation),
		IsMod:             isMod,
		IsBroadcaster:     isBroadcaster,
		IsAdmin:           isAdmin,
		Role:              role,
	}, nil
}

type command struct {
	Run      func(state *models.State, ctx Context) (reply string, err error)
	Metadata metadata
}

type metadata struct {
	Name        string
	Description string
	// Optional text block with extended information.
	// Only displayed on website.
	ExtendedDescription string
	Cooldown            time.Duration
	MinimumRole         int
	Aliases             []string
	Usage               string
	Examples            []example
}

func (m metadata) PrettyRole() string {
	switch m.MinimumRole {
	case RAdmin:
		return "Admin"
	case RBroadcaster:
		return "Broadcaster"
	case RMod:
		return "Mod"
	}
	return "None"
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
			banned,
			cmd,
			followers,
			help,
			id,
			join,
			latestEmotes,
			live,
			notify,
			ping,
			randomEmote,
			subscribe,
			title,
			thumbnail,
		},
		Cooldowns: make(map[int]map[string]time.Time),
	}
}
