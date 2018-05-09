package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func getDefaultGabCommandsAndAliases() (commands GabCommands) {
	commands = make(GabCommands)
	commands["start"] = &GabCommand{
		"start",
		"[game name] [key]",
		T("start_command_desc"),
		true,
		false,
		false,
		startGiveawayCommand,
	}
	commands["stop"] = &GabCommand{
		"stop",
		"",
		T("stop_command_desc"),
		true,
		false,
		false,
		stopGiveawayCommand,
	}
	commands["debug"] = &GabCommand{
		"debug",
		"[dm|mention]",
		T("debug_command_desc"),
		true,
		false,
		false,
		debugCommand,
	}
	commands["roll"] = &GabCommand{
		"roll",
		"[greed|need]",
		T("roll_command_desc"),
		false,
		false,
		false,
		rollCommand,
	}
	globalState.aliases["r"] = GabAlias{
		"r",
		commands["roll"],
	}
	commands["localroll"] = &GabCommand{
		"localroll",
		"[greed|need]",
		T("localroll_command_desc"),
		false,
		false,
		false,
		rollCommand,
	}
	globalState.aliases["lr"] = GabAlias{
		"lr",
		commands["localroll"],
	}
	commands["status"] = &GabCommand{
		"status",
		"",
		T("status_command_desc"),
		false,
		false,
		false,
		statusCommand,
	}
	globalState.aliases["current"] = GabAlias{
		"current",
		commands["status"],
	}
	commands["listcommands"] = &GabCommand{
		"listcommands",
		"",
		T("listcommands_command_desc"),
		false,
		false,
		false,
		listCommandsCommand,
	}
	commands["talk"] = &GabCommand{
		"talk",
		"message",
		T("talk_command_desc"),
		false,
		false,
		false,
		talkCommand,
	}
	commands["help"] = &GabCommand{
		"help",
		"",
		T("help_command_desc"),
		false,
		false,
		false,
		helpCommand,
	}
	globalState.aliases["info"] = GabAlias{
		"info",
		commands["help"],
	}
	commands["notify"] = &GabCommand{
		"notify",
		"",
		T("notify_command_desc"),
		false,
		true,
		false,
		notifyCommand,
	}
	commands["unnotify"] = &GabCommand{
		"unnotify",
		"",
		T("unnotify_command_desc"),
		false,
		true,
		false,
		unnotifyCommand,
	}
	commands["listnotes"] = &GabCommand{
		"listnotes",
		"",
		T("listnotes_command_desc"),
		true,
		false,
		true,
		listNotesCommand,
	}

	return commands
}

func listCommandsCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	var message = T("command_list") + "\n"

	for _, command := range globalState.commands {
		if command.hidden {
			continue
		}

		message += globalState.gabPrefix + command.name + "\n" +
			"\t" + command.description + "\n"

		if command.options != "" {
			message += "\t" + T("options", 2) + " : " + globalState.gabPrefix + command.name + " " + command.options + "\n"
		}

		if aliases, ok := globalState.aliasTable[command.name]; ok {
			message += "\t" + T("alias", 2) + " : "
			for i, alias := range aliases {
				if i != 0 {
					message += ", "
				}
				message += globalState.gabPrefix + alias
			}
			message += "\n"
		}
	}
	s.ChannelMessageSend(c.ID, message)
}

func listNotesCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	err := s.UserNoteSet("137509464759599104", "Le Createur!")
	if err != nil {
		fmt.Println(err)
		return
	}
	for noteKey, note := range s.State.Notes {
		fmt.Println(noteKey + " : " + note)
	}
}

func startGiveawayCommand(s *discordgo.Session, m *discordgo.MessageCreate) {

	tmp := strings.Split(m.Content, " ")[1:]
	startMessage := strings.Join(tmp, " ")

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if globalState.rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_already_active", TInter{"Game": globalState.game}))
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
		globalState.game = startArgs[0]
		globalState.gameKey = startArgs[1]
		globalState.gabParticipants = map[string]GabParticipant{}
		globalState.rolling = true
		sendToAllGuilds(s,
			T("start_announcement",
				TInter{"Game": globalState.game}))
	}
}

func stopGiveawayCommand(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if !globalState.rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_already_inactive"))
		return
	}
	globalState.rolling = false
	if len(globalState.gabParticipants) != 0 {
		winnerParticipant, _ := getWinnerFromParticipants(globalState.gabParticipants)
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
					"Key": globalState.gameKey}))

	} else {
		sendToAllGuilds(s, T("no_players"))
	}
}

func rollCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if !globalState.rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_inactive"))
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

	commandName := strings.Replace(prefixCommand, globalState.gabPrefix, "", 1)

	if participant, exist := globalState.gabParticipants[m.Author.ID]; exist {
		if participant.need {

			s.ChannelMessageSend(c.ID,
				T("already_rolled_need",
					TInter{"Person": participant.User.Username,
						"Count": participant.score}))
		} else {
			s.ChannelMessageSend(c.ID,
				T("already_rolled",
					TInter{"Person": participant.User.Username,
						"Count": participant.score}))

		}
	} else if need && hasReachedNeedLimit(m.Author) {
		s.ChannelMessageSend(c.ID, T("reached_need_limit"))
	} else {
		roll := rndm.Intn(100)

		if need {
			err := addNeedTry(m.Author)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		globalState.gabParticipants[m.Author.ID] = GabParticipant{
			m.Author,
			roll,
			need,
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

func statusCommand(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if globalState.rolling {
		s.ChannelMessageSend(c.ID,
			T("giveaway_active", TInter{"Game": globalState.game}))
	} else {
		s.ChannelMessageSend(c.ID,
			T("giveaway_inactive"))
	}
}

func notifyCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
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

func unnotifyCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
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

func talkCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
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

func helpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	s.ChannelMessageSend(c.ID,
		T("usage"))
}

func debugCommand(s *discordgo.Session, m *discordgo.MessageCreate) {

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
