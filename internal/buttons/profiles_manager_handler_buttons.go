package buttons

import (
	"evo-bot-go/internal/constants"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func ProfilesBackCancelButtons(backCallbackData string) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u25c0\ufe0f Back",
					CallbackData: backCallbackData,
				},
				{
					Text:         "\u274c Cancel",
					CallbackData: constants.AdminProfilesCancelCallback,
				},
			},
		},
	}
}

func ProfilesBackStartCancelButtons(backCallbackData string) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u25c0\ufe0f Back",
					CallbackData: backCallbackData,
				},
				{
					Text:         "\u23ea Start",
					CallbackData: constants.AdminProfilesStartCallback,
				},
				{
					Text:         "\u274c Cancel",
					CallbackData: constants.AdminProfilesCancelCallback,
				},
			},
		},
	}
}

// ProfilesCoffeeBanButtons returns buttons for managing coffee ban status
func ProfilesCoffeeBanButtons(backCallbackData string, hasCoffeeBan bool) gotgbot.InlineKeyboardMarkup {
	var toggleButtonText string
	if hasCoffeeBan {
		toggleButtonText = "\u2705 Allow"
	} else {
		toggleButtonText = "\u274c Ban"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         toggleButtonText,
					CallbackData: constants.AdminProfilesToggleCoffeeBanCallback,
				},
			},
			{
				{
					Text:         "\u25c0\ufe0f Back",
					CallbackData: backCallbackData,
				},
				{
					Text:         "\u274c Cancel",
					CallbackData: constants.AdminProfilesCancelCallback,
				},
			},
		},
	}
}

func ProfilesEditMenuButtons(backCallbackData string) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\U0001f464 First Name",
					CallbackData: constants.AdminProfilesEditFirstnameCallback,
				},
				{
					Text:         "\U0001f464 Last Name",
					CallbackData: constants.AdminProfilesEditLastnameCallback,
				},
				{
					Text:         "\U0001f464 Username",
					CallbackData: constants.AdminProfilesEditUsernameCallback,
				},
			},
			{
				{
					Text:         "\U0001f4dd Bio",
					CallbackData: constants.AdminProfilesEditBioCallback,
				},
				{
					Text:         "\u2615\ufe0f Coffee?",
					CallbackData: constants.AdminProfilesEditCoffeeBanCallback,
				},
			},
			{
				{
					Text:         "\U0001f4e2 Publish (+ preview)",
					CallbackData: constants.AdminProfilesPublishCallback,
				},
				{
					Text:         "\U0001f4e2 Publish (- preview)",
					CallbackData: constants.AdminProfilesPublishNoPreviewCallback,
				},
			},
			{
				{
					Text:         "\u25c0\ufe0f Back",
					CallbackData: backCallbackData,
				},
				{
					Text:         "\u274c Cancel",
					CallbackData: constants.AdminProfilesCancelCallback,
				},
			},
		},
	}
}

func ProfilesMainMenuButtons() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\U0001f4dd Search by Telegram Username",
					CallbackData: constants.AdminProfilesSearchByUsernameCallback,
				},
			},
			{
				{
					Text:         "\U0001f50d Search by Telegram ID",
					CallbackData: constants.AdminProfilesSearchByTelegramIDCallback,
				},
			},
			{
				{
					Text:         "\U0001f50d Search by full name",
					CallbackData: constants.AdminProfilesSearchByFullNameCallback,
				},
			},
			{
				{
					Text:         "\u2795 Create profile (via forward)",
					CallbackData: constants.AdminProfilesCreateByForwardedMessageCallback,
				},
			},
			{
				{
					Text:         "\U0001f194 Create profile by Telegram ID",
					CallbackData: constants.AdminProfilesCreateByTelegramIDCallback,
				},
			},
			{
				{
					Text:         "\u274c Cancel",
					CallbackData: constants.AdminProfilesCancelCallback,
				},
			},
		},
	}
}
