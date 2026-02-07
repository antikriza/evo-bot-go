package formatters

import (
	"evo-bot-go/internal/config"
	"evo-bot-go/internal/constants"
	"fmt"
)

// FormatHelpMessage generates the help message text with appropriate commands based on user permissions
func FormatHelpMessage(isAdmin bool, config *config.Config) string {
	helpText := "<b>📋 Bot Features</b>\n\n" +
		"<b>🏠 Basic Commands</b>\n" +
		"└ /start - Welcome message\n" +
		"└ /help - Show this command list\n" +
		"└ /cancel - Force-cancel any active dialog\n\n" +
		"<b>👤 Profile</b>\n" +
		"└ /profile - Manage your profile, search members, publish your info in the Intro channel\n\n" +
		"<b>🔍 AI Search</b>\n" +
		"└ /tools - Find AI tools from the Tools channel\n" +
		"└ /content - Find content from the Video Content channel\n" +
		"└ /intro - Find member info from the Intro channel (smart profile search)\n\n" +
		"<b>📅 Events</b>\n" +
		"└ /events - View upcoming events\n" +
		"└ /topics - View topics and questions for upcoming events\n" +
		"└ /topicAdd - Suggest a topic or question for an event"

	helpText += "\n\n" +
		"<i>📖 <a href=\"https://antikriza.github.io/BBD-evolution-code-clone/telegram-archive/course/twa/index.html\">Open AI Course (42 topics)</a></i>"

	if isAdmin {
		adminHelpText := "\n\n<b>🔐 Admin Commands</b>\n" +
			fmt.Sprintf("└ /%s - Start an event\n", constants.EventStartCommand) +
			fmt.Sprintf("└ /%s - Create a new event\n", constants.EventSetupCommand) +
			fmt.Sprintf("└ /%s - Edit an event\n", constants.EventEditCommand) +
			fmt.Sprintf("└ /%s - Delete an event\n", constants.EventDeleteCommand) +
			fmt.Sprintf("└ /%s - View topics with <b>delete option</b>\n", constants.ShowTopicsCommand) +
			fmt.Sprintf("└ /%s - Enter auth code for TG client\n", constants.CodeCommand) +
			fmt.Sprintf("└ /%s - Manage member profiles", constants.AdminProfilesCommand)

		testCommandsHelpText := "\n\n<b>⚙️ Test Commands</b>\n" +
			fmt.Sprintf("└ /%s - Send course link in DM\n", constants.TryLinkToLearnCommand)

		helpText += adminHelpText
		helpText += testCommandsHelpText
	}

	return helpText
}
