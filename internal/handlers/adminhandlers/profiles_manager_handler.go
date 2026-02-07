package adminhandlers

import (
	"database/sql"
	"evo-bot-go/internal/buttons"
	"evo-bot-go/internal/config"
	"evo-bot-go/internal/constants"
	"evo-bot-go/internal/database/repositories"
	"evo-bot-go/internal/formatters"
	"evo-bot-go/internal/services"
	"evo-bot-go/internal/utils"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

const (
	// Conversation states
	adminProfilesStateStart                         = "admin_profiles_state_start"
	adminProfilesStateEdit                          = "admin_profiles_state_edit"
	adminProfilesStateAwaitSearchByUsername         = "admin_profiles_state_await_search_by_username"
	adminProfilesStateAwaitSearchUserID             = "admin_profiles_state_await_search_user_id"
	adminProfilesStateAwaitSearchByFullName         = "admin_profiles_state_await_search_by_full_name"
	adminProfilesStateAwaitCreateByForwardedMessage = "admin_profiles_state_await_create_by_forwarded_message"
	adminProfilesStateAwaitCreateByTelegramID       = "admin_profiles_state_await_create_by_telegram_id"
	adminProfilesStateEditProfile                   = "admin_profiles_state_edit_profile"
	adminProfilesStateAwaitBio                      = "admin_profiles_state_await_bio"
	adminProfilesStateAwaitFirstname                = "admin_profiles_state_await_firstname"
	adminProfilesStateAwaitLastname                 = "admin_profiles_state_await_lastname"
	adminProfilesStateAwaitCoffeeBan                = "admin_profiles_state_await_coffee_ban"
	adminProfilesStateAwaitUsername                 = "admin_profiles_state_await_username"

	// UserStore keys
	adminProfilesCtxDataKeyField                   = "admin_profiles_ctx_data_field"
	adminProfilesCtxDataKeyUserID                  = "admin_profiles_ctx_data_user_id"
	adminProfilesCtxDataKeyPreviousMessageID       = "admin_profiles_ctx_data_previous_message_id"
	adminProfilesCtxDataKeyPreviousChatID          = "admin_profiles_ctx_data_previous_chat_id"
	adminProfilesCtxDataKeyProfileID               = "admin_profiles_ctx_data_profile_id"
	adminProfilesCtxDataKeyTelegramID              = "admin_profiles_ctx_data_telegram_id"
	adminProfilesCtxDataKeyTelegramUsername        = "admin_profiles_ctx_data_telegram_username"
	adminProfilesCtxDataKeyLastMessageTimeFromUser = "admin_profiles_ctx_data_last_message_time_from_user"

	// Menu headers
	adminProfilesMenuHeader              = "Admin Menu \"Profile Manager\""
	adminProfilesMenuEditHeader          = "Profile Manager → Edit"
	adminProfilesMenuCreateByIDHeader    = "Profile Manager → Create by ID"
	adminProfilesMenuSearchByIDHeader    = "Profile Manager → Search by ID"
	adminProfilesMenuSearchByNameHeader  = "Profile Manager → Search by Name"
	adminProfilesMenuEditFirstnameHeader = "Profile Manager → Edit → First Name"
	adminProfilesMenuEditLastnameHeader  = "Profile Manager → Edit → Last Name"
	adminProfilesMenuEditBioHeader       = "Profile Manager → Edit → Bio"
	adminProfilesMenuEditUsernameHeader  = "Profile Manager → Edit → Username"
	adminProfilesMenuPublishHeader       = "Profile Manager → Publish"
	adminProfilesMenuCoffeeBanHeader     = "Profile Manager → Coffee Meetings Ban"
)

type adminProfilesHandler struct {
	config               *config.Config
	messageSenderService *services.MessageSenderService
	permissionsService   *services.PermissionsService
	profileService       *services.ProfileService
	userRepository       *repositories.UserRepository
	profileRepository    *repositories.ProfileRepository
	userStore            *utils.UserDataStore
}

func NewAdminProfilesHandler(
	config *config.Config,
	messageSenderService *services.MessageSenderService,
	permissionsService *services.PermissionsService,
	profileService *services.ProfileService,
	userRepository *repositories.UserRepository,
	profileRepository *repositories.ProfileRepository,
) ext.Handler {
	h := &adminProfilesHandler{
		config:               config,
		messageSenderService: messageSenderService,
		permissionsService:   permissionsService,
		profileService:       profileService,
		userRepository:       userRepository,
		profileRepository:    profileRepository,
		userStore:            utils.NewUserDataStore(),
	}

	return handlers.NewConversation(
		[]ext.Handler{
			handlers.NewCommand(constants.AdminProfilesCommand, h.handleCommand),
		},
		map[string][]ext.Handler{
			adminProfilesStateStart: {
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesSearchByUsernameCallback), h.handleSearchByUsernameCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCreateByForwardedMessageCallback), h.handleCreateByForwardedMessageCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCreateByTelegramIDCallback), h.handleCreateByTelegramIDCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesSearchByTelegramIDCallback), h.handleSearchByTelegramIDCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesSearchByFullNameCallback), h.handleSearchByFullNameCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitSearchByUsername: {
				handlers.NewMessage(message.Text, h.handleSearchByUsernameInput),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesStartCallback), h.handleStartCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitSearchUserID: {
				handlers.NewMessage(message.Text, h.handleSearchUserIDInput),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesStartCallback), h.handleStartCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitSearchByFullName: {
				handlers.NewMessage(message.Text, h.handleSearchByFullNameInput),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesStartCallback), h.handleStartCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitCreateByForwardedMessage: {
				handlers.NewMessage(message.All, h.handleCreateByForwardedMessageInput),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesStartCallback), h.handleStartCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitCreateByTelegramID: {
				handlers.NewMessage(message.Text, h.handleCreateByTelegramIDInput),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesStartCallback), h.handleStartCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateEditProfile: {
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditBioCallback), h.handleEditFieldCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditFirstnameCallback), h.handleEditFieldCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditLastnameCallback), h.handleEditFieldCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditUsernameCallback), h.handleEditFieldCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditCoffeeBanCallback), h.handleEditFieldCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditMenuCallback), h.handleEditMenuCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesPublishCallback), h.handlePublishCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesPublishNoPreviewCallback), h.handlePublishNoPreviewCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesStartCallback), h.handleStartCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitBio: {
				handlers.NewMessage(message.Text, h.handleBioInput),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditMenuCallback), h.handleEditMenuCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitFirstname: {
				handlers.NewMessage(message.Text, h.handleFirstnameInput),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditMenuCallback), h.handleEditMenuCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitLastname: {
				handlers.NewMessage(message.Text, h.handleLastnameInput),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditMenuCallback), h.handleEditMenuCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitUsername: {
				handlers.NewMessage(message.Text, h.handleUsernameInput),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditMenuCallback), h.handleEditMenuCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
			adminProfilesStateAwaitCoffeeBan: {
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesToggleCoffeeBanCallback), h.handleToggleCoffeeBanCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesEditMenuCallback), h.handleEditMenuCallback),
				handlers.NewCallback(callbackquery.Equal(constants.AdminProfilesCancelCallback), h.handleCancelCallback),
			},
		},
		&handlers.ConversationOpts{
			Exits: []ext.Handler{handlers.NewCommand(constants.CancelCommand, h.handleCancel)},
			Fallbacks: []ext.Handler{
				handlers.NewMessage(message.Text, func(b *gotgbot.Bot, ctx *ext.Context) error {
					// Delete the message that not matched any state
					b.DeleteMessage(ctx.EffectiveMessage.Chat.Id, ctx.EffectiveMessage.MessageId, nil)
					return nil
				}),
			},
		},
	)
}

