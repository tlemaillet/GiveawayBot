package main

import (
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

type GabCommand struct {
	name        string
	options     string
	description string
	needsAdmin  bool
	needsServer bool
	hidden      bool
	callback    func(session *discordgo.Session, message *discordgo.MessageCreate)
}
type GabCommands map[string]*GabCommand

type GabAlias struct {
	name    string
	command *GabCommand
}
type GabAliases map[string]GabAlias

type GabParticipant struct {
	User  *discordgo.User
	score int
	need  bool
}
type GabParticipants map[string]GabParticipant

type GabState struct {
	gabPrefix string

	commands   GabCommands
	aliases    GabAliases
	aliasTable map[string][]string

	game    string
	gameKey string

	rolling         bool
	gabParticipants GabParticipants
	needLimit       int
}
type GabGuildState struct {
	state GabState
	guild *discordgo.Guild
}

type GabNeedEntry struct {
	game   string
	date   time.Time
}
type GabNeedState map[string][]*GabNeedEntry

const defaultPrefix = "!gab"
const defaultNeedLimit = 2

var globalState GabState
var guildsState map[string]GabGuildState
var needState GabNeedState

var token string

func init() {
	var translationFile string

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&translationFile, "T", "en_US.all.json", "Translation")
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

	globalState = GabState{
		gabPrefix: defaultPrefix,

		commands:   make(GabCommands),
		aliases:    make(GabAliases),
		aliasTable: nil,

		game:    "",
		gameKey: "",

		rolling:         false,
		gabParticipants: nil,
		needLimit:       defaultNeedLimit,
	}
	_ = guildsState // TODO guildsState = make(map[string]GabGuildState)

	// TODO add persistance
	needState = make(GabNeedState)

	globalState.commands = getDefaultGabCommandsAndAliases()

	globalState.aliasTable = makeAliasTable(globalState.aliases)
}

func main() {
	dg, err := discordgo.New("Bot " + token)
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

	// check if the message starts with defined gabPrefix
	if !strings.HasPrefix(m.Content, globalState.gabPrefix) {
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
	commandName := strings.Replace(prefixCommand, globalState.gabPrefix, "", 1)

	var command *GabCommand = nil

	if validAlias, ok := globalState.aliases[commandName]; ok {
		command = validAlias.command
	} else if validCommand, ok := globalState.commands[commandName]; ok {
		command = validCommand
	}

	if command != nil {
		if command.needsAdmin && !isGabsAdmin(m.Author) {
			return
		}

		if command.needsServer {
			_, err := s.Guild(c.GuildID)
			if err != nil {
				s.ChannelMessageSend(c.ID, T("serv_command_only"))
				return
			}
		}

		command.callback(s, m)
		return
	} else {
		s.ChannelMessageSend(c.ID, T("shout_gab"))
	}
}
