package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func listCommandsCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	var message = T("command_list") + "\n"

	for _, command := range commands {
		if command.hidden {
			continue
		}

		message += defaultPrefix + command.name + "\n" +
			"\t" + command.description + "\n"
		if aliases, ok := aliasTable[command.name]; ok {
			message += "\tAliases : "
			for i, alias := range aliases {
				if i != 0 {
					message += ", "
				}
				message += defaultPrefix + alias
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

	if rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_already_active", TInter{"Game": game}))
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
		game = startArgs[0]
		gameKey = startArgs[1]
		participants = map[string]int{}
		rolling = true
		sendToAllGuilds(s,
			T("start_announcement",
				TInter{"Game": game}))
	}
}

func stopGiveawayCommand(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if !rolling {
		s.ChannelMessageSend(c.ID, T("giveaway_already_inactive"))
		return
	}
	rolling = false
	if len(participants) != 0 {
		var winK string
		winV := -1

		for k, v := range participants {
			if v > winV {
				winK = k
				winV = v
			}
		}

		winner, _ := s.User(winK)

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
					"Key": gameKey}))

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

	prefix := strings.Split(m.Content, " ")[0]

	if rolling {
		if roll, exist := participants[m.Author.ID]; exist {
			s.ChannelMessageSend(c.ID,
				T("already_rolled",
					TInter{"Person": m.Author.Username, "Count": roll}))
		} else {
			roll := rndm.Intn(100)
			participants[m.Author.ID] = roll

			switch prefix {
			case "!gabroll", "!gabr":
				sendToAllGuilds(s,
					T("roll_result",
						TInter{"Person": m.Author.Username, "Count": roll}))
			case "!gablocalroll", "!gablr":
				s.ChannelMessageSend(c.ID,
					T("roll_result",
						TInter{"Person": m.Author.Username, "Count": roll}))
			}
		}
	} else {
		s.ChannelMessageSend(c.ID,
			T("giveaway_inactive"))
	}
}

func statusCommand(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	if rolling {
		s.ChannelMessageSend(c.ID,
			T("giveaway_active", TInter{"Game": game}))
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
