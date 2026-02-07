package eventhandlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"evo-bot-go/internal/buttons"
	"evo-bot-go/internal/config"
	"evo-bot-go/internal/constants"
	"evo-bot-go/internal/database/repositories"
	"evo-bot-go/internal/formatters"
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
	eventStartStateSelectEvent = "event_start_state_select_event"
	eventStartStateEnterLink   = "event_start_state_enter_link"
	eventStartStateConfirm     = "event_start_state_confirm"

	// Context data keys
	eventStartCtxDataKeySelectedEventID   = "event_start_ctx_data_key_selected_event_id"
	eventStartCtxDataKeyEventLink         = "event_start_ctx_data_key_event_link"
	eventStartCtxDataKeyPreviousMessageID = "event_start_ctx_data_key_previous_message_id"
	eventStartCtxDataKeyPreviousChatID    = "event_start_ctx_data_key_previous_chat_id"

	// Callbacks names
	eventStartCallbackConfirmYes    = "event_start_callback_confirm_yes"
	eventStartCallbackConfirmCancel = "event_start_callback_confirm_cancel"
)

// Confirmation message options
const (
	eventStartConfirmYes = "yes"
	eventStartConfirmNo  = "no"
)

type eventStartHandler struct {
	config               *config.Config
	eventRepository      *repositories.EventRepository
	messageSenderService *services.MessageSenderService
	userStore            *utils.UserDataStore
	permissionsService   *services.PermissionsService
}

func NewEventStartHandler(
	config *config.Config,
	eventRepository *repositories.EventRepository,
	messageSenderService *services.MessageSenderService,
	permissionsService *services.PermissionsService,
) ext.Handler {
	h := &eventStartHandler{
		config:               config,
		eventRepository:      eventRepository,
		messageSenderService: messageSenderService,
		userStore:            utils.NewUserDataStore(),
		permissionsService:   permissionsService,
	}

	return handlers.NewConversation(
		[]ext.Handler{
			handlers.NewCommand(constants.EventStartCommand, h.startEvent),
		},
		map[string][]ext.Handler{
			eventStartStateSelectEvent: {
				handlers.NewMessage(message.Text, h.handleSelectEvent),
				handlers.NewCallback(callbackquery.Equal(eventStartCallbackConfirmCancel), h.handleCallbackCancel),
			},
			eventStartStateEnterLink: {
				handlers.NewMessage(message.Text, h.handleEnterLink),
				handlers.NewCallback(callbackquery.Equal(eventStartCallbackConfirmCancel), h.handleCallbackCancel),
			},
			eventStartStateConfirm: {
				handlers.NewCallback(callbackquery.Equal(eventStartCallbackConfirmYes), h.handleCallbackYes),
				handlers.NewCallback(callbackquery.Equal(eventStartCallbackConfirmCancel), h.handleCallbackCancel),
				handlers.NewMessage(message.All, h.handleTextDuringConfirmation),
			},
		},
		&handlers.ConversationOpts{
			Exits: []ext.Handler{handlers.NewCommand(constants.CancelCommand, h.handleCancel)},
		},
	)
}

// 1. startEvent is the entry point handler for the start conversation
func (h *eventStartHandler) startEvent(b *gotgbot.Bot, ctx *ext.Context) error {
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

	// Get a list of active events
	events, err := h.eventRepository.GetLastEvents(constants.EventEditGetLastLimit)
	if err != nil {
		h.messageSenderService.Reply(msg, "An error occurred while retrieving the list of current events.", nil)
		log.Printf("%s: Error during event retrieval: %v", utils.GetCurrentTypeName(), err)
		return handlers.EndConversation()
	}

	if len(events) == 0 {
		h.messageSenderService.Reply(msg, "No events available to start.", nil)
		return handlers.EndConversation()
	}

	title := fmt.Sprintf("Last %d events:", len(events))
	actionDescription := "that you want to start"
	formattedResponse := formatters.FormatEventListForAdmin(events, title, constants.CancelCommand, actionDescription)

	sentMsg, _ := h.messageSenderService.ReplyMarkdownWithReturnMessage(msg, formattedResponse, &gotgbot.SendMessageOpts{
		ReplyMarkup: buttons.CancelButton(eventStartCallbackConfirmCancel),
	})

	h.SavePreviousMessageInfo(ctx.EffectiveUser.Id, sentMsg)
	return handlers.NextConversationState(eventStartStateSelectEvent)
}

