package eventhandlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"evo-bot-go/internal/buttons"
	"evo-bot-go/internal/config"
	"evo-bot-go/internal/constants"
	"evo-bot-go/internal/database/repositories"
	"evo-bot-go/internal/services"
	"evo-bot-go/internal/utils"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

const (
	// Conversation states names
	eventSetupStateAskEventName      = "event_setup_state_ask_event_name"
	eventSetupStateAskEventType      = "event_setup_state_ask_event_type"
	eventSetupStateAskEventStartedAt = "event_setup_state_ask_event_started_at"

	// Context data keys
	eventSetupCtxDataKeyEventName         = "event_setup_ctx_data_event_name"
	eventSetupCtxDataKeyEventID           = "event_setup_ctx_data_event_id"
	eventSetupCtxDataKeyPreviousMessageID = "event_setup_ctx_data_previous_message_id"
	eventSetupCtxDataKeyPreviousChatID    = "event_setup_ctx_data_previous_chat_id"

	// Callback data
	eventSetupCallbackConfirmCancel = "event_setup_callback_confirm_cancel"
)

type eventSetupHandler struct {
	config               *config.Config
	eventRepository      *repositories.EventRepository
	messageSenderService *services.MessageSenderService
	userStore            *utils.UserDataStore
	permissionsService   *services.PermissionsService
}

func NewEventSetupHandler(
	config *config.Config,
	eventRepository *repositories.EventRepository,
	messageSenderService *services.MessageSenderService,
	permissionsService *services.PermissionsService,
) ext.Handler {
	h := &eventSetupHandler{
		config:               config,
		eventRepository:      eventRepository,
		messageSenderService: messageSenderService,
		userStore:            utils.NewUserDataStore(),
		permissionsService:   permissionsService,
	}

	return handlers.NewConversation(
		[]ext.Handler{
			handlers.NewCommand(constants.EventSetupCommand, h.startSetup),
		},
		map[string][]ext.Handler{
			eventSetupStateAskEventName: {
				handlers.NewMessage(message.Text, h.handleEventName),
				handlers.NewCallback(callbackquery.Equal(eventSetupCallbackConfirmCancel), h.handleCallbackCancel),
			},
			eventSetupStateAskEventType: {
				handlers.NewMessage(message.Text, h.handleEventType),
				handlers.NewCallback(callbackquery.Equal(eventSetupCallbackConfirmCancel), h.handleCallbackCancel),
			},
			eventSetupStateAskEventStartedAt: {
				handlers.NewMessage(message.Text, h.handleEventStartedAt),
				handlers.NewCallback(callbackquery.Equal(eventSetupCallbackConfirmCancel), h.handleCallbackCancel),
			},
		},
		&handlers.ConversationOpts{
			Exits: []ext.Handler{handlers.NewCommand(constants.CancelCommand, h.handleCancel)},
		},
	)
}

// 1. startSetup is the entry point handler for the setup conversation
func (h *eventSetupHandler) startSetup(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	// Check if user has admin permissions and is in a private chat
	if !h.permissionsService.CheckAdminAndPrivateChat(msg, constants.ShowTopicsCommand) {
		log.Printf("%s: User %d (%s) tried to use /%s without admin permissions.",
			utils.GetCurrentTypeName(),
			ctx.EffectiveUser.Id,
			ctx.EffectiveUser.Username,
			constants.ShowTopicsCommand,
		)
		return handlers.EndConversation()
	}

	sentMsg, _ := h.messageSenderService.ReplyWithReturnMessage(
		msg,
		"Please enter a name for the new event:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.CancelButton(eventSetupCallbackConfirmCancel),
		},
	)

	h.SavePreviousMessageInfo(ctx.EffectiveUser.Id, sentMsg)
	return handlers.NextConversationState(eventSetupStateAskEventName)
}

