package bot

import (
	"github.com/bwmarrin/discordgo"
)


var MechfeedLogoEmbed *discordgo.MessageEmbed = &discordgo.MessageEmbed{
	Color: 0xe671dc,
	Image: &discordgo.MessageEmbedImage{
		URL: "https://cdn.discordapp.com/attachments/1218992423313739847/1272775005108961362/mechfeed.jpg?ex=66bc3398&is=66bae218&hm=746cdf7f7d3cf710e2e3adb42efd172cc0bfdba33cb58fd04f9c7ee456ec7a58&",
	},
}

var MechfeedIntroEmbed *discordgo.MessageEmbed = &discordgo.MessageEmbed{
	Color: 0xe671dc,
	Title: "Welcome to Mechfeed!",
	Description: "Mechfeed is your go-to bot for managing alerts and staying updated. Here's how to get started:",
	Fields: []*discordgo.MessageEmbedField{
		{
			Name:   "Setting Alerts",
			Value:  "Use `!set alerts <parameters>` to configure your alerts. You can customize them based on your preferences.",
			Inline: false,
		},
		{
			Name:   "Viewing Alerts",
			Value:  "Use `!view alerts` to see a list of your current alerts.",
			Inline: false,
		},
		{
			Name:   "Help",
			Value:  "Need more help? Use `!help` to see a list of all available commands.",
			Inline: false,
		},
	},
}
