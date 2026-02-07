package topicshandlers

import (
	"context"
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
	topicAddStateSelectEvent = "topic_add_state_select_event"
	topicAddStateEnterTopic  = "topic_add_state_enter_topic"

	// Context data keys
	topicAddCtxDataKeySelectedEventID   = "topic_add_ctx_data_selected_event_id"
	topicAddCtxDataKeySelectedEventName = "topic_add_ctx_data_selected_event_name"
	topicAddCtxDataKeyCancelFunc        = "topic_add_ctx_data_cancel_func"
	topicAddCtxDataKeyPreviousMessageID = "topic_add_ctx_data_previous_message_id"
	topicAddCtxDataKeyPreviousChatID    = "topic_add_ctx_data_previous_chat_id"

	// Callback data
	topicAddCallbackConfirmCancel = "topic_add_callback_confirm_cancel"
)

type topicAddHandler struct {
	config               *config.Config
	topicRepository      *repositories.TopicRepository
	eventRepository      *repositories.EventRepository
	messageSenderService *services.MessageSenderService
	userStore            *utils.UserDataStore
	permissionsService   *services.PermissionsService
}

func NewTopicAddHandler(
	config *config.Config,
	topicRepository *repositories.TopicRepository,
	eventRepository *repositories.EventRepository,
	messageSenderService *services.MessageSenderService,
	permissionsService *services.PermissionsService,
) ext.Handler {
	h := &topicAddHandler{
		config:               config,
		topicRepository:      topicRepository,
		eventRepository:      eventRepository,
		messageSenderService: messageSenderService,
		userStore:            utils.NewUserDataStore(),
		permissionsService:   permissionsService,
	}

	return handlers.NewConversation(
		[]ext.Handler{
			handlers.NewCommand(constants.TopicAddCommand, h.startTopicAdd),
		},
		map[string][]ext.Handler{
			topicAddStateSelectEvent: {
				handlers.NewMessage(message.All, h.handleEventSelection),
				handlers.NewCallback(callbackquery.Equal(topicAddCallbackConfirmCancel), h.handleCallbackCancel),
			},
			topicAddStateEnterTopic: {
				handlers.NewMessage(message.All, h.handleTopicEntry),
				handlers.NewCallback(callbackquery.Equal(topicAddCallbackConfirmCancel), h.handleCallbackCancel),
			},
		},
		&handlers.ConversationOpts{
			Exits: []ext.Handler{handlers.NewCommand(constants.CancelCommand, h.handleCancel)},
		},
	)
}

// 1. startTopicAdd is the entry point handler for adding a topic
func (h *topicAddHandler) startTopicAdd(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	// Only proceed if this is a private chat
	if !h.permissionsService.CheckPrivateChatType(msg) {
		return handlers.EndConversation()
	}

	// Check if user is a club member
	if !h.permissionsService.CheckClubMemberPermissions(msg, constants.TopicAddCommand) {
		return handlers.EndConversation()
	}

	// Get last actual events to show for selection
	events, err := h.eventRepository.GetLastActualEvents(10)
	if err != nil {
		h.messageSenderService.Reply(msg, "Error retrieving the list of events.", nil)
		log.Printf("%s: Error during events retrieval: %v", utils.GetCurrentTypeName(), err)
		return handlers.EndConversation()
	}

	if len(events) == 0 {
		h.messageSenderService.Reply(msg, "No events available for adding topics and questions.", nil)
		return handlers.EndConversation()
	}

	// Format and display event list for selection
	formattedEvents := formatters.FormatEventListForTopicsView(
		events,
		fmt.Sprintf("Select the event ID you want to add topics or questions to"),
	)

	sentMsg, _ := h.messageSenderService.ReplyMarkdownWithReturnMessage(
		msg,
		formattedEvents,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.CancelButton(topicAddCallbackConfirmCancel),
		},
	)

	h.SavePreviousMessageInfo(ctx.EffectiveUser.Id, sentMsg)
	return handlers.NextConversationState(topicAddStateSelectEvent)
}

// 2. handleEventSelection processes the user's event selection
func (h *topicAddHandler) handleEventSelection(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userInput := strings.TrimSpace(strings.Replace(msg.Text, "/", "", 1))

	// Check if the input is a valid event ID
	eventID, err := strconv.Atoi(userInput)
	if err != nil {
		h.messageSenderService.Reply(
			msg,
			fmt.Sprintf("Please send a valid event ID or use /%s to cancel.", constants.CancelCommand),
			nil,
		)
		return nil // Stay in the same state
	}

	// Get the event information
	event, err := h.eventRepository.GetEventByID(eventID)
	if err != nil {
		h.messageSenderService.Reply(
			msg,
			fmt.Sprintf("Could not find an event with ID %d. Please check the ID.", eventID),
			nil,
		)
		log.Printf("%s: Error during event retrieval: %v", utils.GetCurrentTypeName(), err)
		return nil // Stay in the same state
	}

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)

	// Store the selected event ID for later use when creating a new topic
	h.userStore.Set(ctx.EffectiveUser.Id, topicAddCtxDataKeySelectedEventID, eventID)
	h.userStore.Set(ctx.EffectiveUser.Id, topicAddCtxDataKeySelectedEventName, event.Name)

	// Prompt user to enter a topic
	sentMsg, _ := h.messageSenderService.ReplyMarkdownWithReturnMessage(
		msg,
		fmt.Sprintf("Send me topics and questions for the event *%s*:", event.Name),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.CancelButton(topicAddCallbackConfirmCancel),
		},
	)

	h.SavePreviousMessageInfo(ctx.EffectiveUser.Id, sentMsg)
	return handlers.NextConversationState(topicAddStateEnterTopic)
}