// Entry point for the /profiles command
func (h *adminProfilesHandler) handleCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	// Check if user has admin permissions and is in a private chat
	if !h.permissionsService.CheckAdminAndPrivateChat(msg, constants.AdminProfilesCommand) {
		log.Printf("%s: User %d (%s) tried to use /%s without admin permissions.",
			utils.GetCurrentTypeName(),
			ctx.EffectiveUser.Id,
			ctx.EffectiveUser.Username,
			constants.AdminProfilesCommand,
		)
		return handlers.EndConversation()
	}

	return h.showMainMenu(b, ctx.EffectiveMessage, ctx.EffectiveUser.Id)
}

// Handle the "Start" button click - goes back to the main menu
func (h *adminProfilesHandler) handleStartCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.showMainMenu(b, ctx.EffectiveMessage, ctx.EffectiveUser.Id)
}

// Shows the main profiles menu for admin
func (h *adminProfilesHandler) showMainMenu(b *gotgbot.Bot, msg *gotgbot.Message, userId int64) error {
	h.RemovePreviousMessage(b, &userId)

	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		fmt.Sprintf("<b>%s</b>", adminProfilesMenuHeader)+
			"\n\nHere you can edit user profiles or create a new profile based on a forwarded message.",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.ProfilesMainMenuButtons(),
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send message in showMainMenu: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return handlers.NextConversationState(adminProfilesStateStart)
}

// Handle the "Search by Telegram Username" button click
func (h *adminProfilesHandler) handleSearchByUsernameCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id

	h.RemovePreviousMessage(b, &userId)
	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		fmt.Sprintf("<b>%s</b>", adminProfilesMenuHeader)+
			"\n\nEnter the username (with or without @) to search:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send message in handleCallbackEdit: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return handlers.NextConversationState(adminProfilesStateAwaitSearchByUsername)
}

// Handle the "Search by Telegram ID" button click
func (h *adminProfilesHandler) handleSearchByTelegramIDCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id

	h.RemovePreviousMessage(b, &userId)
	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		fmt.Sprintf("<b>%s</b>", adminProfilesMenuSearchByIDHeader)+
			"\n\nEnter the Telegram user ID to search for a profile:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send message in handleSearchByIDCallback: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return handlers.NextConversationState(adminProfilesStateAwaitSearchUserID)
}

// Handle the "Search by Name" button click
func (h *adminProfilesHandler) handleSearchByFullNameCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id

	h.RemovePreviousMessage(b, &userId)
	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		fmt.Sprintf("<b>%s</b>", adminProfilesMenuSearchByNameHeader)+
			"\n\nEnter the user's first and last name (separated by a space) to search for a profile:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send message in handleSearchByNameCallback: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return handlers.NextConversationState(adminProfilesStateAwaitSearchByFullName)
}

// Handle the "Create profile" button click
func (h *adminProfilesHandler) handleCreateByForwardedMessageCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id

	h.RemovePreviousMessage(b, &userId)
	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		fmt.Sprintf("<b>%s</b>", adminProfilesMenuHeader)+
			"\n\nForward me a message from the user for whom you want to create a profile:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send message in handleCallbackCreate: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return handlers.NextConversationState(adminProfilesStateAwaitCreateByForwardedMessage)
}