// 2. handleEventName processes the event name input
func (h *eventSetupHandler) handleEventName(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	eventName := strings.TrimSpace(msg.Text)

	if eventName == "" {
		h.messageSenderService.Reply(
			msg,
			"Name cannot be empty. Please enter a name for the event or use the cancel button.",
			nil,
		)
		return nil // Stay in the same state
	}

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)

	// Store the event name
	h.userStore.Set(ctx.EffectiveUser.Id, eventSetupCtxDataKeyEventName, eventName)

	// Ask for event type
	eventTypeOptions := []string{}
	for i, eventType := range constants.AllEventTypes {
		eventTypeOptions = append(eventTypeOptions, fmt.Sprintf("/%d. %s", i+1, eventType))
	}
	typeOptions := fmt.Sprintf("Select event type (enter a number):\n%s",
		strings.Join(eventTypeOptions, "\n"),
	)

	sentMsg, _ := h.messageSenderService.ReplyWithReturnMessage(
		msg,
		typeOptions,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.CancelButton(eventSetupCallbackConfirmCancel),
		},
	)

	h.SavePreviousMessageInfo(ctx.EffectiveUser.Id, sentMsg)
	return handlers.NextConversationState(eventSetupStateAskEventType)
}

// 3. handleEventType processes the event type selection and creates the event
func (h *eventSetupHandler) handleEventType(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	typeSelection := strings.TrimSpace(strings.Replace(msg.Text, "/", "", 1))

	var eventType constants.EventType

	// Convert typeSelection to integer
	index, err := strconv.Atoi(typeSelection)
	if err != nil || index < 1 || index > len(constants.AllEventTypes) {
		h.messageSenderService.Reply(
			msg,
			fmt.Sprintf("Invalid selection. Please enter a number from 1 to %d, or use the cancel button.",
				len(constants.AllEventTypes),
			),
			nil,
		)
		return nil // Stay in the same state
	}

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)

	// Arrays are 0-indexed but our options start from 1
	eventType = constants.AllEventTypes[index-1]

	// Get the event name from user data store
	eventNameVal, ok := h.userStore.Get(ctx.EffectiveUser.Id, eventSetupCtxDataKeyEventName)
	if !ok {
		h.messageSenderService.Reply(
			msg,
			fmt.Sprintf("An internal error occurred. Could not find the event name. Try starting over with /%s.",
				constants.EventSetupCommand,
			),
			nil,
		)
		return handlers.EndConversation()
	}

	eventName, ok := eventNameVal.(string)
	if !ok {
		h.messageSenderService.Reply(
			msg,
			fmt.Sprintf("An internal error occurred (invalid name type). Try starting over with /%s.",
				constants.EventSetupCommand,
			),
			nil,
		)
		return handlers.EndConversation()
	}

	// Create event in the database
	id, err := h.eventRepository.CreateEvent(eventName, eventType)
	if err != nil {
		h.messageSenderService.Reply(msg, "An error occurred while creating the event record.", nil)
		log.Printf("%s: Error during event creation: %v", utils.GetCurrentTypeName(), err)
		return handlers.EndConversation()
	}

	// Store the event ID
	h.userStore.Set(ctx.EffectiveUser.Id, eventSetupCtxDataKeyEventID, id)

	// Ask for start date
	sentMsg, _ := h.messageSenderService.ReplyWithReturnMessage(
		msg,
		"When does the event start? Enter date and time in DD.MM.YYYY HH:MM format:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.CancelButton(eventSetupCallbackConfirmCancel),
		},
	)

	h.SavePreviousMessageInfo(ctx.EffectiveUser.Id, sentMsg)
	return handlers.NextConversationState(eventSetupStateAskEventStartedAt)
}

