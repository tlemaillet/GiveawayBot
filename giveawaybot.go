package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/i18n"
)

var rndm = rand.New(rand.NewSource(time.Now().UnixNano()))
var T i18n.TranslateFunc

type TInter map[string]interface{}

const defaultPrefix = "!gab"
const defaultNeedLimit = 2

var globalState GlobalState

var token string

func init() {
	var translationFile string
	var dataDirectory string
	var globalStateFile string

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&translationFile, "T", "en_US.all.json", "Translation")
	flag.StringVar(&dataDirectory, "D", "./gabData", "dataDirectory")
	flag.StringVar(&globalStateFile, "g", "globalState.gob", "Global State file")
	flag.Parse()

	if token == "" {
		log.Fatal("No token")
		os.Exit(1)
	}

	translationName := strings.Split(path.Base(translationFile), ".")[0]

	i18n.MustLoadTranslationFile(translationFile)
	var err error
	T, err = i18n.Tfunc(translationName)
	if err != nil {
		fmt.Errorf("error creating translation function:\n %s\n", err)
		os.Exit(1)
	}

	gsReader, err := os.OpenFile(dataDirectory + "/" + globalStateFile, os.O_RDONLY, 0600)
	if perr, ok := err.(*os.PathError); ok {
		fmt.Println(perr)
		initGlobalState(dataDirectory, globalStateFile)
	} else if err != nil {
		log.Fatal("data file opening error", err)
		os.Exit(1)
	} else {

		dec := gob.NewDecoder(gsReader)

		var pState PersitentGlobalState
		err = dec.Decode(&pState)
		if err != nil {
			fmt.Println("decode error:", err)
			initGlobalState(dataDirectory, globalStateFile)
		} else {
			fmt.Println(pState)
			fmt.Println(pState.Alliances["Toto"].State.Participants)
			fmt.Println(pState.Alliances["Toto"].State.NeedState)
			reinitGlobalState(pState, dataDirectory, globalStateFile)
		}
	}
	gsReader.Close()
}

func initGlobalState(dataDirectory string, globalStateFile string){
	globalState = GlobalState{
		Alliances:       make(map[string]*Alliance),
		GuildTable:      make(map[string][]string),
		DataDirectory:   dataDirectory,
		GlobalStateFile: globalStateFile,
		BotToken:        token,
	}
	initCommandList()
}

func reinitGlobalState(state PersitentGlobalState, dataDirectory string, globalStateFile string){
	globalState = GlobalState{
		Alliances:       state.Alliances,
		GuildTable:      makeGuildTable(state.Alliances),
		DataDirectory:   dataDirectory,
		GlobalStateFile: globalStateFile,
		BotToken:        token,
	}

	initCommandList()
}

func main() {
	dg, err := discordgo.New("Bot " + globalState.BotToken)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	// Cleanly close down the Discord session on return.
	defer persistGlobalData(globalState)
	defer dg.Close()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("GiveawayBot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	s.UpdateStatus(0, "!gabhelp")
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	/*if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(
				channel.ID,
				T("bot_ready"))
			return
		}
	}*/
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by other bots or the bot itself
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	var state *State
	// Find the alliance that the message came from
	alliance, err := getAllianceFromMessage(s, m)
	if err != nil {
		// No alliance found for message, defaulting to default prefix and disabling most commands
		state = &State{
			GabPrefix: defaultPrefix,
		}
		state.Commands, state.Aliases = getfallbackCommandsAndAliases()
	} else {
		state = alliance.State
	}

	// check if the message starts with defined gabPrefix
	if !strings.HasPrefix(m.Content, state.GabPrefix) {
		return
	}

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	prefixCommand := strings.Split(m.Content, " ")[0]
	fmt.Printf("%s : %s\n", m.Author.Username, prefixCommand)
	commandName := strings.Replace(prefixCommand, state.GabPrefix, "", 1)

	var command *Command = nil
	if validAlias, ok := state.Aliases[commandName]; ok {
		alias, _ := getAliasFromName(validAlias)
		command = alias.Command
	} else if validCommand, ok := state.Commands[commandName]; ok {
		command, err = getCommandFromName(validCommand)
	}

	if command != nil {
		if command.NeedsCreator && !isGabCreator(m.Author) {
			return
		}

		if command.NeedsGuild {
			_, err := s.Guild(c.GuildID)
			if err != nil {
				s.ChannelMessageSend(c.ID, T("serv_command_only"))
				return
			}
		}

		if command.Callback == nil {
			s.ChannelMessageSend(c.ID, T("wtf"))
			s.ChannelMessageSend(c.ID, command.Name + "/n")
			return
		}

		command.Callback(s, m, alliance)
	} else {
		s.ChannelMessageSend(c.ID, T("unknown_command"))
	}
	return
}
