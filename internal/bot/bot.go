package bot

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aarnphm/iris/internal/configs"
	"github.com/aarnphm/iris/internal/db"
	"github.com/aarnphm/iris/internal/log"
	"github.com/bwmarrin/discordgo"
)

const (
	discordBotPrefix    = "Bot "
	baseAuthURLTemplate = "https://discordapp.com/api/oauth2/authorize?client_id=%s&scope=bot"
)

var pomDuration = time.Minute * 25

type cmdHandler func(s *discordgo.Session, m *discordgo.MessageCreate, ex string)

type botCommand struct {
	handler       cmdHandler
	desc          string
	exampleParams string
}

type Iris struct {
	Config      configs.Config
	secrets     configs.Secrets
	cmdHandlers map[string]botCommand
	discord     *discordgo.Session
	logger      log.Logger

	helpMessage   string
	inviteMessage string
	poms          db.UserPomodoroMap

	// record metrics here
	// metrics metrics.Recorder
}

func NewIris(config configs.Config, secrets configs.Secrets, logger log.Logger) *Iris {
	ir := &Iris{
		Config:  config,
		secrets: secrets,
		logger:  logger.Name("iris"),
		poms:    db.NewUserPomodoroMap(),
	}

	ir.registerCmdHandlers()
	ir.inviteMessage = fmt.Sprintf("Click here: <"+baseAuthURLTemplate+"> to invite me to the server", ir.secrets.ClientID)
	ir.helpMessage = ir.buildHelpMessage()
	return ir
}

func (ir *Iris) registerCmdHandlers() {
	ir.cmdHandlers = map[string]botCommand{
		"help":   {handler: ir.onCmdHelp, desc: "Show this help message", exampleParams: ""},
		"pom":    {handler: ir.onCmdStartPom, desc: "Start a pom work cycle. You can optionally specify the task you are working and and the period of time (optional)", exampleParams: "`for studying` 50"},
		"cancel": {handler: ir.onCmdCancelPom, desc: "cancle current pom cycle", exampleParams: ""},
		"invite": {handler: ir.onCmdInvite, desc: "Create an invite link you can use to have the bot join the server", exampleParams: ""},
		// "simp":   {handler: ir.onCmdSimp, desc: "notify another friend with the good stuff", exampleParams: ""},
	}
}

func (ir *Iris) buildHelpMessage() string {
	helpBuffer := bytes.Buffer{}
	helpBuffer.WriteString("Made by aarnphm, for my lord iris")

	// just use map iteration order
	for cmdStr, cmd := range ir.cmdHandlers {
		helpBuffer.WriteString(fmt.Sprintf("\n•  **%s**  •  %s\n", cmdStr, cmd.desc))
		helpBuffer.WriteString(fmt.Sprintf("- Example: `%s%s %s`\n", ir.Config.CmdPrefix, cmdStr, cmd.exampleParams))
	}

	helpBuffer.WriteString("\n" + ir.inviteMessage)

	return helpBuffer.String()
}

// Start will start the bot, blocking til completed
func (ir *Iris) Start() error {
	if ir.secrets.AuthToken == "" {
		return errors.New("No authToken found")
	}

	var err error
	ir.discord, err = discordgo.New(discordBotPrefix + ir.secrets.AuthToken)
	if err != nil {
		return err
	}

	// ir.discord.AddHandler(ir.onReady)
	ir.discord.AddHandler(ir.onMessageReceived)
	// keep track of how many guild we join
	// ir.discord.AddHandler(ir.onGuildCreate)
	// ir.discord.AddHandler(ir.onGuildDelete)

	err = ir.discord.Open()
	if err != nil {
		return err
	}

	ir.logger.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return ir.discord.Close()
}

// onMessageReceived will be called everytime a new message is created on any channel that the bot is listenning to
// It will dispatch know commands to command handlers, passing along necessary info
func (ir *Iris) onMessageReceived(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore message created by the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	msg := m.Content

	cmdPrefixLen := len(ir.Config.CmdPrefix)

	// dispatch the command iff we have our prefix, (case-insensitive) otherwise throws an errors
	if len(msg) > cmdPrefixLen && strings.EqualFold(ir.Config.CmdPrefix, msg[0:cmdPrefixLen]) {
		afterPrefix := msg[cmdPrefixLen:]
		cmd := strings.SplitN(afterPrefix, "", 2)

		if f, ok := ir.cmdHandlers[strings.ToLower(cmd[0])]; ok {
			rest := ""
			if len(cmd) > 1 {
				rest = cmd[1]
			}

			if f.handler != nil {
				f.handler(s, m, rest)
			} else {
				ir.logger.Fatal(errors.New(fmt.Sprintf("nil handlers for command %#v", cmd)))
				s.ChannelMessageSend(m.ChannelID, "Command error - dm @aarnphm")
			}
		}
	}
}

