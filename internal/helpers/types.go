// Package helpers contains all static and support variables for iris.
package helpers

import (
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	// ----------------------- build related -ldflags.

	// AppVersion is version of iridaceae.
	AppVersion = "TEST_BUILD"
	// AppCommit tracks iridaceae commits.
	AppCommit = "TEST_BUILD"
	// Release defines whether our app is ready for production.
	Release = "FALSE"
	// Repo is where the source code live.
	Repo = "https://github.com/Iridaceae/iridaceae"

	// ----------------------- statistic/metrics.

	// StatsStartupTime tracks our startup time.
	StatsStartupTime = time.Now()
	// StatsCommandsExecuted tracks # of commands executed.
	StatsCommandsExecuted = 0
	// StatsMessageAnalyzed will help us log bot command handling.
	StatsMessageAnalyzed = 0

	// ----------------------- errors definition.

	// ErrSessionNotDefined is thrown when discordgo.Session is nil.
	ErrSessionNotDefined = errors.New("session not defined")
	// ErrEmbedNotDefined is thrown when discordgo.MessageEmbed is not defined.
	ErrEmbedNotDefined = errors.New("embed not defined")

	// ----------------------- karma level.

	// PermLvlBotOwner is my level =).
	PermLvlBotOwner = 69420
	// PermLvlGuildOwner is guild owner level.
	PermLvlGuildOwner = 100

	// DefaultAdminRules defines admins rules for each guild.
	DefaultAdminRules = []string{
		"+rs.guild.*",
		"+rs.etc.*",
		"+rs.chat.*",
	}
	// DefaultUserRules defines users rules for each guild.
	DefaultUserRules = []string{
		"+rs.etc.*",
		"+rs.chat.*",
	}
	// AdditionPermission allows users to configure karma and unban.
	AdditionPermission = []string{
		"rs.guild.configparser.karma",
		"rs.guild.configparser.unbanrequest",
	}
)

const (

	// ----------------------- invitation permission.

	// InvitePermission defines Iridaceae perms.
	// Refers to Discord API docs.
	InvitePermission = 0x1 | // instant invite
		0x10 | // Manage channel
		0x20 | // manage guild
		0x40 | // add reaction
		0x400 | // view channel
		0x800 | // send messages
		0x2000 | // manage messages
		0x4000 | // embed links
		0x8000 | // attach files
		0x10000 | // read message history
		0x20000 | // mentions @everyone
		0x40000 | // use external emojis
		0x4000000 | // change nickname
		0x8000000 | // manage nickname
		0x10000000 | // manage roles
		0x20000000 | // manage webhooks
		0x40000000 // manage emoji

	// ----------------------- intent settings.

	// Intents defines Iridaceae event system from discordgo.
	Intents = discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildBans |
		discordgo.IntentsGuildEmojis |
		discordgo.IntentsGuildIntegrations |
		discordgo.IntentsGuildInvites |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildPresences |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuilds

	// ----------------------- colors.

	EmbedColorDefault  = 0x6A5ACD
	EmbedColorError    = 0xE53935
	EmbedColorUpdated  = 0x8BC34A
	EmbedColorGray     = 0xB0BEC5
	EmbedColorOrange   = 0xFB8C00
	EmbedColorGreen    = 0x8BC34A
	EmbedColorCyan     = 0x00BCD4
	EmbedColorYellow   = 0xFFC107
	EmbedColorViolet   = 0x6A1B9A
	ReportRevokedColor = 0x9C27B0
)
