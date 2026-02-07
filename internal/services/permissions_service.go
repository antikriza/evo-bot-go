package services

import (
	"evo-bot-go/internal/config"
	"evo-bot-go/internal/constants"
	"evo-bot-go/internal/utils"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type PermissionsService struct {
	config               *config.Config
	bot                  *gotgbot.Bot
	messageSenderService *MessageSenderService
}

func NewPermissionsService(
	config *config.Config,
	bot *gotgbot.Bot,
	messageSenderService *MessageSenderService,
) *PermissionsService {
	return &PermissionsService{
		config:               config,
		bot:                  bot,
		messageSenderService: messageSenderService,
	}
}

// CheckAdminPermissions checks if the user has admin permissions and returns an appropriate error response
// Returns true if user has permission, false otherwise
func (s *PermissionsService) CheckAdminPermissions(msg *gotgbot.Message, commandName string) bool {
	if !utils.IsUserAdminOrCreator(s.bot, msg.From.Id, s.config) {
		if err := s.messageSenderService.Reply(
			msg,
			"This command is only available to administrators.",
			nil,
		); err != nil {
			log.Printf("%s: Failed to send admin-only message: %v", utils.GetCurrentTypeName(), err)
		}
		log.Printf("%s: User %d tried to use %s without admin rights", utils.GetCurrentTypeName(), msg.From.Id, commandName)
		return false
	}

	return true
}

// CheckPrivateChatType checks if the command is used in a private chat and returns an appropriate error response
// Returns true if used in private chat, false otherwise
func (s *PermissionsService) CheckPrivateChatType(msg *gotgbot.Message) bool {
	if msg.Chat.Type != constants.PrivateChatType {
		if err := s.messageSenderService.ReplyWithCleanupAfterDelayWithPing(
			msg,
			"*My apologies* 🧐\n\nThis command only works in a _private chat_ with me. "+
				"Send me a DM and I'll be happy to help (I pinged you there if we've chatted before)."+
				"\n\nThis message and your command will be automatically deleted in 10 seconds.",
			10, &gotgbot.SendMessageOpts{
				ParseMode: "Markdown",
			}); err != nil {
			log.Printf("%s: Failed to send private-only message: %v", utils.GetCurrentTypeName(), err)
		}
		return false
	}

	return true
}

func (s *PermissionsService) CheckClubMemberPermissions(msg *gotgbot.Message, commandName string) bool {
	if !utils.IsUserClubMember(s.bot, msg.From.Id, s.config) {
		if err := s.messageSenderService.Reply(
			msg,
			"This command is only available to group members.",
			nil,
		); err != nil {
			log.Printf("%s: Failed to send club-only message: %v", utils.GetCurrentTypeName(), err)
		}
		log.Printf("%s: User %d tried to use %s without club member rights", utils.GetCurrentTypeName(), msg.From.Id, commandName)
		return false
	}

	return true
}

// CheckAdminAndPrivateChat combines permission and chat type checking for admin-only private commands
// Returns true if all checks pass, false otherwise
func (s *PermissionsService) CheckAdminAndPrivateChat(msg *gotgbot.Message, commandName string) bool {
	if !s.CheckAdminPermissions(msg, commandName) {
		return false
	}

	if !s.CheckPrivateChatType(msg) {
		return false
	}

	return true
}