// 2. handleSelectEvent processes the user's selection of an event to finish
func (h *eventStartHandler) handleSelectEvent(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	eventIDStr := strings.TrimSpace(strings.Replace(msg.Text, "/", "", 1))

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		h.messageSenderService.Reply(msg, fmt.Sprintf("Invalid ID. Please enter a numeric ID or use /%s to cancel.", constants.CancelCommand), nil)
		return nil // Stay in the same state
	}

	// Store the selected event ID
	h.userStore.Set(ctx.EffectiveUser.Id, eventStartCtxDataKeySelectedEventID, eventID)

	// Get event details to show in the prompt
	event, err := h.eventRepository.GetEventByID(eventID)
	if err != nil {
		h.messageSenderService.Reply(msg, fmt.Sprintf("Error retrieving event with ID %d", eventID), nil)
		log.Printf("%s: Error during event retrieval: %v", utils.GetCurrentTypeName(), err)
		return handlers.EndConversation()
	}

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)
	sentMsg, _ := h.messageSenderService.ReplyMarkdownWithReturnMessage(
		msg,
		fmt.Sprintf(
			"🔗 Send me the link to the event '%s' (ID: %d)\nThis link will be sent to the announcements chat.",
			event.Name, event.ID,
		),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.CancelButton(eventStartCallbackConfirmCancel),
		},
	)

	h.SavePreviousMessageInfo(ctx.EffectiveUser.Id, sentMsg)
	return handlers.NextConversationState(eventStartStateEnterLink)
}

// 3. handleEnterLink processes the user's input of event link
func (h *eventStartHandler) handleEnterLink(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	eventLink := strings.TrimSpace(msg.Text)

	// Simple validation for the link
	if !strings.HasPrefix(eventLink, "http://") && !strings.HasPrefix(eventLink, "https://") {
		h.messageSenderService.Reply(msg, fmt.Sprintf(
			"Please enter a valid link starting with http:// or https:// (or use /%s to cancel):",
			constants.CancelCommand,
		), nil)
		return nil // Stay in the same state
	}

	// Store the event link
	h.userStore.Set(ctx.EffectiveUser.Id, eventStartCtxDataKeyEventLink, eventLink)

	// Get the event ID
	eventIDVal, ok := h.userStore.Get(ctx.EffectiveUser.Id, eventStartCtxDataKeySelectedEventID)
	if !ok {
		h.messageSenderService.Reply(msg, fmt.Sprintf(
			"An error occurred while retrieving the selected event. Please start over with /%s",
			constants.EventStartCommand,
		), nil)
		log.Printf("%s: Error during event retrieval.", utils.GetCurrentTypeName())
		return handlers.EndConversation()
	}

	eventID, ok := eventIDVal.(int)
	if !ok {
		log.Println("Invalid event ID type:", eventIDVal)
		h.messageSenderService.Reply(msg, fmt.Sprintf(
			"An internal error occurred (invalid ID type). Please start over with /%s",
			constants.EventStartCommand,
		), nil)
		log.Printf("%s: Invalid event ID type: %v", utils.GetCurrentTypeName(), eventIDVal)
		return handlers.EndConversation()
	}

	// Get event details to show in the confirmation message
	event, err := h.eventRepository.GetEventByID(eventID)
	if err != nil {
		h.messageSenderService.Reply(msg, fmt.Sprintf("Error retrieving event with ID %d", eventID), nil)
		log.Printf("%s: Error during event retrieval: %v", utils.GetCurrentTypeName(), err)
		return handlers.EndConversation()
	}

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)

	sentMsg, err := h.messageSenderService.ReplyMarkdownWithReturnMessage(msg, fmt.Sprintf(
		"*Event launch confirmation*\n\n🎯 *%s* _(ID: %d)_\n\n🔗 Link: `%s`\n\nThis link will be sent to the announcements chat.\n\nClick the button below to confirm or cancel",
		event.Name, event.ID, eventLink,
	), &gotgbot.SendMessageOpts{
		ReplyMarkup: buttons.ConfirmAndCancelButton(eventStartCallbackConfirmYes, eventStartCallbackConfirmCancel),
	})

	h.SavePreviousMessageInfo(ctx.EffectiveUser.Id, sentMsg)
	return handlers.NextConversationState(eventStartStateConfirm)
}

