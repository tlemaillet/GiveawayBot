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
	Callback           func(session *discordgo.Session, message *discordgo.MessageCreate, state *State)
}
type Commands map[string]*Command

type Alias struct {
	Name    string
	Command *Command
}
type Aliases map[string]*Alias

type Participant struct {
	User  *discordgo.User
	Score int
	Need  bool
}
type Participants map[string]*Participant

type State struct {
	GabPrefix string

	Commands   map[string]*Command
	Aliases    map[string]*Alias
	AliasTable map[string][]string

	Game    string
	GameKey string

	Rolling      bool
	Participants map[string]*Participant
	NeedLimit    int
}
type Alliance struct {
	Name       string
	Admin      string
	State      *State
	NeedState  *NeedState
	Moderators []string
	Guilds     map[string]*discordgo.Guild
}

type GlobalState struct {
	Alliances     map[string]*Alliance
	DataDirectory string
	CommandList   map[string]*Command
	AliasList     map[string]*Alias
	BotToken      string
}

type NeedEntry struct {
	Game string
	Date time.Time
}
type NeedState map[string][]NeedEntry