// Handle the "Create profile by ID" button click
func (h *adminProfilesHandler) handleCreateByTelegramIDCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id

	h.RemovePreviousMessage(b, &userId)
	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		fmt.Sprintf("<b>%s</b>", adminProfilesMenuCreateByIDHeader)+
			"\n\nEnter the Telegram user ID to create a profile:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send message in handleCreateByIDCallback: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return handlers.NextConversationState(adminProfilesStateAwaitCreateByTelegramID)
}

// Handle the userID input for profile creation
func (h *adminProfilesHandler) handleCreateByTelegramIDInput(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userIDStr := msg.Text
	userId := ctx.EffectiveUser.Id

	// Convert user ID string to int64
	telegramID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuCreateByIDHeader)+
				fmt.Sprintf("\n\nInvalid ID format: <b>%s</b>. Enter a numeric Telegram user ID:", userIDStr),
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
			})
		if err != nil {
			return fmt.Errorf("%s: failed to send message in handleUserIDInput: %w", utils.GetCurrentTypeName(), err)
		}

		h.SavePreviousMessageInfo(userId, editedMsg)
		return nil
	}

	// Check if user exists, create if not
	dbUser, err := h.userRepository.GetByTelegramID(telegramID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("%s: failed to get user in handleUserIDInput: %w", utils.GetCurrentTypeName(), err)
	}

	// If user not found, create a new user with minimal info
	if err == sql.ErrNoRows {
		userID, err := h.userRepository.Create(telegramID, "", "", "")
		if err != nil {
			return fmt.Errorf("%s: failed to create user in handleUserIDInput: %w", utils.GetCurrentTypeName(), err)
		}

		dbUser, err = h.userRepository.GetByID(userID)
		if err != nil {
			return fmt.Errorf("%s: failed to get created user in handleUserIDInput: %w", utils.GetCurrentTypeName(), err)
		}
	}

	// Store the user ID for future use
	h.userStore.Set(userId, adminProfilesCtxDataKeyUserID, dbUser.ID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramID, dbUser.TgID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramUsername, dbUser.TgUsername)

	// Find or create the profile
	profile, err := h.profileRepository.GetOrCreate(dbUser.ID)
	if err != nil {
		_ = h.messageSenderService.Reply(msg,
			"An error occurred while retrieving or creating the profile.", nil)
		return fmt.Errorf("%s: failed to get/create profile in handleUserIDInput: %w", utils.GetCurrentTypeName(), err)
	}

	h.userStore.Set(userId, adminProfilesCtxDataKeyProfileID, profile.ID)

	// Delete the input message
	b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
	h.RemovePreviousMessage(b, &userId)

	// Show the profile edit menu
	return h.showProfileEditMenu(b, msg, userId, dbUser, profile)
}

// Handle the username input for profile search
func (h *adminProfilesHandler) handleSearchByUsernameInput(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	username := msg.Text
	userId := ctx.EffectiveUser.Id

	// Remove @ if present
	username = strings.TrimPrefix(username, "@")

	dbUser, err := h.userRepository.GetByTelegramUsername(username)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("%s: failed to get user in handleUsernameInput: %w", utils.GetCurrentTypeName(), err)
	}

	// If user not found, show search again
	if err == sql.ErrNoRows {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuHeader)+
				fmt.Sprintf("\n\nUser <b>%s</b> not found.", username)+
				"\n\nTry again, or go back:",
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
			})
		if err != nil {
			return fmt.Errorf("%s: failed to send message in handleUsernameInput: %w", utils.GetCurrentTypeName(), err)
		}

		h.SavePreviousMessageInfo(userId, editedMsg)
		return nil
	}

	// Store the user ID for future use
	h.userStore.Set(userId, adminProfilesCtxDataKeyUserID, dbUser.ID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramID, dbUser.TgID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramUsername, dbUser.TgUsername)

	// Find or create the profile
	profile, err := h.profileRepository.GetOrCreate(dbUser.ID)
	if err != nil {
		_ = h.messageSenderService.Reply(msg,
			"An error occurred while retrieving or creating the profile.", nil)
		return fmt.Errorf("%s: failed to get/create profile in handleUsernameInput: %w", utils.GetCurrentTypeName(), err)
	}

	h.userStore.Set(userId, adminProfilesCtxDataKeyProfileID, profile.ID)

	// Delete the input message
	b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
	h.RemovePreviousMessage(b, &userId)

	// Show the profile edit menu
	return h.showProfileEditMenu(b, msg, userId, dbUser, profile)
}

// Handle forwarded message for profile creation
func (h *adminProfilesHandler) handleCreateByForwardedMessageInput(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id
	msgType := msg.ForwardOrigin.GetType()

	// Check if this is a forwarded message with user origin
	if msg.ForwardOrigin == nil || msgType != "user" {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		msgText := fmt.Sprintf("<b>%s</b>", adminProfilesMenuHeader) +
			"\n\nThis is not a forwarded message from a user. Please forward a message from the user for whom you want to create a profile:"
		if msgType == "hidden_user" {
			msgText += "\n\n<i>A message from a hidden user cannot be used to create a profile.</i>"
		}
		editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			msgText,
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
			})

		if err != nil {
			return fmt.Errorf("%s: failed to send message in handleForwardedMessage: %w", utils.GetCurrentTypeName(), err)
		}

		h.SavePreviousMessageInfo(userId, editedMsg)
		return nil // Stay in the current state
	}

	// Cast ForwardOrigin to MessageOriginUser to get user info
	forwardedUser := msg.ForwardOrigin.MergeMessageOrigin().SenderUser

	// Get the user from the database if exists, or create a new one
	dbUser, err := h.userRepository.GetOrCreate(forwardedUser)
	if err != nil {
		_ = h.messageSenderService.Reply(msg,
			"An error occurred while creating the user.", nil)
		return fmt.Errorf("%s: failed to create user in handleForwardedMessage: %w", utils.GetCurrentTypeName(), err)
	}

	// Store the user ID for future use
	h.userStore.Set(userId, adminProfilesCtxDataKeyUserID, dbUser.ID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramID, dbUser.TgID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramUsername, dbUser.TgUsername)

	// Find or create the profile
	profile, err := h.profileRepository.GetOrCreateWithBio(dbUser.ID, msg.Text)
	if err != nil {
		_ = h.messageSenderService.Reply(msg,
			"An error occurred while creating the profile.", nil)
		return fmt.Errorf("%s: failed to create profile in handleForwardedMessage: %w", utils.GetCurrentTypeName(), err)
	}

	h.userStore.Set(userId, adminProfilesCtxDataKeyProfileID, profile.ID)

	// Delete the forwarded message
	b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
	h.RemovePreviousMessage(b, &userId)

	// Show the profile edit menu
	return h.showProfileEditMenu(b, msg, userId, dbUser, profile)
}

