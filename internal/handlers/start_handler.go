package handlers

import (
	"evo-bot-go/internal/config"
	"evo-bot-go/internal/constants"
	"evo-bot-go/internal/formatters"
	"evo-bot-go/internal/services"
	"evo-bot-go/internal/utils"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
)

const (
	// Conversation states names
	startHandlerStateProcessCallback = "start_handler_state_process_callback"
	// Callbacks names
	startHandlerCallbackHelp = "start_handler_callback_help"
)

type startHandler struct {
	config               *config.Config
	messageSenderService *services.MessageSenderService
	permissionsService   *services.PermissionsService
}

func NewStartHandler(
	config *config.Config,
	messageSenderService *services.MessageSenderService,
	permissionsService *services.PermissionsService,
) ext.Handler {
	h := &startHandler{
		config:               config,
		messageSenderService: messageSenderService,
		permissionsService:   permissionsService,
	}
	return handlers.NewConversation(
		[]ext.Handler{
			handlers.NewCommand(constants.StartCommand, h.handleStart),
		},
		map[string][]ext.Handler{
			startHandlerStateProcessCallback: {
				handlers.NewCallback(callbackquery.Equal(startHandlerCallbackHelp), h.handleCallbackHelp),
			},
		},
		nil,
	)
}

func (h *startHandler) handleStart(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveUser
	msg := ctx.EffectiveMessage
	// Only proceed if this is a private chat
	if !h.permissionsService.CheckPrivateChatType(msg) {
		return handlers.EndConversation()
	}

	userName := ""
	if user.FirstName != "" {
		userName = user.FirstName
	}

	greeting := "Welcome"
	if userName != "" {
		greeting += ", *" + userName + "*"
	}
	greeting += "! 🎓"

	// Check if user is a member of the group
	isGroupMember := utils.IsUserClubMember(b, user.Id, h.config)

	var message string
	var inlineKeyboard gotgbot.InlineKeyboardMarkup
	if isGroupMember {
		// Message for group members
		message = greeting + "\n\n" +
			"I'm the *AI & Programming Course Bot* — your assistant for learning AI, discovering tools, and connecting with fellow learners. 🤖\n\n" +
			"Use /help to see what I can do for you!"

		inlineKeyboard = gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         "📋 Show commands",
						CallbackData: startHandlerCallbackHelp,
					},
				},
				{
					{
						Text: "📖 Open AI Course",
						Url:  "https://antikriza.github.io/BBD-evolution-code-clone/telegram-archive/course/twa/index.html",
					},
				},
			},
		}
	} else {
		// Message for non-members
		message = greeting + "\n\n" +
			"I'm the *AI & Programming Course Bot*. 🤖\n\n" +
			"Join our community to access AI tools search, daily summaries, member profiles, and a structured AI course with 42 topics!"

		inlineKeyboard = gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         "📋 Show commands",
						CallbackData: startHandlerCallbackHelp,
					},
				},
			},
		}
	}

	h.messageSenderService.ReplyMarkdown(
		msg,
		message,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: inlineKeyboard,
		},
	)

	return handlers.NextConversationState(startHandlerStateProcessCallback)
}

func (h *startHandler) handleCallbackHelp(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	_, _ = cb.Answer(b, nil)

	user := ctx.EffectiveUser
	isAdmin := utils.IsUserAdminOrCreator(b, user.Id, h.config)
	helpText := formatters.FormatHelpMessage(isAdmin, h.config)

	h.messageSenderService.ReplyHtml(ctx.EffectiveMessage, helpText, nil)

	return handlers.EndConversation()
}