// 3. handleTopicEntry handles the user's topic input
func (h *topicAddHandler) handleTopicEntry(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	topicText := strings.TrimSpace(msg.Text)

	if topicText == "" {
		h.messageSenderService.Reply(
			msg,
			"Topic cannot be empty. Please enter the topic text or cancel the operation.",
			nil,
		)
		log.Printf("%s: Empty topic text", utils.GetCurrentTypeName())
		return nil // Stay in the same state
	}

	// Get the selected event ID from user store
	eventIDInterface, ok := h.userStore.Get(ctx.EffectiveUser.Id, topicAddCtxDataKeySelectedEventID)
	if !ok {
		h.messageSenderService.Reply(
			msg,
			"An error occurred: the selected event was not found. Please start over.",
			nil,
		)
		log.Printf("%s: Event ID not found in user store", utils.GetCurrentTypeName())
		return handlers.EndConversation()
	}

	eventID := eventIDInterface.(int)
	userNickname := "not specified"
	if ctx.EffectiveUser.Username != "" {
		userNickname = ctx.EffectiveUser.Username
	}

	// Create the new topic
	_, err := h.topicRepository.CreateTopic(topicText, userNickname, eventID)
	if err != nil {
		h.messageSenderService.Reply(msg, "Oops! Something went wrong...", nil)
		log.Printf("%s: Error during topic creation in database: %v", utils.GetCurrentTypeName(), err)
		return handlers.EndConversation()
	}

	// Send notification to admin about new topic
	eventName, _ := h.userStore.Get(ctx.EffectiveUser.Id, topicAddCtxDataKeySelectedEventName)
	adminChatID := h.config.AdminUserID

	adminMsg := fmt.Sprintf(
		"🔔 *New topic added*\n\n"+
			"_Event:_ %s\n"+
			"_Author:_ @%s\n"+
			"_Topic:_ %s",
		eventName,
		userNickname,
		topicText,
	)

	h.messageSenderService.SendMarkdown(adminChatID, adminMsg, nil)

	h.messageSenderService.Reply(
		msg,
		fmt.Sprintf(
			"Added! \nUse the /%s command to view all topics and questions for the event, or /%s to add new topics and questions.",
			constants.TopicsCommand,
			constants.TopicAddCommand,
		),
		nil,
	)

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)
	// Clean up user data
	h.userStore.Clear(ctx.EffectiveUser.Id)

	return handlers.EndConversation()
}

// handleCallbackCancel processes the cancel button click
func (h *topicAddHandler) handleCallbackCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	// Answer the callback query to remove the loading state on the button
	cb := ctx.Update.CallbackQuery
	_, _ = cb.Answer(b, nil)

	return h.handleCancel(b, ctx)
}

// 4. handleCancel handles the /cancel command
func (h *topicAddHandler) handleCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	// Check if there's an ongoing operation to cancel
	if cancelFunc, ok := h.userStore.Get(ctx.EffectiveUser.Id, topicAddCtxDataKeyCancelFunc); ok {
		// Call the cancel function to stop any ongoing API calls
		if cf, ok := cancelFunc.(context.CancelFunc); ok {
			cf()
			h.messageSenderService.Reply(msg, "Topic addition operation cancelled.", nil)
		}
	} else {
		h.messageSenderService.Reply(msg, "Topic addition operation cancelled.", nil)
	}

	h.MessageRemoveInlineKeyboard(b, &ctx.EffectiveUser.Id)

	// Clean up user data
	h.userStore.Clear(ctx.EffectiveUser.Id)

	return handlers.EndConversation()
}

func (h *topicAddHandler) MessageRemoveInlineKeyboard(b *gotgbot.Bot, userID *int64) {
	var chatID, messageID int64

	// If userID provided, get stored message info using the utility method
	if userID != nil {
		messageID, chatID = h.userStore.GetPreviousMessageInfo(
			*userID,
			topicAddCtxDataKeyPreviousMessageID,
			topicAddCtxDataKeyPreviousChatID,
		)
	}

	// Skip if we don't have valid chat and message IDs
	if chatID == 0 || messageID == 0 {
		return
	}

	// Use message sender service to remove the inline keyboard
	_ = h.messageSenderService.RemoveInlineKeyboard(chatID, messageID)
}

func (h *topicAddHandler) SavePreviousMessageInfo(userID int64, sentMsg *gotgbot.Message) {
	h.userStore.SetPreviousMessageInfo(userID, sentMsg.MessageId, sentMsg.Chat.Id,
		topicAddCtxDataKeyPreviousMessageID, topicAddCtxDataKeyPreviousChatID)
}
