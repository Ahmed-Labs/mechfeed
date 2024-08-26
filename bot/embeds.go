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

var HelpInformation = []*discordgo.MessageEmbedField{
	{
		Name:   "Setting Alerts",
		Value:  "Use `!add`, example: `!add gmk,dandy,-daisy`\n" +
				"```- Include 'gmk' and 'dandy' and exclude 'daisy'.\n" +
				"- Alerts are case-insensitive.\n" +
				"- For multiple alerts separate them with a space.```",
		Inline: false,
	},
	{
		Name:   "Viewing Alerts",
		Value:  "Use `!list` to see a numbered list of your current alerts.\n" +
				"",
		Inline: false,
	},
	{
		Name:   "Deleting Alerts",
		Value:  "Use `!delete`, example: `!delete 3`\n" + 
				"```- Enter the number that corresponds to the alert you want to delete " +
				"based on the numbered list you can see with '!list'.\n" +
				"- For multiple deletions, separate numbers with a space\n" +
				"- To delete all alerts, use '!delete all'```",
		Inline: false,
	},
}

var MechfeedIntroEmbed *discordgo.MessageEmbed = &discordgo.MessageEmbed{
	Color: 0xe671dc,
	Title: "Welcome to mechfeed!",
 	Description: "This bot is currently in development. Please share any feedback and suggestions in the mechfeed #feedback channel.",
	Fields: append(HelpInformation, &discordgo.MessageEmbedField{
		Name:   "Help",
		Value:  "Use `!help` to see a list of all available commands.",
		Inline: false,
	}),
}

var MechfeedHelpEmbed *discordgo.MessageEmbed = &discordgo.MessageEmbed{
	Color: 0xe671dc,
	Title: "Commands",
	Fields: HelpInformation,
}