// Handle the "Start" button click - goes back to the main menu
func (h *adminProfilesHandler) handleEditMenuCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id

	h.RemovePreviousMessage(b, &userId)
	userIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyUserID)
	if !ok {
		return fmt.Errorf("%s: user ID not found in user store", utils.GetCurrentTypeName())
	}
	dbUserID := userIDVal.(int)

	user, err := h.userRepository.GetByID(dbUserID)
	if err != nil {
		return fmt.Errorf("%s: failed to get user in handleEditMenuCallback: %w", utils.GetCurrentTypeName(), err)
	}

	profile, err := h.profileRepository.GetOrCreate(dbUserID)
	if err != nil {
		return fmt.Errorf("%s: failed to get profile in handleEditMenuCallback: %w", utils.GetCurrentTypeName(), err)
	}

	return h.showProfileEditMenu(b, msg, userId, user, profile)
}

// Handle the user ID input for profile search
func (h *adminProfilesHandler) handleSearchUserIDInput(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userIDStr := msg.Text
	userId := ctx.EffectiveUser.Id

	// Convert user ID string to int64
	telegramID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuSearchByIDHeader)+
				fmt.Sprintf("\n\nInvalid ID format: <b>%s</b>. Enter a numeric Telegram user ID:", userIDStr),
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
			})
		if err != nil {
			return fmt.Errorf("%s: failed to send message in handleSearchUserIDInput: %w", utils.GetCurrentTypeName(), err)
		}

		h.SavePreviousMessageInfo(userId, editedMsg)
		return nil
	}

	// Check if user exists
	dbUser, err := h.userRepository.GetByTelegramID(telegramID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("%s: failed to get user in handleSearchUserIDInput: %w", utils.GetCurrentTypeName(), err)
	}

	// If user not found, show search again
	if err == sql.ErrNoRows {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuSearchByIDHeader)+
				fmt.Sprintf("\n\nUser with ID <b>%d</b> not found.", telegramID)+
				"\n\nTry again, or go back:",
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
			})
		if err != nil {
			return fmt.Errorf("%s: failed to send message in handleSearchUserIDInput: %w", utils.GetCurrentTypeName(), err)
		}

		h.SavePreviousMessageInfo(userId, editedMsg)
		return nil
	}

	// Store the user ID for future use
	h.userStore.Set(userId, adminProfilesCtxDataKeyUserID, dbUser.ID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramID, dbUser.TgID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramUsername, dbUser.TgUsername)

	// Find or create the profile
	profile, err := h.profileRepository.GetOrCreate(dbUser.ID)
	if err != nil {
		_ = h.messageSenderService.Reply(msg,
			"An error occurred while retrieving the profile.", nil)
		return fmt.Errorf("%s: failed to get profile in handleSearchUserIDInput: %w", utils.GetCurrentTypeName(), err)
	}

	h.userStore.Set(userId, adminProfilesCtxDataKeyProfileID, profile.ID)

	// Delete the input message
	b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
	h.RemovePreviousMessage(b, &userId)

	// Show the profile edit menu
	return h.showProfileEditMenu(b, msg, userId, dbUser, profile)
}

// Handle the name input for profile search
func (h *adminProfilesHandler) handleSearchByFullNameInput(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	fullName := msg.Text
	userId := ctx.EffectiveUser.Id

	// Split into first and last name
	parts := strings.Fields(fullName)
	if len(parts) < 2 {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuSearchByNameHeader)+
				fmt.Sprintf("\n\nInvalid name format: <b>%s</b>.", fullName)+
				"\n\nPlease enter the user's first and last name separated by a space:",
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
			})
		if err != nil {
			return fmt.Errorf("%s: failed to send message in handleSearchByFullNameInput: %w", utils.GetCurrentTypeName(), err)
		}

		h.SavePreviousMessageInfo(userId, editedMsg)
		return nil
	}

	firstname := parts[0]
	lastname := strings.Join(parts[1:], " ")

	// Search for the user by first and last name
	user, err := h.userRepository.SearchByName(firstname, lastname)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("%s: failed to search user in handleSearchNameInput: %w", utils.GetCurrentTypeName(), err)
	}

	// If user not found, show search again
	if err == sql.ErrNoRows {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuSearchByNameHeader)+
				fmt.Sprintf("\n\nUser with name <b>%s %s</b> not found.", firstname, lastname)+
				"\n\nTry again, or go back:",
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesStartCallback),
			})
		if err != nil {
			return fmt.Errorf("%s: failed to send message in handleSearchNameInput: %w", utils.GetCurrentTypeName(), err)
		}

		h.SavePreviousMessageInfo(userId, editedMsg)
		return nil
	}

	// Store the user ID for future use
	h.userStore.Set(userId, adminProfilesCtxDataKeyUserID, user.ID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramID, user.TgID)
	h.userStore.Set(userId, adminProfilesCtxDataKeyTelegramUsername, user.TgUsername)

	// Find or create the profile
	profile, err := h.profileRepository.GetOrCreate(user.ID)
	if err != nil {
		_ = h.messageSenderService.Reply(msg,
			"An error occurred while retrieving the profile.", nil)
		return fmt.Errorf("%s: failed to get profile in handleSearchNameInput: %w", utils.GetCurrentTypeName(), err)
	}

	h.userStore.Set(userId, adminProfilesCtxDataKeyProfileID, profile.ID)

	// Delete the input message
	b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
	h.RemovePreviousMessage(b, &userId)

	// Show the profile edit menu
	return h.showProfileEditMenu(b, msg, userId, user, profile)
}