func (ir *Iris) onCmdHelp(s *discordgo.Session, m *discordgo.MessageCreate, ex string) {
	s.ChannelMessageSend(m.ChannelID, ir.helpMessage)
}

func (ir *Iris) onCmdInvite(s *discordgo.Session, m *discordgo.MessageCreate, ex string) {
	s.ChannelMessageSend(m.ChannelID, ir.inviteMessage)
}

func (ir *Iris) onCmdStartPom(s *discordgo.Session, m *discordgo.MessageCreate, ex string) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ir.logger.Fatal(err)
	}

	// make sure the users' text can't break out of our quote box
	// ex[0] will be message ex[1] will be time period (optional)
	ex = strings.Replace(ex, "`", "", -1)
	exa := strings.SplitN(ex, " ", 2)

	title := exa[0]
	if exa[1] != "" {
		newDuration, _ := strconv.Atoi(exa[1])
		pomDuration = time.Minute * time.Duration(newDuration)
	}

	notif := db.NotifyInfo{
		TitleID: title,
		User: &db.User{
			DiscordID:  m.Author.ID,
			DiscordTag: m.Author.Discriminator,
			GuidID:     channel.GuildID,
			ChannelID:  m.ChannelID,
		},
	}

	if ir.poms.CreateIfEmpty(pomDuration, ir.onPomEnded, notif) {
		taskStr := "Started task - "
		if len(notif.TitleID) > 0 {
			taskStr = fmt.Sprintf("```md\n%s\n```", notif.TitleID)
		}

		msg := fmt.Sprintf("%s**%.1f minutes** remaining!", taskStr, pomDuration.Minutes())
		s.ChannelMessageSend(m.ChannelID, msg)
		//metrics here
		// ir.metrics.RecordStartPom()
		// ir.metrics.RecordRunningPoms(int64(ir.poms.Count()))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("A pomodoro is already running for %s", m.Author.ID))
	}

}

func (ir *Iris) onCmdCancelPom(s *discordgo.Session, m *discordgo.MessageCreate, ex string) {
	if exists := ir.poms.RemoveIfExists(m.Author.ID); !exists {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No pom is currently running for %s", m.Author.ID))
	}
	// if this removal is success then call onPomEnded
}

func (ir *Iris) onPomEnded(notif db.NotifyInfo, completed bool) {
	var err error
	var hash string
	if completed {
		// update users' progress to databse
		if err = db.FetchUser(notif.User.DiscordID); err != nil {
			// create new users entry
			hash, err = db.NewUser(notif.User.DiscordID, notif.User.DiscordTag, notif.User.GuidID, pomDuration.String())
			ir.logger.Info("inserted %s to database. Hash: %s", notif.User.DiscordID, hash)
			if err != nil {
				ir.logger.Fatal(err)
			}
		} else {
			// users already in database, just updates timing
			addedPom, _ := strconv.Atoi(pomDuration.String())
			if err = db.UpdateUser(notif.User.DiscordID, addedPom); err != nil {
				ir.logger.Fatal(err)
			}
		}

		message := "Work cycle complete. Time to take a break!"
		var toMention []string
		if len(notif.TitleID) > 0 {
			message = fmt.Sprintf("```md\n%s\n```%s", notif.TitleID, message)
		}

		user, err := ir.discord.User(notif.User.DiscordID)
		if err == nil {
			toMention = append(toMention, user.Mention())
		}

		if len(toMention) > 0 {
			mentions := strings.Join(toMention, " ")
			message = fmt.Sprintf("%s\n%s", message, mentions)
		}

		ir.discord.ChannelMessageSend(notif.User.ChannelID, message)
	} else {
		ir.discord.ChannelMessageSend(notif.User.ChannelID, "Pom cancelled!")
	}

	// ir.metrics.RecordRunningPoms(int64(ir.poms.Count()))

}

// onGuildCreate is called when a Guild adds the bot
// func (ir *Iris) onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
// 	ir.metrics.RecordConnectedServers(int64(len(s.State.Guilds)))
// }

// onGuildDelete is called when a Guild removes the bot.
// func (ir *Iris) onGuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
// 	ir.metrics.RecordConnectedServers(int64(len(s.State.Guilds)))
// }
