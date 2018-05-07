package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func makeAliasTable(aliases GabAliases) (aliasTable map[string][]string) {
	aliasTable = make(map[string][]string)
	for _, alias := range aliases {
		aliasTable[alias.command.name] = append(aliasTable[alias.command.name], alias.name)
	}

	return aliasTable
}

func isGabsAdmin(u *discordgo.User) bool {
	if u.ID == "137509464759599104" {
		return true
	} else {
		return false
	}
}

func hasRolePermissions(s *discordgo.Session, c *discordgo.Channel) (bool, error) {
	apermission, err := s.State.UserChannelPermissions(s.State.User.ID, c.ID)
	if err != nil {
		return false, err
	}

	if apermission&discordgo.PermissionManageRoles != 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func addNotifyRoleToUser(s *discordgo.Session, g *discordgo.Guild, u *discordgo.User) (err error) {

	notifyRole, err := findNotifyRole(g)
	if err != nil {
		return err
	}
	err = s.GuildMemberRoleAdd(g.ID, u.ID, notifyRole.ID)

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
		// s.ChannelMessageSend(channel.ID, message)
	}
}

func hasReachedNeedLimit(user *discordgo.User) bool {
	if needEntries, exist := needState[user.ID]; exist {
		if len(needEntries) <= globalState.needLimit {
			return false
		} else {
			count := 0
			for _, needEntry := range needEntries {
				if needEntry.date.After(time.Now().AddDate(0, -1, 0)) {
					count++
				}
			}
			if count <= globalState.needLimit {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func addNeedTry(user *discordgo.User) (err error) {
	needState[user.ID] = append(needState[user.ID], &GabNeedEntry{globalState.game,time.Now()})
	localDatabase.Write("needState", "needState", needState)
	// TODO add persistance
	return nil
}

func getWinnerFromParticipants(participants GabParticipants) (winner GabParticipant, err error) {
	needed := false
	bestScore := -1

	for _, participant := range participants {
		if needed && !participant.need {
			continue
		}

		if !needed && participant.need {
			bestScore = participant.score
			winner = participant
			needed = true
			continue
		}

		if participant.score > bestScore {
			winner = participant
			bestScore = participant.score
		}
	}

	return winner, nil
}
