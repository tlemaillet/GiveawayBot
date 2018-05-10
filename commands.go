package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func initCommandList() {
	commands := make(Commands)
	aliases := make(Aliases)

	commands["start"] = &Command{
		Name:             "start",
		Options:          "[game name] [key]",
		Description:      T("start_command_desc"),
		NeedsAllianceMod: true,
		Callback:         startGiveawayCommand,
	}
	commands["stop"] = &Command{
		Name:             "stop",
		Options:          "",
		Description:      T("stop_command_desc"),
		NeedsAllianceMod: true,
		Callback:         stopGiveawayCommand,
	}
	commands["debug"] = &Command{
		Name:         "debug",
		Options:      "[dm|mention]",
		Description:  T("debug_command_desc"),
		NeedsCreator: true,
		Callback:     debugCommand,
	}
	commands["roll"] = &Command{
		Name:        "roll",
		Options:     "[greed|need]",
		Description: T("roll_command_desc"),
		Callback:    rollCommand,
	}
	aliases["r"] = &Alias{
		Name:    "r",
		Command: commands["roll"],
	}
	commands["localroll"] = &Command{
		Name:        "localroll",
		Options:     "[greed|need]",
		Description: T("localroll_command_desc"),
		Callback:    rollCommand,
	}
	aliases["lr"] = &Alias{
		Name:    "lr",
		Command: commands["localroll"],
	}
	commands["status"] = &Command{
		Name:        "status",
		Options:     "",
		Description: T("status_command_desc"),
		Callback:    statusCommand,
	}
	aliases["current"] = &Alias{
		Name:    "current",
		Command: commands["status"],
	}
	commands["listcommands"] = &Command{
		Name:        "listcommands",
		Options:     "",
		Description: T("listcommands_command_desc"),
		Callback:    listCommandsCommand,
	}
	commands["talk"] = &Command{
		Name:        "talk",
		Options:     "message",
		Description: T("talk_command_desc"),
		Callback:    talkCommand,
	}
	commands["help"] = &Command{
		Name:        "help",
		Options:     "",
		Description: T("help_command_desc"),
		Callback:    helpCommand,
	}
	commands["inithelp"] = &Command{
		Name:        "help",
		Options:     "",
		Hidden:      true,
		Description: T("inithelp_command_desc"),
		Callback:    initHelpCommand,
	}
	aliases["info"] = &Alias{
		Name:    "info",
		Command: commands["help"],
	}
	commands["notify"] = &Command{
		Name:        "notify",
		Options:     "",
		Description: T("notify_command_desc"),
		NeedsGuild:  true,
		Callback:    notifyCommand,
	}
	commands["unnotify"] = &Command{
		Name:        "unnotify",
		Options:     "",
		Description: T("unnotify_command_desc"),
		NeedsGuild:  true,
		Callback:    unnotifyCommand,
	}
	commands["listnotes"] = &Command{
		Name:         "listnotes",
		Options:      "",
		Description:  T("listnotes_command_desc"),
		NeedsCreator: true,
		Hidden:       true,
		Callback:     listNotesCommand,
	}

	globalState.CommandList = commands
	globalState.AliasList = aliases
}

func getDefaultGabCommandsAndAliases() (commands Commands, aliases Aliases) {
	commands = globalState.CommandList
	aliases = globalState.AliasList

	return commands, aliases
}

func getfallbackCommandsAndAliases() (commands Commands, aliases Aliases) {
	commands = make(Commands)
	aliases = make(Aliases)
	commands["help"] = globalState.CommandList["inithelp"]

	return commands, aliases
}

func listCommandsCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	var message = T("command_list") + "\n"

	for _, command := range state.Commands {
		if command.Hidden {
			continue
		}

		message += state.GabPrefix + command.Name + "\n" +
			"\t" + command.Description + "\n"

		if command.Options != "" {
			message += "\t" + T("options", 2) + " : " +
				state.GabPrefix + command.Name + " " + command.Options + "\n"
		}

		if aliases, ok := state.AliasTable[command.Name]; ok {
			message += "\t" + T("alias", 2) + " : "
			for i, alias := range aliases {
				if i != 0 {
					message += ", "
				}
				message += state.GabPrefix + alias
			}
			message += "\n"
		}
	}
	s.ChannelMessageSend(c.ID, message)
}

func listNotesCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {
	err := s.UserNoteSet("137509464759599104", "Le Createur!")
	if err != nil {
		fmt.Println(err)
		return
	}
	for noteKey, note := range s.State.Notes {
		fmt.Println(noteKey + " : " + note)
	}
}

func startGiveawayCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {

	tmp := strings.Split(m.Content, " ")[1:]
	startMessage := strings.Join(tmp, " ")

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if state.Rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_already_active", TInter{"Game": state.Game}))
		return
	}
	startArgs, err := parseArguments(startMessage)
	if err != nil {
		s.ChannelMessageSend(c.ID, T("start_usage"))
		return
	}
	for _, arg := range startArgs {
		fmt.Println(arg)
	}

	if len(startArgs) != 2 {
		s.ChannelMessageSend(c.ID, T("start_usage"))
	} else {
		state.Game = startArgs[0]
		state.GameKey = startArgs[1]
		state.Participants = map[string]*Participant{}
		state.Rolling = true
		sendToAllGuilds(s,
			T("start_announcement",
				TInter{"Game": state.Game}))
	}
}

func stopGiveawayCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if !state.Rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_already_inactive"))
		return
	}
	state.Rolling = false
	if len(state.Participants) != 0 {
		winnerParticipant, _ := getWinnerFromParticipants(state.Participants)
		winner := winnerParticipant.User

		sendToAllGuilds(s, T("winner_announcement_start"))
		time.Sleep(time.Second * 4)
		if rndm.Intn(30) != 0 {

			sendToAllGuilds(s, T("winner_announcement",
				TInter{"Person": winner.Username}))

		} else {
			sendToAllGuilds(s,
				T("winner_announcement_cena"))
			time.Sleep(time.Second * 20)
			sendToAllGuilds(s,
				T("winner_announcement_final",
					TInter{"Person": winner.Username}))
		}

		channel, _ := s.UserChannelCreate(winner.ID)

		s.ChannelMessageSend(channel.ID,
			T("winner_dm",
				TInter{"Person": winner.Username,
					"Key": state.GameKey}))

	} else {
		sendToAllGuilds(s, T("no_players"))
	}
}

func rollCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if !state.Rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_inactive"))
		return
	}

	if isGabCreator(m.Author) {
		sendToAllGuilds(s,
			T("roll_result",
				TInter{"Person": m.Author.Username, "Count": "∞"}))
		return
	}

	message := strings.Split(m.Content, " ")
	prefixCommand := message[0]

	var need bool

	if len(message) >= 2 {
		switch message[1] {
		case "greed":
			need = false
		case "need":
			need = true
		default:
			s.ChannelMessageSend(c.ID,
				T("syntax_error", TInter{"Command": message[0], "Option": message[1]}))
			return
		}
	}

	commandName := strings.Replace(prefixCommand, state.GabPrefix, "", 1)

	if participant, exist := state.Participants[m.Author.ID]; exist {
		if participant.Need {

			s.ChannelMessageSend(c.ID,
				T("already_rolled_need",
					TInter{"Person": participant.User.Username,
						"Count": participant.Score}))
		} else {
			s.ChannelMessageSend(c.ID,
				T("already_rolled",
					TInter{"Person": participant.User.Username,
						"Count": participant.Score}))

		}
	} else if need && hasReachedNeedLimit(m.Author, state) {
		s.ChannelMessageSend(c.ID, T("reached_need_limit"))
	} else {
		roll := rndm.Intn(100)

		if need {
			err := addNeedTry(m.Author, state)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		state.Participants[m.Author.ID] = &Participant{
			User:  m.Author,
			Score: roll,
			Need:  need,
		}

		switch commandName {
		case "roll", "r":
			sendToAllGuilds(s,
				T("roll_result",
					TInter{"Person": m.Author.Username, "Count": roll}))
		case "localroll", "lr":
			s.ChannelMessageSend(c.ID,
				T("roll_result",
					TInter{"Person": m.Author.Username, "Count": roll}))
		}
	}
}

func statusCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if state.Rolling {
		s.ChannelMessageSend(c.ID,
			T("giveaway_active", TInter{"Game": state.Game}))
	} else {
		s.ChannelMessageSend(c.ID,
			T("giveaway_inactive"))
	}
}

func notifyCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	userGuild, err := s.Guild(c.GuildID)
	if err != nil {
		fmt.Println(err)
		fmt.Println(err.Error())
		return
	}
	perm, err := hasRolePermissions(s, c)
	if err != nil {
		fmt.Println(err)
		fmt.Println(err.Error())
		return
	}
	if perm {
		s.ChannelMessageSend(c.ID, T("not_enough_permissions"))
		return
	}

	user, _ := s.User(m.Author.ID)
	err = addNotifyRoleToUser(s, userGuild, user)
	if err != nil {
		s.ChannelMessageSend(c.ID, T("define_notify_role"))
		return
	}

	s.ChannelMessageSend(c.ID, T("added_notify_role"))
}

func unnotifyCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	userGuild, err := s.Guild(c.GuildID)
	if err != nil {
		fmt.Println(err)
		fmt.Println(err.Error())
		return
	}
	perm, err := hasRolePermissions(s, c)
	if err != nil {
		fmt.Println(err)
		fmt.Println(err.Error())
		return
	}
	if perm {
		s.ChannelMessageSend(c.ID, T("not_enough_permissions"))
		return
	}

	user, _ := s.User(m.Author.ID)
	err = removeNotifyRoleFromUser(s, userGuild, user)
	if err != nil {
		return
	}

	s.ChannelMessageSend(c.ID, T("removed_notify_role"))
}

func talkCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {
	message := strings.Join(strings.Split(m.Content, " ")[1:], " ")

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	var gName string
	g, err := s.Guild(c.GuildID)
	if err != nil {
		gName = T("nowhere")
	} else {
		gName = g.Name
	}
	say := T("gab_talk",
		TInter{"Author": m.Author.Username,
			"Server": gName,
			"Message": message})
	sendToAllGuilds(s, say)
}

func helpCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	s.ChannelMessageSend(c.ID,
		T("usage"))
}

func initHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	s.ChannelMessageSend(c.ID,
		T("initHelp"))
}

func shoutCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	s.ChannelMessageSend(c.ID,
		T("shout_gab"))
}

func debugCommand(s *discordgo.Session, m *discordgo.MessageCreate, state *State) {

	debugMessage := strings.Join(strings.Split(m.Content, " ")[1:], " ")

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	debugArgs, err := parseArguments(debugMessage)
	if err != nil {
		s.ChannelMessageSend(c.ID, T("debug_usage"))
		return
	}

	switch debugArgs[0] {
	case "dm":
		channel, _ := s.UserChannelCreate(m.Author.ID)

		s.ChannelMessageSend(channel.ID,
			T("mention_person",
				TInter{"Person": m.Author.Mention()}))
	case "mention":
		s.ChannelMessageSend(c.ID,
			T("mention_person",
				TInter{"Person": m.Author.Mention()}))

	default:
		s.ChannelMessageSend(c.ID, T("debug_usage"))

	}
}
