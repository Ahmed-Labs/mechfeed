package bot

import (
	"errors"
	"fmt"
	"mechfeed/users"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	BotSession.active = true
	BotSession.dg = dg

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

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

var commands = map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	"!help": handleHelp,
	"!start": handleOnboard,
	// "!info": handleInfo,
}

var protected_commands = map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	"!add": handleAdd,
	"!list": handleList,
	"!delete": handleDelete,
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "" || m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	input := strings.Split(m.Content, " ")
	cmd := input[0]
	args := input[1:]

	if handler, ok := commands[cmd]; ok {
		fmt.Println(m.Author.Username, ":", cmd, " command received.")
		err := handler(s, m, args)
		if err != nil {
			SendTextDM(s, m.Author.ID, err.Error())
		}
	} else if handler, ok := protected_commands[cmd]; ok {
		repo, _ := users.DBConnection()
		exists, _ := repo.Queries.GetUserExistence(repo.Ctx, m.Author.ID)
		if exists == 0 {
			SendTextDM(s, m.Author.ID, "Please use the `!start` command before using alert features.")
			return
		}
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

func SendEmbedDM(s *discordgo.Session, userID string, embed *discordgo.MessageEmbed) {
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		fmt.Println("Error creating channel:", err)
		return
	}
	_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
	if err != nil {
		fmt.Println("Error sending DM embeds:", err)
	}
}
// External use
func IsolatedSendEmbedDM(user_id string, embed *discordgo.MessageEmbed) {
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

// Handlers

func handleHelp(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	SendEmbedDM(s, m.Author.ID, MechfeedHelpEmbed)
	return nil
}

// func handleInfo(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
// 	SendTextDM(s, m.Author.ID, "Info about Mechfeed: ...")
// 	return nil
// }

func handleOnboard(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	// Send the onboarding embed
	embeds := []*discordgo.MessageEmbed{
		MechfeedLogoEmbed,
		MechfeedIntroEmbed,
	}
	SendMultipleEmbedsDM(s, m.Author.ID, embeds)

	repo, err := users.DBConnection()
	if err != nil {
		fmt.Println("failed to get DB connection.")
		return nil
	}
	repo.Queries.CreateUser(repo.Ctx, users.CreateUserParams{
		ID: m.Author.ID,
		Username: m.Author.Username,
	})
	return nil
}
 
func handleList(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
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
		var sb strings.Builder
		for i, alert := range alerts {
			sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, alert.Keyword))
		}
		SendTextDM(s, m.Author.ID, "```" + sb.String() + "```")
	}
	
	return nil
}

func handleAdd(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	repo, err := users.DBConnection()
	if err != nil {
		fmt.Println("failed to get DB connection.")
		return errors.New("failed to add alerts, please contact dev or try again later")
	}
	if len(args) == 0 {
		fmt.Println("no alerts provided.")
		return errors.New("no alerts provided")
	}

	failure := false
	for _, arg := range args {
		err := repo.Queries.CreateAlert(repo.Ctx, users.CreateAlertParams{
			ID: m.Author.ID,
			Keyword: arg,
		})
		if err != nil {
			failure = true
		}
	}	
	if failure {
		fmt.Println("failed to store all alerts")
		return errors.New("failed to add alerts, please contact dev or try again later")
	} else {
		var msg string
		if len(args) == 1 {
			msg = "Successfully added alert!"
		} else {
			msg = "Successfully added alerts!"
		}
 		SendTextDM(s, m.Author.ID, msg)
	}

	return nil
}


func handleDelete(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	repo, err := users.DBConnection()
	if err != nil {
		fmt.Println("failed to get DB connection.")
		return errors.New("failed to delete alerts, please contact dev or try again later")
	}
	if len(args) == 0 {
		fmt.Println("no input provided.")
		return errors.New("no input provided")
	}

	alerts, err := repo.Queries.GetUserAlerts(repo.Ctx, m.Author.ID)
	if err != nil {
		fmt.Println("failed to query DB for alerts before deletion")
		return errors.New("failed to delete alerts, please contact dev or try again later")
	}

	if len(alerts) == 0 {
		SendTextDM(s, m.Author.ID, "You have 0 alerts added. Add some with the `!add` command!")
		return nil
	}

	if args[0] == "all" {
		err := repo.Queries.DeleteAllAlerts(repo.Ctx, m.Author.ID)

		if err != nil {
			SendTextDM(s, m.Author.ID, "Failed to delete all alerts")
		} else {
			SendTextDM(s, m.Author.ID, "Successfully deleted all alerts!")
		}

		return nil
	}

	
	var alert_id_map = map[string]int32{}
	for i, alert := range alerts {
		alert_id_map[fmt.Sprintf("%d", i+1)] = alert.AlertID
	}

	deleted := 0
	delete_count := len(args)

	for _, arg := range args {
		alert_id, ok := alert_id_map[arg]
		if !ok {
			continue
		}
		err := repo.Queries.DeleteAlert(repo.Ctx, alert_id)
		if err != nil {
			continue
		}
		deleted++
	}	
	if deleted != delete_count {
		SendTextDM(s, m.Author.ID, fmt.Sprintf("Deleted %d/%d provided alerts", deleted, delete_count))
	} else {
		var msg string
		if delete_count == 1 {
			msg = "Successfully deleted alert!"
		} else {
			msg = fmt.Sprintf("Successfully deleted %d alerts!", deleted)
		}
 		SendTextDM(s, m.Author.ID, msg)
	}

	return nil
}