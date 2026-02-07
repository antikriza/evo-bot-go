package buttons

import (
	"evo-bot-go/internal/constants"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func ProfileMainButtons() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u270f\ufe0f Edit",
					CallbackData: constants.ProfileEditMyProfileCallback,
				},
			},
			{
				{
					Text:         "\U0001f50e Search profile by name/username",
					CallbackData: constants.ProfileSearchProfileCallback,
				},
			},
			{
				{
					Text:         "\u274c Cancel",
					CallbackData: constants.ProfileFullCancel,
				},
			},
		},
	}
}

func ProfileEditBackCancelButtons(backCallbackData string) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u25c0\ufe0f Back",
					CallbackData: backCallbackData,
				},
				{
					Text:         "\u270f\ufe0f Edit",
					CallbackData: constants.ProfileEditMyProfileCallback,
				},
				{
					Text:         "\u274c Cancel",
					CallbackData: constants.ProfileFullCancel,
				},
			},
		},
	}
}

func ProfileBackCancelButtons(backCallbackData string) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u25c0\ufe0f Back",
					CallbackData: backCallbackData,
				},
				{
					Text:         "\u274c Cancel",
					CallbackData: constants.ProfileFullCancel,
				},
			},
		},
	}
}

func ProfileEditButtons(backCallbackData string) gotgbot.InlineKeyboardMarkup {
	buttons := [][]gotgbot.InlineKeyboardButton{
		{
			{
				Text:         "\U0001f464 First Name",
				CallbackData: constants.ProfileEditFirstnameCallback,
			},
			{
				Text:         "\U0001f464 Last Name",
				CallbackData: constants.ProfileEditLastnameCallback,
			},
			{
				Text:         "\U0001f4dd Bio",
				CallbackData: constants.ProfileEditBioCallback,
			},
		},
		{
			{
				Text:         "\u25c0\ufe0f Back",
				CallbackData: backCallbackData,
			},
			{
				Text:         "\u274c Cancel",
				CallbackData: constants.ProfileFullCancel,
			},
		},
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}
}
