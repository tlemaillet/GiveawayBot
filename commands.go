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
	aliases["lc"] = &Alias{
		Name:    "lc",
		Command: commands["listcommands"],
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
	commands["alliancecreate"] = &Command{
		Name:            "alliancecreate",
		Options:         "<alliance name> ",
		Description:     T("alliancecreate_command_desc"),
		NeedsGuildAdmin: true,
		NeedsGuild:      true,
		Callback:        createAllianceCommand,
	}
	aliases["ac"] = &Alias{
		Name:    "ac",
		Command: commands["alliancecreate"],
	}
	commands["addguildtoalliance"] = &Command{
		Name:            "addguildtoalliance",
		Options:         "<alliance name> ",
		Description:     T("addguildtoalliance_command_desc"),
		NeedsGuildAdmin: true,
		NeedsGuild:      true,
		Callback:        addGuildToAllianceCommand,
	}
	aliases["ag"] = &Alias{
		Name:    "ag",
		Command: commands["addguildtoalliance"],
	}
	commands["alliancedelete"] = &Command{
		Name:               "alliancedelete",
		Options:            "<alliance name> ",
		Description:        T("alliancedelete_command_desc"),
		NeedsAllianceAdmin: true,
		Callback:           deleteAllianceCommand,
	}
	aliases["ad"] = &Alias{
		Name:    "ad",
		Command: commands["alliancedelete"],
	}

	globalState.CommandList = commands
	globalState.AliasList = aliases
}

func getDefaultGabCommandsAndAliases() (commands map[string]string, aliases map[string]string) {
	commands = make(map[string]string)
	for commandKey, command := range globalState.CommandList {
		commands[commandKey] = command.Name
	}
	aliases = make(map[string]string)
	for aliasKey, alias := range globalState.AliasList {
		aliases[aliasKey] = alias.Name
	}

	return commands, aliases
}

func getfallbackCommandsAndAliases() (commands map[string]string, aliases map[string]string) {
	commands = make(map[string]string)
	aliases = make(map[string]string)
	commands["help"] = "inithelp"
	commands["alliancecreate"] = "alliancecreate"
	aliases["ac"] = "ac"
	commands["addguildtoalliance"] = "addguildtoalliance"
	aliases["ag"] = "ag"


	return commands, aliases
}

func listCommandsCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	var message = T("command_list") + "\n"

	for _, commandName := range alliance.State.Commands {
		command := globalState.CommandList[commandName]
		if command.Hidden {
			continue
		}

		message += alliance.State.GabPrefix + command.Name + "\n" +
			"\t" + command.Description + "\n"

		if command.Options != "" {
			message += "\t" + T("options", 2) + " : " +
				alliance.State.GabPrefix + command.Name + " " + command.Options + "\n"
		}

		if aliases, ok := alliance.State.AliasTable[command.Name]; ok {
			message += "\t" + T("alias", 2) + " : "
			for i, alias := range aliases {
				if i != 0 {
					message += ", "
				}
				message += alliance.State.GabPrefix + alias
			}
			message += "\n"
		}
	}
	s.ChannelMessageSend(c.ID, message)
}

func listNotesCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {
	err := s.UserNoteSet("137509464759599104", "Le Createur!")
	if err != nil {
		fmt.Println(err)
		return
	}
	for noteKey, note := range s.State.Notes {
		fmt.Println(noteKey + " : " + note)
	}
}

func createAllianceCommand(session *discordgo.Session, message *discordgo.MessageCreate, alliance *Alliance) {
	// Find the channel that the message came from.
	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	args, err := parseArguments(message.Content)
	if err != nil {
		session.ChannelMessageSend(channel.ID, T("createalliance_usage"))
		return
	}

	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		// Could not find channel.
		return
	}
	if len(args) != 2 || args[1] == "" {
		session.ChannelMessageSend(channel.ID, T("createalliance_usage"))
		return
	}
	allianceName := args[1]

	if alliance, exists := globalState.Alliances[allianceName]; exists {
		owner, err := session.User(alliance.Admin)
		if err != nil {
			session.ChannelMessageSend(channel.ID, T("error_creating_alliance"))
			return
		}

		session.ChannelMessageSend(channel.ID,
			T("alliance_already_exists",
				TInter{"Alliance": alliance.Name, "Owner": owner.Username}))

		return
	}

	newAlliance, err := createAlliance(allianceName, guild, message.Author)
	if err != nil {
		session.ChannelMessageSend(channel.ID, T("error_creating_alliance"))
		return
	}

	globalState.Alliances[allianceName] = newAlliance
	globalState.GuildTable = makeGuildTable(globalState.Alliances)
	session.ChannelMessageSend(channel.ID, T("created_alliance"))
}

func addGuildToAllianceCommand (session *discordgo.Session, message *discordgo.MessageCreate, alliance *Alliance) {
	channel, args, err := getChannelAndArgsFromMessage(session, message)
	if err != nil || args == nil || len(args) < 2 {
		session.ChannelMessageSend(channel.ID, T("addguildtoalliance_usage"))
		return
	}

	allianceName := args[1]
	if alliance, exists := globalState.Alliances[allianceName]; exists {
		guild , err := session.Guild(channel.GuildID)
		if err != nil {
			session.ChannelMessageSend(channel.ID, T("error_adding_guild"))
			return
		}

		globalState.Alliances[alliance.Name].Guilds = append(globalState.Alliances[alliance.Name].Guilds, guild.ID)
		globalState.GuildTable = makeGuildTable(globalState.Alliances)
		session.ChannelMessageSend(channel.ID, T("added_guild"))
	}
}

