package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

func makeAliasTable(aliases Aliases) (aliasTable map[string][]string) {
	aliasTable = make(map[string][]string)
	for _, alias := range aliases {
		aliasTable[alias.Command.Name] = append(aliasTable[alias.Command.Name], alias.Name)
	}

	return aliasTable
}

func makeGuildTable(alliances map[string]*Alliance) (guildTable map[string][]string) {
	guildTable = make(map[string][]string)
	for name, alliance := range alliances {
		for _, guildID := range alliance.Guilds {
			guildTable[guildID] = append(guildTable[guildID], name)
		}
	}

	return guildTable
}

func isGabCreator(user *discordgo.User) bool {
	if user.ID == "137509464759599104" {
		return true
	} else {
		return false
	}
}

func checkPermissionsForCommand(s *State, c *Command, u *discordgo.User) bool {
	if c.NeedsCreator && !isGabCreator(u) {
		return false
	} else if isGabCreator(u) {
		return true
	}

	return false
}

func getAllianceFromMessage(session *discordgo.Session, message *discordgo.MessageCreate) (*Alliance, error) {
	guild, err := getGuildFromMessage(session, message)
	if err != nil {
		return nil, err
	}
	if alliances, exists := globalState.GuildTable[guild.ID]; exists {

		switch len(alliances) {
		case 0:
			return nil, errors.New("damn son, not even a single alliance wants you")
		case 1:
			return globalState.Alliances[alliances[0]], err
		default:
			return nil, errors.New("multiple alliances not yet supported") // TODO multiple alliances
		}
	}
	return nil, errors.New("damn son, not even a single alliance wants you")
}

func getChannelAndArgsFromMessage(session *discordgo.Session, message *discordgo.MessageCreate) (channel *discordgo.Channel, args []string, err error) {
	// Parse arguments from where the message came from.
	args, err = parseArguments(message.Content)
	if err != nil {
		args = nil
	}

	// Find the channel that the message came from.
	channel, err = session.State.Channel(message.ChannelID)
	if err != nil {
		return nil, args, err
	}

	return channel, args, nil
}

func getGuildFromMessage(session *discordgo.Session, message *discordgo.MessageCreate) (guild *discordgo.Guild, err error) {
	// Find the channel that the message came from.
	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		return nil, err
	}

	guild, err = session.State.Guild(channel.GuildID)
	if err != nil {
		return nil, err
	}

	return guild,nil
}


func createAlliance(name string, guild *discordgo.Guild, admin *discordgo.User) (alliance *Alliance) {

	defaultState := &State{
		GabPrefix: defaultPrefix,
		Rolling:   false,
		NeedLimit: defaultNeedLimit,
	}
	defaultState.Commands, defaultState.Aliases = getDefaultGabCommandsAndAliases()
	defaultState.AliasTable = makeAliasTable(defaultState.Aliases)

	alliance = &Alliance{
		Name:   name,
		State:  defaultState,
		Admin:  admin.ID,
		Guilds: append([]string{}, guild.ID),
	}

	alliance.NeedState = make(NeedState)

	return alliance
}