// Shows the profile edit menu
func (h *adminProfilesHandler) showProfileEditMenu(b *gotgbot.Bot, msg *gotgbot.Message, userId int64, user *repositories.User, profile *repositories.Profile) error {
	profileText := fmt.Sprintf("<b>%s</b>\n\n%s", adminProfilesMenuEditHeader, formatters.FormatProfileManagerView(user, profile, user.HasCoffeeBan, h.config))

	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		profileText,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.ProfilesEditMenuButtons(constants.AdminProfilesStartCallback),
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send message in showProfileEditMenu: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return handlers.NextConversationState(adminProfilesStateEditProfile)
}

// Handle button clicks for editing different profile fields
func (h *adminProfilesHandler) handleEditFieldCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	callback := ctx.Update.CallbackQuery
	data := callback.Data
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id

	// Get stored user and profile IDs
	userIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyUserID)
	if !ok {
		return fmt.Errorf("%s: user ID not found in user store", utils.GetCurrentTypeName())
	}
	dbUserID := userIDVal.(int)

	// Get the user from database
	dbUser, err := h.userRepository.GetByID(dbUserID)
	if err != nil {
		return fmt.Errorf("%s: failed to get user in handleCallbackEditField: %w", utils.GetCurrentTypeName(), err)
	}

	// Get the profile from database
	profileIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyProfileID)
	if !ok {
		return fmt.Errorf("%s: profile ID not found in user store", utils.GetCurrentTypeName())
	}
	profileID := profileIDVal.(int)

	profile, err := h.profileRepository.GetByID(profileID)
	if err != nil {
		return fmt.Errorf("%s: failed to get profile in handleCallbackEditField: %w", utils.GetCurrentTypeName(), err)
	}

	var callToAction string
	var menuHeader string
	var nextState string
	var oldField string

	// Determine which field is being edited
	switch data {
	case constants.AdminProfilesEditFirstnameCallback:
		callToAction = "Enter a new value for the <b>first name</b> field"
		menuHeader = adminProfilesMenuEditFirstnameHeader
		nextState = adminProfilesStateAwaitFirstname
		oldField = "Current value: <code>" + dbUser.Firstname + "</code>"
	case constants.AdminProfilesEditLastnameCallback:
		callToAction = "Enter a new value for the <b>last name</b> field"
		menuHeader = adminProfilesMenuEditLastnameHeader
		nextState = adminProfilesStateAwaitLastname
		oldField = "Current value: <code>" + dbUser.Lastname + "</code>"
	case constants.AdminProfilesEditUsernameCallback:
		callToAction = "Enter a new value for the <b>username</b> field (without @)"
		menuHeader = adminProfilesMenuEditUsernameHeader
		nextState = adminProfilesStateAwaitUsername
		oldField = "Current value: <code>" + dbUser.TgUsername + "</code>"
	case constants.AdminProfilesEditBioCallback:
		callToAction = fmt.Sprintf("Enter a new value for the <b>bio</b> field (up to %d characters)", constants.ProfileBioLengthLimit)
		menuHeader = adminProfilesMenuEditBioHeader
		nextState = adminProfilesStateAwaitBio
		profile.Bio = strings.ReplaceAll(profile.Bio, "<", "&lt;")
		profile.Bio = strings.ReplaceAll(profile.Bio, ">", "&gt;")
		oldField = "Current value: <pre>" + profile.Bio + "</pre>"
	case constants.AdminProfilesEditCoffeeBanCallback:
		callToAction = "Click the button to change the coffee meetings status"
		menuHeader = adminProfilesMenuCoffeeBanHeader
		nextState = adminProfilesStateAwaitCoffeeBan
		if dbUser.HasCoffeeBan {
			oldField = "Current value: ❌ Banned"
		} else {
			oldField = "Current value: ✅ Allowed"
		}
	default:
		return fmt.Errorf("%s: unknown callback data: %s", utils.GetCurrentTypeName(), data)
	}

	// Store field being edited for use in input handlers
	h.userStore.Set(userId, adminProfilesCtxDataKeyField, data)

	h.RemovePreviousMessage(b, &userId)

	var replyMarkup gotgbot.InlineKeyboardMarkup
	if data == constants.AdminProfilesEditCoffeeBanCallback {
		// For coffee ban, use toggle buttons
		replyMarkup = buttons.ProfilesCoffeeBanButtons(constants.AdminProfilesEditMenuCallback, dbUser.HasCoffeeBan)
	} else {
		// For other fields, use standard back/cancel buttons
		replyMarkup = buttons.ProfilesBackCancelButtons(constants.AdminProfilesEditMenuCallback)
	}

	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		fmt.Sprintf("<b>%s</b>", menuHeader)+
			fmt.Sprintf("\n\n%s", oldField)+
			fmt.Sprintf("\n\n%s", callToAction),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: replyMarkup,
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send message in handleCallbackEditField: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return handlers.NextConversationState(nextState)
}

