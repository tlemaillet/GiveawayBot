package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Name               string
	Options            string
	Description        string
	NeedsCreator       bool
	NeedsAllianceAdmin bool
	NeedsAllianceMod   bool
	NeedsGuild         bool
	NeedsGuildOwner    bool
	NeedsGuildAdmin    bool
	Hidden             bool
	Callback           func(session *discordgo.Session, message *discordgo.MessageCreate, alliance *Alliance)
}
type Commands map[string]*Command

type Alias struct {
	Name    string
	Command *Command
}
type Aliases map[string]*Alias

type Participant struct {
	UserID  string
	Score int
	Need  bool
}
type Participants map[string]Participant

type State struct {
	GabPrefix string

	Commands   map[string]string
	Aliases    map[string]string
	AliasTable map[string][]string

	Game    string
	GameKey string

	Rolling      bool
	Participants map[string]Participant
	NeedLimit    int
	NeedState  NeedState
}
type Alliance struct {
	Name       string
	Admin      string
	State      *State
	Moderators []string
	MainGuild  string
	Guilds     []string
}

type PersitentGlobalState struct  {
	Alliances       map[string]*Alliance
}

type GlobalState struct {
	Alliances       map[string]*Alliance
	GuildTable		map[string][]string
	DataDirectory   string
	GlobalStateFile string
	CommandList     map[string]*Command
	AliasList       map[string]*Alias
	BotToken        string
}

type NeedEntry struct {
	Game string
	Date time.Time
}
type NeedState map[string][]NeedEntry