func deleteAllianceCommand(session *discordgo.Session, message *discordgo.MessageCreate, alliance *Alliance) {

	channel, args, err := getChannelAndArgsFromMessage(session, message)
	if err != nil || args == nil || len(args) < 2 {
		session.ChannelMessageSend(channel.ID, T("deletealliance_usage"))
		return
	}
	allianceName := args[1]

	delete(globalState.Alliances, allianceName)

	globalState.GuildTable = makeGuildTable(globalState.Alliances)

	session.ChannelMessageSend(channel.ID,
		T("deleted_alliance",
			TInter{"Alliance": allianceName}))
}

func startGiveawayCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {

	tmp := strings.Split(m.Content, " ")[1:]
	startMessage := strings.Join(tmp, " ")

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if alliance.State.Rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_already_active", TInter{"Game": alliance.State.Game}))
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

		alliance.State.Game = startArgs[0]
		alliance.State.GameKey = startArgs[1]
		alliance.State.Participants = map[string]Participant{}
		alliance.State.Rolling = true
		sendToAllianceGuilds(s, alliance,
			T("start_announcement",
				TInter{"Game": alliance.State.Game}))
	}
}

func stopGiveawayCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if !alliance.State.Rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_already_inactive"))
		return
	}
	alliance.State.Rolling = false
	if len(alliance.State.Participants) != 0 {
		winnerParticipant, _ := getWinnerFromParticipants(alliance.State.Participants)
		winner, err := s.User(winnerParticipant.UserID)
		if err != nil {
			sendToAllianceGuilds(s, alliance, T("winner_doesnt_exists"))
			return
		}


		sendToAllianceGuilds(s, alliance, T("winner_announcement_start"))
		time.Sleep(time.Second * 4)
		if rndm.Intn(30) != 0 {

			sendToAllianceGuilds(s, alliance, T("winner_announcement",
				TInter{"Person": winner.Username}))

		} else {
			sendToAllianceGuilds(s, alliance,
				T("winner_announcement_cena"))
			time.Sleep(time.Second * 20)
			sendToAllianceGuilds(s, alliance,
				T("winner_announcement_final",
					TInter{"Person": winner.Username}))
		}

		channel, _ := s.UserChannelCreate(winner.ID)

		s.ChannelMessageSend(channel.ID,
			T("winner_dm",
				TInter{"Person": winner.Username,
					"Key": alliance.State.GameKey}))

	} else {
		sendToAllianceGuilds(s, alliance, T("no_players"))
	}
}

func rollCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if !alliance.State.Rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_inactive"))
		return
	}

	if isGabCreator(m.Author) {
		sendToAllianceGuilds(s, alliance,
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

	commandName := strings.Replace(prefixCommand, alliance.State.GabPrefix, "", 1)

	if participant, exist := alliance.State.Participants[m.Author.ID]; exist {
		if participant.Need {

			s.ChannelMessageSend(c.ID,
				T("already_rolled_need",
					TInter{"Person": m.Author.Username,
						"Count": participant.Score}))
		} else {
			s.ChannelMessageSend(c.ID,
				T("already_rolled",
					TInter{"Person": m.Author.Username,
						"Count": participant.Score}))

		}
	} else if alliance, err := getAllianceFromMessage(s, m);
		err != nil && need && hasReachedNeedLimit(m.Author, alliance) {
		s.ChannelMessageSend(c.ID, T("reached_need_limit"))
	} else {
		roll := rndm.Intn(100)

		if need {
			err := addNeedTry(m.Author, alliance.State)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		alliance.State.Participants[m.Author.ID] = Participant{
			UserID:  m.Author.ID,
			Score: roll,
			Need:  need,
		}
		var tRollResult = "roll_result"
		if need {
			tRollResult = "roll_result_need"
		}

		switch commandName {
		case "roll", "r":
			sendToAllianceGuilds(s, alliance,
				T(tRollResult,
					TInter{"Person": m.Author.Username, "Count": roll}))
		case "localroll", "lr":
			s.ChannelMessageSend(c.ID,
				T(tRollResult,
					TInter{"Person": m.Author.Username, "Count": roll}))
		}
	}
}

func statusCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if alliance.State.Rolling {
		s.ChannelMessageSend(c.ID,
			T("giveaway_active", TInter{"Game": alliance.State.Game}))
	} else {
		s.ChannelMessageSend(c.ID,
			T("giveaway_inactive"))
	}
}

func notifyCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {
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

func unnotifyCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {
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

func talkCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {
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
	sendToAllianceGuilds(s, alliance, say)
}

func helpCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {

	helpMessage := strings.Join(strings.Split(m.Content, " ")[1:], " ")

	var message string

	switch helpMessage {
	case "greed", "need":
		message = T("need_help")
	case "alliances":
		message = T("alliances_help")
	default:
		message = T("usage")
	}

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	s.ChannelMessageSend(c.ID, message)
}

func initHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	s.ChannelMessageSend(c.ID,
		T("initHelp"))
}

func shoutCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	s.ChannelMessageSend(c.ID,
		T("shout_gab"))
}

func debugCommand(s *discordgo.Session, m *discordgo.MessageCreate, alliance *Alliance) {

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