// Handle publishing a profile to intro topic (with preview)
func (h *adminProfilesHandler) handlePublishCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.handlePublishProfile(b, ctx, false)
}

// Handle publishing a profile to intro topic without link preview
func (h *adminProfilesHandler) handlePublishNoPreviewCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.handlePublishProfile(b, ctx, true)
}

// Handle publishing a profile to intro topic
func (h *adminProfilesHandler) handlePublishProfile(b *gotgbot.Bot, ctx *ext.Context, withoutPreview bool) error {
	userId := ctx.EffectiveUser.Id
	msg := ctx.EffectiveMessage

	// Get user data
	userIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyUserID)
	if !ok {
		return fmt.Errorf("%s: user ID not found in user store", utils.GetCurrentTypeName())
	}
	dbUserID := userIDVal.(int)

	// Get user from database
	dbUser, err := h.userRepository.GetByID(dbUserID)
	if err != nil {
		return fmt.Errorf("%s: failed to get user in handlePublishProfile: %w", utils.GetCurrentTypeName(), err)
	}

	// Get profile from database
	profileIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyProfileID)
	if !ok {
		return fmt.Errorf("%s: profile ID not found in user store", utils.GetCurrentTypeName())
	}
	profileID := profileIDVal.(int)

	profile, err := h.profileRepository.GetByID(profileID)
	if err != nil {
		return fmt.Errorf("%s: failed to get profile in handlePublishProfile: %w", utils.GetCurrentTypeName(), err)
	}

	firstNameString := "└ ❌ First Name"
	lastNameString := "└ ❌ Last Name"
	bioString := "└ ❌ Bio"
	if dbUser != nil {
		if dbUser.Firstname != "" {
			firstNameString = "└ ✅ First Name"
		}
		if dbUser.Lastname != "" {
			lastNameString = "└ ✅ Last Name"
		}
	}

	if profile != nil {
		if profile.Bio != "" {
			bioString = "└ ✅ Bio"
		}
	}

	if !h.profileService.IsProfileComplete(dbUser, profile) {
		h.RemovePreviousMessage(b, &userId)
		editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuPublishHeader)+
				"\n\n⚠️ The user profile is incomplete. "+
				fmt.Sprintf("\n\nTo publish it in the \"<a href='%s'>Intro</a>\" channel, the following must be provided: ",
					utils.GetIntroTopicLink(h.config))+
				"\n"+firstNameString+
				"\n"+lastNameString+
				"\n"+bioString,
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesEditMenuCallback),
			})

		if err != nil {
			return fmt.Errorf("%s: failed to send message in handlePublishProfile: %w", utils.GetCurrentTypeName(), err)
		}

		h.SavePreviousMessageInfo(userId, editedMsg)
		return nil // Stay in current state
	}

	// Format profile text for publishing
	publicMessageText := formatters.FormatPublicProfileForMessage(dbUser, profile, false)

	var publishedMsg *gotgbot.Message
	// Check if we need to update existing message or create a new one
	if profile.PublishedMessageID.Valid {
		// Try to edit existing message
		_, _, err := b.EditMessageText(
			publicMessageText,
			&gotgbot.EditMessageTextOpts{
				ChatId:    utils.ChatIdToFullChatId(h.config.SuperGroupChatID),
				MessageId: profile.PublishedMessageID.Int64,
				ParseMode: "HTML",
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: withoutPreview,
				},
			})
		// If editing fails, create a new message if the error is not about the message being exactly the same
		if err != nil && !strings.Contains(err.Error(), "are exactly the same") {
			publishedMsg, err = h.messageSenderService.SendHtmlWithReturnMessage(
				utils.ChatIdToFullChatId(h.config.SuperGroupChatID),
				publicMessageText,
				&gotgbot.SendMessageOpts{
					MessageThreadId: int64(h.config.IntroTopicID),
					LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
						IsDisabled: withoutPreview,
					},
				})
			if err != nil {
				return fmt.Errorf("%s: failed to publish profile: %w", utils.GetCurrentTypeName(), err)
			}
		} else {
			// Message updated successfully, store the message ID for database update
			messageID := profile.PublishedMessageID.Int64
			publishedMsg = &gotgbot.Message{
				MessageId: messageID,
			}
		}
	} else {
		// Create a new message
		publishedMsg, err = h.messageSenderService.SendHtmlWithReturnMessage(
			utils.ChatIdToFullChatId(h.config.SuperGroupChatID),
			publicMessageText,
			&gotgbot.SendMessageOpts{
				MessageThreadId: int64(h.config.IntroTopicID),
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: withoutPreview,
				},
			})
		if err != nil {
			return fmt.Errorf("%s: failed to publish profile: %w", utils.GetCurrentTypeName(), err)
		}
	}

	// Update profile with the published message ID
	err = h.profileRepository.UpdatePublishedMessageID(profile.ID, publishedMsg.MessageId)
	if err != nil {
		return fmt.Errorf("%s: failed to update published message ID: %w", utils.GetCurrentTypeName(), err)
	}

	// Show success message
	h.RemovePreviousMessage(b, &userId)
	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		fmt.Sprintf("<b>%s</b>", adminProfilesMenuPublishHeader)+
			fmt.Sprintf("\n\n✅ User profile successfully published in the \"<a href='%s'>Intro</a>\" channel!", utils.GetIntroMessageLink(h.config, profile.PublishedMessageID.Int64)),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.ProfilesBackStartCancelButtons(constants.AdminProfilesEditMenuCallback),
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send success message: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return nil // Stay in current state
}