// 4. handleEventStartedAt processes the start date input and updates the event
func (h *eventSetupHandler) handleEventStartedAt(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	dateTimeStr := strings.TrimSpace(msg.Text)

	// Parse the start date
	startedAt, err := time.Parse("02.01.2006 15:04", dateTimeStr)
	if err != nil {
		h.messageSenderService.Reply(
			msg,
			"Invalid date format. Please enter date and time in DD.MM.YYYY HH:MM format or use the cancel button.",
			nil,
		)
		return nil // Stay in the same state
	}

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)

	// Get event ID from user data store
	eventIDVal, ok := h.userStore.Get(ctx.EffectiveUser.Id, eventSetupCtxDataKeyEventID)
	if !ok {
		h.messageSenderService.Reply(
			msg,
			fmt.Sprintf("An internal error occurred. Could not find the event ID. Try starting over with /%s.",
				constants.EventSetupCommand,
			),
			nil,
		)
		return handlers.EndConversation()
	}

	eventID, ok := eventIDVal.(int)
	if !ok {
		h.messageSenderService.Reply(
			msg,
			fmt.Sprintf("An internal error occurred (invalid ID type). Try starting over with /%s.",
				constants.EventSetupCommand,
			),
			nil,
		)
		return handlers.EndConversation()
	}

	// Update the started_at field
	err = h.eventRepository.UpdateEventStartedAt(eventID, startedAt)
	if err != nil {
		h.messageSenderService.Reply(msg, "An error occurred while updating the event start date.", nil)
		log.Printf("%s: Error during event update: %v", utils.GetCurrentTypeName(), err)
		return handlers.EndConversation()
	}

	// Get event name for the success message
	eventNameVal, ok := h.userStore.Get(ctx.EffectiveUser.Id, eventSetupCtxDataKeyEventName)
	if !ok {
		h.messageSenderService.Reply(msg, "Event successfully created with start date.", nil)
		h.userStore.Clear(ctx.EffectiveUser.Id)
		return handlers.EndConversation()
	}

	eventName, ok := eventNameVal.(string)
	if !ok {
		h.messageSenderService.Reply(msg, "Event successfully created with start date.", nil)
		h.userStore.Clear(ctx.EffectiveUser.Id)
		return handlers.EndConversation()
	}

	// Success message
	h.messageSenderService.Reply(
		msg,
		fmt.Sprintf(
			"Event record '*%s*' successfully created with ID: %d and start date: *%s*\n\nTo edit the event, use the /%s command.\nTo view all commands, use /%s",
			eventName, eventID, startedAt.Format("02.01.2006 15:04"), constants.EventEditCommand, constants.HelpCommand,
		),
		&gotgbot.SendMessageOpts{
			ParseMode: "Markdown",
		},
	)

	// Clean up user data
	h.userStore.Clear(ctx.EffectiveUser.Id)

	return handlers.EndConversation()
}

// handleCallbackCancel processes the cancel button click
func (h *eventSetupHandler) handleCallbackCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	// Answer the callback query to remove the loading state on the button
	cb := ctx.Update.CallbackQuery
	_, _ = cb.Answer(b, nil)

	return h.handleCancel(b, ctx)
}

// 5. handleCancel handles the /cancel command
func (h *eventSetupHandler) handleCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)
	h.messageSenderService.Reply(msg, "Event creation operation canceled.", nil)

	// Clean up user data
	h.userStore.Clear(ctx.EffectiveUser.Id)

	return handlers.EndConversation()
}

func (h *eventSetupHandler) MessageRemoveInlineKeyboard(b *gotgbot.Bot, userID *int64) {
	var chatID, messageID int64

	// If userID provided, get stored message info using the utility method
	if userID != nil {
		messageID, chatID = h.userStore.GetPreviousMessageInfo(
			*userID,
			eventSetupCtxDataKeyPreviousMessageID,
			eventSetupCtxDataKeyPreviousChatID,
		)
	}

	// Skip if we don't have valid chat and message IDs
	if chatID == 0 || messageID == 0 {
		return
	}

	// Use message sender service to remove the inline keyboard
	_ = h.messageSenderService.RemoveInlineKeyboard(chatID, messageID)
}

func (h *eventSetupHandler) SavePreviousMessageInfo(userID int64, sentMsg *gotgbot.Message) {
	h.userStore.SetPreviousMessageInfo(userID, sentMsg.MessageId, sentMsg.Chat.Id,
		eventSetupCtxDataKeyPreviousMessageID, eventSetupCtxDataKeyPreviousChatID)
}