// handleCallbackYes processes the confirmation button click
func (h *eventStartHandler) handleCallbackYes(b *gotgbot.Bot, ctx *ext.Context) error {
	// Answer the callback query to remove the loading state on the button
	cb := ctx.Update.CallbackQuery
	_, _ = cb.Answer(b, nil)

	// Get the selected event ID
	eventIDVal, ok := h.userStore.Get(ctx.EffectiveUser.Id, eventStartCtxDataKeySelectedEventID)
	if !ok {
		h.messageSenderService.Reply(ctx.EffectiveMessage, fmt.Sprintf(
			"An error occurred while retrieving the selected event. Please start over with /%s",
			constants.EventStartCommand,
		), nil)
		log.Printf("%s: Error during event retrieval.", utils.GetCurrentTypeName())
		return handlers.EndConversation()
	}

	eventID, ok := eventIDVal.(int)
	if !ok {
		log.Println("Invalid event ID type:", eventIDVal)
		h.messageSenderService.Reply(ctx.EffectiveMessage, fmt.Sprintf(
			"An internal error occurred (invalid ID type). Please start over with /%s",
			constants.EventStartCommand,
		), nil)
		log.Printf("%s: Invalid event ID type: %v", utils.GetCurrentTypeName(), eventIDVal)
		return handlers.EndConversation()
	}

	// Get the event link
	eventLinkVal, ok := h.userStore.Get(ctx.EffectiveUser.Id, eventStartCtxDataKeyEventLink)
	if !ok {
		h.messageSenderService.Reply(ctx.EffectiveMessage, fmt.Sprintf(
			"An error occurred while retrieving the event link. Please start over with /%s",
			constants.EventStartCommand,
		), nil)
		log.Printf("%s: Error during event link retrieval.", utils.GetCurrentTypeName())
		return handlers.EndConversation()
	}

	eventLink, ok := eventLinkVal.(string)
	if !ok {
		log.Println("Invalid event link type:", eventLinkVal)
		h.messageSenderService.Reply(ctx.EffectiveMessage, fmt.Sprintf(
			"An internal error occurred (invalid link type). Please start over with /%s",
			constants.EventStartCommand,
		), nil)
		log.Printf("%s: Invalid event link type: %v", utils.GetCurrentTypeName(), eventLinkVal)
		return handlers.EndConversation()
	}

	// Get the event details for the success message
	event, err := h.eventRepository.GetEventByID(eventID)
	if err != nil {
		h.messageSenderService.Reply(ctx.EffectiveMessage, fmt.Sprintf("Error retrieving event with ID %d", eventID), nil)
		log.Printf("%s: Error during event retrieval: %v", utils.GetCurrentTypeName(), err)
		return handlers.EndConversation()
	}

	// Update the event status to active (or use the appropriate constant)
	err = h.eventRepository.UpdateEventStatus(eventID, constants.EventStatusFinished) // When ivent already started in DB we need to set status to finished
	if err != nil {
		h.messageSenderService.Reply(ctx.EffectiveMessage, "An error occurred while updating the event status.", nil)
		log.Printf("%s: Error during event update: %v", utils.GetCurrentTypeName(), err)
		return handlers.EndConversation()
	}

	buttonWithLink := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text: "🔗 Zoom Link",
					Url:  eventLink,
				},
			},
		},
	}

	// Send announcement message with the event link to the announcement topic if configured
	announcementMsg := fmt.Sprintf(
		"🔴 *EVENT STARTING!* 🔴\n\n%s *%s*\n",
		formatters.GetTypeEmoji(constants.EventType(event.Type)),
		event.Name,
	)

	if event.Type == string(constants.EventTypeClubCall) {
		announcementMsg += fmt.Sprintf("💡 [About the format and rules of community calls](https://t.me/c/2069889012/127/33823)\n")
	}

	announcementMsg += fmt.Sprintf("\nUse the button below to join ⬇️")

	sentAnnouncementMsg, err := h.messageSenderService.SendMarkdownWithReturnMessage(
		utils.ChatIdToFullChatId(h.config.SuperGroupChatID),
		announcementMsg,
		&gotgbot.SendMessageOpts{
			MessageThreadId: int64(h.config.AnnouncementTopicID),
			ReplyMarkup:     buttonWithLink,
		},
	)

	// Pin the announcement message with notification for all users
	if err == nil && sentAnnouncementMsg != nil {
		err = h.messageSenderService.PinMessage(
			sentAnnouncementMsg.Chat.Id,
			sentAnnouncementMsg.MessageId,
			true,
		)
		if err != nil {
			log.Printf("%s: Error pinning announcement message: %v", utils.GetCurrentTypeName(), err)
		}
	}

	// Confirmation message
	h.messageSenderService.ReplyMarkdown(
		ctx.EffectiveMessage,
		fmt.Sprintf("✅ *Event successfully started!*\n\n🎯 *%s* _(ID: %d)_\n\n📢 Link sent to the announcements chat.", event.Name, event.ID),
		nil,
	)

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)

	// Clean up user data
	h.userStore.Clear(ctx.EffectiveUser.Id)

	return handlers.EndConversation()
}