// Handle bio input
func (h *adminProfilesHandler) handleBioInput(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	bio := msg.Text
	userId := ctx.EffectiveUser.Id
	bioLength := utils.Utf16CodeUnitCount(bio)

	// skip if it is sequential message from user with the same date
	lastMessageDate, ok := h.userStore.Get(msg.From.Id, adminProfilesCtxDataKeyLastMessageTimeFromUser)
	if ok && lastMessageDate == msg.Date {
		// Skip processing - same message date detected
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		return nil
	}

	// Store current message date to avoid duplicate processing
	h.userStore.Set(msg.From.Id, adminProfilesCtxDataKeyLastMessageTimeFromUser, msg.Date)

	if bioLength > constants.ProfileBioLengthLimit {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		errMsg, _ := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuEditBioHeader)+
				fmt.Sprintf("\n\nCurrent length: %d characters", bioLength)+
				fmt.Sprintf("\n\nPlease shorten to %d characters and send again:", constants.ProfileBioLengthLimit),
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesEditMenuCallback),
			})

		h.SavePreviousMessageInfo(userId, errMsg)
		return nil // Stay in current state
	}

	// Get profile ID from store
	profileIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyProfileID)
	if !ok {
		return fmt.Errorf("%s: profile ID not found in user store", utils.GetCurrentTypeName())
	}
	profileID := profileIDVal.(int)

	// Save the bio
	err := h.profileRepository.Update(profileID, map[string]interface{}{
		"bio": bio,
	})
	if err != nil {
		return fmt.Errorf("%s: failed to update bio: %w", utils.GetCurrentTypeName(), err)
	}

	return h.returnToProfileView(b, ctx)
}

// Handle firstname input
func (h *adminProfilesHandler) handleFirstnameInput(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	firstname := msg.Text
	userId := ctx.EffectiveUser.Id

	if len(firstname) > 30 {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		errMsg, _ := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuEditFirstnameHeader)+
				"\n\nFirst name is too long. Please enter a shorter first name (no more than 30 characters):",
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesEditMenuCallback),
			})

		h.SavePreviousMessageInfo(userId, errMsg)
		return nil // Stay in current state
	}

	// Get user ID from store
	userIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyUserID)
	if !ok {
		return fmt.Errorf("%s: user ID not found in user store", utils.GetCurrentTypeName())
	}
	dbUserID := userIDVal.(int)

	// Save the firstname
	err := h.userRepository.Update(dbUserID, map[string]interface{}{
		"firstname": firstname,
	})
	if err != nil {
		return fmt.Errorf("%s: failed to update firstname: %w", utils.GetCurrentTypeName(), err)
	}

	return h.returnToProfileView(b, ctx)
}

// Handle lastname input
func (h *adminProfilesHandler) handleLastnameInput(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	lastname := msg.Text
	userId := ctx.EffectiveUser.Id

	if len(lastname) > 30 {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		errMsg, _ := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuEditLastnameHeader)+
				"\n\nLast name is too long. Please enter a shorter last name (no more than 30 characters):",
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesEditMenuCallback),
			})

		h.SavePreviousMessageInfo(userId, errMsg)
		return nil // Stay in current state
	}

	// Get user ID from store
	userIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyUserID)
	if !ok {
		return fmt.Errorf("%s: user ID not found in user store", utils.GetCurrentTypeName())
	}
	dbUserID := userIDVal.(int)

	// Save the lastname
	err := h.userRepository.Update(dbUserID, map[string]interface{}{
		"lastname": lastname,
	})
	if err != nil {
		return fmt.Errorf("%s: failed to update lastname: %w", utils.GetCurrentTypeName(), err)
	}

	return h.returnToProfileView(b, ctx)
}

// Handle username input
func (h *adminProfilesHandler) handleUsernameInput(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	username := msg.Text
	userId := ctx.EffectiveUser.Id

	// Remove @ prefix if present
	if len(username) > 0 && username[0] == '@' {
		username = username[1:]
	}

	if len(username) > 32 {
		h.RemovePreviousMessage(b, &userId)
		b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		errMsg, _ := h.messageSenderService.SendHtmlWithReturnMessage(
			msg.Chat.Id,
			fmt.Sprintf("<b>%s</b>", adminProfilesMenuEditUsernameHeader)+
				"\n\nUsername is too long. Please enter a shorter username (no more than 32 characters):",
			&gotgbot.SendMessageOpts{
				ReplyMarkup: buttons.ProfilesBackCancelButtons(constants.AdminProfilesEditMenuCallback),
			})

		h.SavePreviousMessageInfo(userId, errMsg)
		return nil // Stay in current state
	}

	// Get user ID from store
	userIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyUserID)
	if !ok {
		return fmt.Errorf("%s: user ID not found in user store", utils.GetCurrentTypeName())
	}
	dbUserID := userIDVal.(int)

	// Save the username
	err := h.userRepository.Update(dbUserID, map[string]interface{}{
		"tg_username": username,
	})
	if err != nil {
		return fmt.Errorf("%s: failed to update username: %w", utils.GetCurrentTypeName(), err)
	}

	return h.returnToProfileView(b, ctx)
}

