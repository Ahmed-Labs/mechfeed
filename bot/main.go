package bot

import (
	"errors"
	"fmt"
	"mechfeed/users"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)


var DISCORD_BOT_TOKEN string

var BotSession struct {
	active bool
	dg *discordgo.Session
}

func init_bot() error {
	DISCORD_BOT_TOKEN = os.Getenv("DISCORD_BOT_TOKEN")
	if DISCORD_BOT_TOKEN == "" {
		return errors.New("no discord bot token found")
	}
	
	return nil
}

func MechfeedBot() {
	err := init_bot()
	if err != nil {
		fmt.Println("error starting mechfeed bot,", err)
		return
	}
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + DISCORD_BOT_TOKEN)
	if err != nil {
		fmt.Println("error creating Discord session,", err.Error())
		return
	}
	defer dg.Close()

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages |
						  discordgo.IntentsDirectMessages |
						  discordgo.IntentsDirectMessageReactions

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	BotSession.active = true
	BotSession.dg = dg

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	// sc := make(chan os.Signal, 1)
	// signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	// <-sc
}

var Commands = map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	// Core
	"!help": handleHelp,
	"!info": handleInfo,
	"!start": handleOnboard,
	// Alerts
	"!add": handleAdd,
	"!view": handleView,
	// "!delete": handleDelete,
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "" || m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	input := strings.Split(m.Content, " ")
	cmd := input[0]
	args := input[1:]

	if handler, ok := Commands[cmd]; ok {
		fmt.Println(m.Author.Username, ":", cmd, " command received.")
		err := handler(s, m, args)
		if err != nil {
			SendTextDM(s, m.Author.ID, err.Error())
		}
	} else {
		fmt.Println("Invalid command:", cmd)
	}
}

func SendTextDM(s *discordgo.Session, userID, content string) {
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		fmt.Println("Error creating channel:", err)
		return
	}
	_, err = s.ChannelMessageSend(channel.ID, content)
	if err != nil {
		fmt.Println("Error sending DM message:", err)
	}
}

// External use
func SendEmbedDM(user_id string, embed *discordgo.MessageEmbed) {
	if !BotSession.active {
		fmt.Println("bot session inactive.")
		return
	}
	channel, err := BotSession.dg.UserChannelCreate(user_id)
	if err != nil {
		// can happen if no mutual servers
		fmt.Println("error creating channel: ", err)
		return
	}

	_, err = BotSession.dg.ChannelMessageSendEmbed(channel.ID, embed)
	if err != nil {
		// dont share server / disabled DM in settings
		fmt.Println("error sending DM message:", err)
	}
}

func SendMultipleEmbedsDM(s *discordgo.Session, userID string, embeds []*discordgo.MessageEmbed) {
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		fmt.Println("Error creating channel:", err)
		return
	}
	_, err = s.ChannelMessageSendEmbeds(channel.ID, embeds)
	if err != nil {
		fmt.Println("Error sending DM embeds:", err)
	}
}

// Handlers

func handleHelp(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	SendTextDM(s, m.Author.ID, "Here are the available commands: ...")
	return nil
}

func handleInfo(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	SendTextDM(s, m.Author.ID, "Info about Mechfeed: ...")
	return nil
}

func handleOnboard(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	// Send the onboarding embed
	embeds := []*discordgo.MessageEmbed{
		MechfeedLogoEmbed,
		MechfeedIntroEmbed,
	}
	SendMultipleEmbedsDM(s, m.Author.ID, embeds)
	return nil
}
 
func handleView(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	repo, err := users.DBConnection()
	if err != nil {
		fmt.Println("failed to get DB connection.")
		SendTextDM(s, m.Author.ID, "Failed to get alerts. Please contact dev or try again later!")
		return nil
	}

	alerts, err := repo.Queries.GetUserAlerts(repo.Ctx, m.Author.ID)
	if err != nil {
		fmt.Println("failed to query DB for alerts")
		SendTextDM(s, m.Author.ID, "Failed to get alerts. Please contact dev or try again later!")
		return nil
	}

	fmt.Println(m.Author.Username, "Alerts: ", alerts)

	if len(alerts) == 0 {
		SendTextDM(s, m.Author.ID, "No alerts found.")
	} else {
		for c := 0; c < 5; c++ {
			alerts = append(alerts, alerts...)
		}
		var sb strings.Builder
		for i, alert := range alerts {
			sb.WriteString(fmt.Sprintf("%2d. %s\n", i+1, alert.Keyword))
		}
		SendTextDM(s, m.Author.ID, "```" + sb.String() + "```")
	}
	
	return nil
}

func handleAdd(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	return nil
}