func hasRolePermissions(session *discordgo.Session, channel *discordgo.Channel) (bool, error) {
	apermission, err := session.State.UserChannelPermissions(session.State.User.ID, channel.ID)
	if err != nil {
		return false, err
	}

	if apermission&discordgo.PermissionManageRoles != 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func addNotifyRoleToUser(session *discordgo.Session, guild *discordgo.Guild, user *discordgo.User) (err error) {

	notifyRole, err := findNotifyRole(guild)
	if err != nil {
		return err
	}
	err = session.GuildMemberRoleAdd(guild.ID, user.ID, notifyRole.ID)

	return err
}

func removeNotifyRoleFromUser(s *discordgo.Session, g *discordgo.Guild, u *discordgo.User) (err error) {

	notifyRole, err := findNotifyRole(g)
	if err != nil {
		return err
	}
	err = s.GuildMemberRoleRemove(g.ID, u.ID, notifyRole.ID)
	if err != nil {
		return err
	}
	return nil
}

func findNotifyRole(g *discordgo.Guild) (role *discordgo.Role, err error) {
	var notifyRole *discordgo.Role

	for _, role := range g.Roles {
		if role.Name == "Gab Notifications" {
			notifyRole = role
			fmt.Println(g.Roles)
		}
	}
	if notifyRole == nil {
		return nil, errors.New("no_notify_role")
	} else {
		return notifyRole, nil
	}
}

func getGuildsMainChannel(s *discordgo.Session) []*discordgo.Channel {

	/* // Find the channel that the message came from.
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		// Could not find channel.
		return
	}
	member, err := s.State.Member(g.ID, m.Author.ID)

	for _, role := range member.Roles {
		r, _ := s.State.Role(g.ID, role)
		fmt.Println(m.Author.Username + " " +
			r.Name + " " +
			strconv.Itoa(r.Permissions))

	}*/
	var mainChannels []*discordgo.Channel
	for _, guild := range s.State.Guilds {
		// fmt.Println(guild.Name)

		for _, channel := range guild.Channels {
			if channel.Type == 0 && channel.Position == 0 {
				// fmt.Printf("\t- %s\n", channel.Name)
				mainChannels = append(mainChannels, channel)
			}
		}
	}

	return mainChannels
}

func getGuildsGiveawayChannel(s *discordgo.Session) []*discordgo.Channel {
	var mainChannels []*discordgo.Channel
	for _, guild := range s.State.Guilds {
		// fmt.Println(guild.Name)

		for _, channel := range guild.Channels {
			if channel.Type == 0 && channel.Name == "giveaway" {
				// fmt.Printf("\t- %s\n", channel.Name)
				mainChannels = append(mainChannels, channel)
			}
		}
	}

	return mainChannels
}

func sendToAllGuilds(s *discordgo.Session, message string) {
	channels := getGuildsGiveawayChannel(s)
	sendToChannels(s, channels, message)
}

func sendToAllGuildsMainChannel(s *discordgo.Session, message string) {
	channels := getGuildsMainChannel(s)
	sendToChannels(s, channels, message)
}

func sendToChannels(s *discordgo.Session, channels []*discordgo.Channel, message string) {
	fmt.Println("Message : " + message)
	for _, channel := range channels {
		guild, _ := s.Guild(channel.GuildID)
		fmt.Printf("to %s : %s\n", guild.Name, channel.Name)
		s.ChannelMessageSend(channel.ID, message)
	}
}

func hasReachedNeedLimit(user *discordgo.User, alliance *Alliance) bool {
	if needEntries, exist := alliance.NeedState[user.ID]; exist {
		if len(needEntries) < alliance.State.NeedLimit {
			return false
		} else {
			count := 0
			for _, needEntry := range needEntries {
				if needEntry.Date.After(time.Now().AddDate(0, -1, 0)) {
					count++
				}
			}
			if count < alliance.State.NeedLimit {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func addNeedTry(user *discordgo.User, state *State) (err error) {
	needState[user.ID] = append(needState[user.ID],
		NeedEntry{Game: state.Game, Date: time.Now()})

	err = persistGlobalData(globalState)
	if err != nil {
		needState[user.ID] = needState[user.ID][:len(needState[user.ID])-1]
		return err
	}
	return nil
}

func getWinnerFromParticipants(participants Participants) (winner *Participant, err error) {
	needed := false
	bestScore := -1

	for _, participant := range participants {
		if needed && !participant.Need {
			continue
		}

		if !needed && participant.Need {
			bestScore = participant.Score
			winner = participant
			needed = true
			continue
		}

		if participant.Score > bestScore {
			winner = participant
			bestScore = participant.Score
		}
	}

	return winner, nil
}

func persistGlobalData(state GlobalState) (err error) {
	reader, err := os.OpenFile(globalState.DataDirectory+"/"+globalState.GlobalStateFile, os.O_WRONLY, 0600)
	defer reader.Close()
	if err != nil {
		fmt.Println("data file write opening error:", err)
		return err
	}
	err = gob.NewEncoder(reader).Encode(state)
	if err != nil {
		return err
	}
	return nil
}