// handleTextDuringConfirmation handles text messages during the confirmation state
func (h *eventStartHandler) handleTextDuringConfirmation(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Printf("%s: User %d sent text during confirmation", utils.GetCurrentTypeName(), ctx.EffectiveUser.Id)

	h.messageSenderService.Reply(
		ctx.EffectiveMessage,
		fmt.Sprintf("Please click one of the buttons above, or use /%s to cancel.", constants.CancelCommand),
		nil,
	)
	return nil // Stay in the same state
}

// handleCallbackCancel processes the cancel button click
func (h *eventStartHandler) handleCallbackCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	// Answer the callback query to remove the loading state on the button
	cb := ctx.Update.CallbackQuery
	_, _ = cb.Answer(b, nil)

	return h.handleCancel(b, ctx)
}

// handleCancel handles the /cancel command
func (h *eventStartHandler) handleCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	h.messageSenderService.Reply(msg, "Event start operation canceled.", nil)
	log.Printf("%s: Event start canceled", utils.GetCurrentTypeName())

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)

	// Clean up user data
	h.userStore.Clear(ctx.EffectiveUser.Id)

	return handlers.EndConversation()
}

func (h *eventStartHandler) MessageRemoveInlineKeyboard(b *gotgbot.Bot, userID *int64) {
	var chatID, messageID int64

	// If userID provided, get stored message info using the utility method
	if userID != nil {
		messageID, chatID = h.userStore.GetPreviousMessageInfo(
			*userID,
			eventStartCtxDataKeyPreviousMessageID,
			eventStartCtxDataKeyPreviousChatID,
		)
	}

	// Skip if we don't have valid chat and message IDs
	if chatID == 0 || messageID == 0 {
		return
	}

	// Use message sender service to remove the inline keyboard
	_ = h.messageSenderService.RemoveInlineKeyboard(chatID, messageID)
}

func (h *eventStartHandler) SavePreviousMessageInfo(userID int64, sentMsg *gotgbot.Message) {
	h.userStore.SetPreviousMessageInfo(userID, sentMsg.MessageId, sentMsg.Chat.Id,
		eventStartCtxDataKeyPreviousMessageID, eventStartCtxDataKeyPreviousChatID)
}
