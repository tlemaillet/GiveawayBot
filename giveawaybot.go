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
	options		string
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
	User    *discordgo.User
	score	int
	need 	bool
}
type GabParticipants map[string]GabParticipant

type GabState struct {
	gabPrefix string

	commands GabCommands
	aliases GabAliases
	aliasTable map[string][]string

	game string
	gameKey string

	rolling bool
	gabParticipants GabParticipants
	needLimit int
}
type GabGuildState struct {
	state GabState
	guild *discordgo.Guild
}


const defaultPrefix = "!gab"
const defaultNeedLimit = 2

var globalState GabState
var guildsState map[string]GabGuildState

var token string

var gabPrefix string

var commands GabCommands
var aliases GabAliases
var aliasTable map[string][]string

var game string
var gameKey string

var rolling bool
var gabParticipants GabParticipants
var needLimit int


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
		return
	}
	_ = globalState // TODO Il faudra l'initialiser ici
	_ = guildsState // TODO guildsState = make(map[string]GabGuildState)

	gabPrefix = defaultPrefix

	commands = make(map[string]*GabCommand)
	aliases = make(map[string]GabAlias)

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
	aliases["r"] = GabAlias{
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
	aliases["lr"] = GabAlias{
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
	aliases["current"] = GabAlias{
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
	aliases["info"] = GabAlias{
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

	aliasTable = makeAliasTable(aliases)

	needLimit = defaultNeedLimit
}

func makeAliasTable(aliases GabAliases) (aliasTable map[string][]string) {
	aliasTable = make(map[string][]string)
	for _, alias := range aliases {
		aliasTable[alias.command.name] = append(aliasTable[alias.command.name], alias.name)
	}

	return aliasTable
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
	if !strings.HasPrefix(m.Content, gabPrefix) {
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
	commandName := strings.Replace(prefixCommand, gabPrefix, "", 1)

	var command *GabCommand = nil

	if validAlias, ok := aliases[commandName]; ok {
		command = validAlias.command
	} else if validCommand, ok := commands[commandName]; ok {
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