// Helper function to return to profile view after an update
func (h *adminProfilesHandler) returnToProfileView(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id

	// Clean input message
	b.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
	h.RemovePreviousMessage(b, &userId)

	// Get user data
	userIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyUserID)
	if !ok {
		return fmt.Errorf("%s: user ID not found in user store", utils.GetCurrentTypeName())
	}
	dbUserID := userIDVal.(int)

	// Get user from database
	dbUser, err := h.userRepository.GetByID(dbUserID)
	if err != nil {
		return fmt.Errorf("%s: failed to get user in returnToProfileView: %w", utils.GetCurrentTypeName(), err)
	}

	// Get profile from database
	profileIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyProfileID)
	if !ok {
		return fmt.Errorf("%s: profile ID not found in user store", utils.GetCurrentTypeName())
	}
	profileID := profileIDVal.(int)

	profile, err := h.profileRepository.GetByID(profileID)
	if err != nil {
		return fmt.Errorf("%s: failed to get profile in returnToProfileView: %w", utils.GetCurrentTypeName(), err)
	}

	successMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		"✅ Value saved successfully!",
		nil)

	if err != nil {
		return fmt.Errorf("%s: failed to send success message: %w", utils.GetCurrentTypeName(), err)
	}

	// Show updated profile after a brief delay
	time.Sleep(1 * time.Second)
	b.DeleteMessage(msg.Chat.Id, successMsg.MessageId, nil)

	// Show profile edit menu with updated data
	return h.showProfileEditMenu(b, msg, userId, dbUser, profile)
}

// Handle the toggle coffee ban button click
func (h *adminProfilesHandler) handleToggleCoffeeBanCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	userId := ctx.EffectiveUser.Id
	msg := ctx.EffectiveMessage

	// Get user data
	userIDVal, ok := h.userStore.Get(userId, adminProfilesCtxDataKeyUserID)
	if !ok {
		return fmt.Errorf("%s: user ID not found in user store", utils.GetCurrentTypeName())
	}
	dbUserID := userIDVal.(int)

	// Get user from database
	dbUser, err := h.userRepository.GetByID(dbUserID)
	if err != nil {
		return fmt.Errorf("%s: failed to get user in handleToggleCoffeeBanCallback: %w", utils.GetCurrentTypeName(), err)
	}

	// Toggle the coffee ban status
	newStatus := !dbUser.HasCoffeeBan
	err = h.userRepository.SetCoffeeBan(dbUserID, newStatus)
	if err != nil {
		return fmt.Errorf("%s: failed to update coffee ban status: %w", utils.GetCurrentTypeName(), err)
	}

	// Update the message with new buttons
	h.RemovePreviousMessage(b, &userId)

	var statusText string
	if newStatus {
		statusText = "❌ Banned"
	} else {
		statusText = "✅ Allowed"
	}

	editedMsg, err := h.messageSenderService.SendHtmlWithReturnMessage(
		msg.Chat.Id,
		fmt.Sprintf("<b>%s</b>", adminProfilesMenuCoffeeBanHeader)+
			fmt.Sprintf("\n\nCurrent value: %s", statusText)+
			"\n\nEnter a new value for the <b>coffee meetings status</b> field:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: buttons.ProfilesCoffeeBanButtons(constants.AdminProfilesEditMenuCallback, newStatus),
		})

	if err != nil {
		return fmt.Errorf("%s: failed to send message in handleToggleCoffeeBanCallback: %w", utils.GetCurrentTypeName(), err)
	}

	h.SavePreviousMessageInfo(userId, editedMsg)
	return nil // Stay in current state
}

func (h *adminProfilesHandler) handleCancelCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.handleCancel(b, ctx)
}

func (h *adminProfilesHandler) handleCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	userId := ctx.EffectiveUser.Id

	h.RemovePreviousMessage(b, &userId)
	_ = h.messageSenderService.Send(
		msg.Chat.Id, "Admin profile management session ended.", nil)
	h.userStore.Clear(ctx.EffectiveUser.Id)

	return handlers.EndConversation()
}

func (h *adminProfilesHandler) MessageRemoveInlineKeyboard(b *gotgbot.Bot, userID *int64) {
	var chatID, messageID int64

	if userID != nil {
		messageID, chatID = h.userStore.GetPreviousMessageInfo(
			*userID,
			adminProfilesCtxDataKeyPreviousMessageID,
			adminProfilesCtxDataKeyPreviousChatID,
		)
	}

	if chatID == 0 || messageID == 0 {
		return
	}

	_ = h.messageSenderService.RemoveInlineKeyboard(chatID, messageID)
}

func (h *adminProfilesHandler) RemovePreviousMessage(b *gotgbot.Bot, userID *int64) {
	var chatID, messageID int64

	if userID != nil {
		messageID, chatID = h.userStore.GetPreviousMessageInfo(
			*userID,
			adminProfilesCtxDataKeyPreviousMessageID,
			adminProfilesCtxDataKeyPreviousChatID,
		)
	}

	if chatID == 0 || messageID == 0 {
		return
	}

	b.DeleteMessage(chatID, messageID, nil)
}

func (h *adminProfilesHandler) SavePreviousMessageInfo(userID int64, sentMsg *gotgbot.Message) {
	if sentMsg == nil {
		return
	}
	h.userStore.SetPreviousMessageInfo(userID, sentMsg.MessageId, sentMsg.Chat.Id,
		adminProfilesCtxDataKeyPreviousMessageID, adminProfilesCtxDataKeyPreviousChatID)
}
