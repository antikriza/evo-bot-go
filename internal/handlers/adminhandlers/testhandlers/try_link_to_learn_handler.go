package testhandlers

import (
	"evo-bot-go/internal/config"
	"evo-bot-go/internal/constants"
	"evo-bot-go/internal/services"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

type tryLinkToLearnHandler struct {
	config             *config.Config
	messageService     *services.MessageSenderService
	permissionsService *services.PermissionsService
}

func NewTryLinkToLearnHandler(
	cfg *config.Config,
	messageService *services.MessageSenderService,
	permissionsService *services.PermissionsService,
) ext.Handler {
	h := &tryLinkToLearnHandler{
		config:             cfg,
		messageService:     messageService,
		permissionsService: permissionsService,
	}

	return handlers.NewCommand(constants.TryLinkToLearnCommand, h.handle)
}

func (h *tryLinkToLearnHandler) handle(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	if !h.permissionsService.CheckAdminAndPrivateChat(msg, constants.TryLinkToLearnCommand) {
		return nil
	}

	return h.messageService.Send(
		msg.Chat.Id,
		"AI & Programming Course (42 topics, 5 levels) \u2192",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
					{
						Text: "\U0001f4d6 Open AI Course",
						Url:  "https://antikriza.github.io/BBD-evolution-code-clone/telegram-archive/course/twa/index.html",
					},
				}},
			},
		},
	)
}
