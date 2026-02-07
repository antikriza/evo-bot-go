package buttons

import "github.com/PaulSonOfLars/gotgbot/v2"

func CancelButton(callbackData string) gotgbot.InlineKeyboardMarkup {
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u274c Cancel",
					CallbackData: callbackData,
				},
			},
		},
	}

	return inlineKeyboard
}

func ConfirmButton(callbackData string) gotgbot.InlineKeyboardMarkup {
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u2705 Confirm",
					CallbackData: callbackData,
				},
			},
		},
	}

	return inlineKeyboard
}

func ConfirmAndCancelButton(callbackDataYes string, callbackDataNo string) gotgbot.InlineKeyboardMarkup {
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u2705 Confirm",
					CallbackData: callbackDataYes,
				},
				{
					Text:         "\u274c Cancel",
					CallbackData: callbackDataNo,
				},
			},
		},
	}

	return inlineKeyboard
}

func BackAndCancelButton(callbackDataBack string, callbackDataCancel string) gotgbot.InlineKeyboardMarkup {
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u25c0\ufe0f Back",
					CallbackData: callbackDataBack,
				},
				{
					Text:         "\u274c Cancel",
					CallbackData: callbackDataCancel,
				},
			},
		},
	}

	return inlineKeyboard
}

func SearchTypeSelectionButton(callbackDataFast string, callbackDataDeep string, callbackDataCancel string) gotgbot.InlineKeyboardMarkup {
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "\u26a1 Fast",
					CallbackData: callbackDataFast,
				},
				{
					Text:         "\U0001f50d Deep",
					CallbackData: callbackDataDeep,
				},
			},
			{
				{
					Text:         "\u274c Cancel",
					CallbackData: callbackDataCancel,
				},
			},
		},
	}

	return inlineKeyboard
}
