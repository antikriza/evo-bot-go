package privatehandlers

import (
	"evo-bot-go/internal/config"
	"evo-bot-go/internal/constants"
	"evo-bot-go/internal/database/repositories"
	"evo-bot-go/internal/formatters"
	"evo-bot-go/internal/services"
	"evo-bot-go/internal/utils"
	"fmt"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

type eventsHandler struct {
	config               *config.Config
	eventRepository      *repositories.EventRepository
	messageSenderService *services.MessageSenderService
	permissionsService   *services.PermissionsService
}

func NewEventsHandler(
	config *config.Config,
	eventRepository *repositories.EventRepository,
	messageSenderService *services.MessageSenderService,
	permissionsService *services.PermissionsService,
) ext.Handler {
	h := &eventsHandler{
		config:               config,
		eventRepository:      eventRepository,
		messageSenderService: messageSenderService,
		permissionsService:   permissionsService,
	}

	return handlers.NewCommand(constants.EventsCommand, h.handleCommand)
}

func (h *eventsHandler) handleCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	// Only proceed if this is a private chat
	if !h.permissionsService.CheckPrivateChatType(msg) {
		return nil
	}

	// Check if user is a club member
	if !h.permissionsService.CheckClubMemberPermissions(msg, constants.EventsCommand) {
		return nil
	}

	// Get actual events to show
	events, err := h.eventRepository.GetLastActualEvents(10) // Fetch last 10 actual events
	if err != nil {
		h.messageSenderService.Reply(msg, "Error retrieving the list of events.", nil)
		log.Printf("%s: Error during events retrieval: %v", utils.GetCurrentTypeName(), err)
		return nil
	}

	if len(events) == 0 {
		h.messageSenderService.Reply(msg, "There are no upcoming events at the moment.", nil)
		return nil
	}

	// Format and display event list
	formattedEvents := formatters.FormatEventListForEventsView(
		events,
		"📋 Upcoming Events",
	)
	formattedEvents += fmt.Sprintf("\nAdd topics and questions /%s. ", constants.TopicAddCommand)
	formattedEvents += fmt.Sprintf("View topics and questions /%s. ", constants.TopicsCommand)
	formattedEvents += "For more event information, check the group for updates."
	h.messageSenderService.ReplyMarkdown(msg, formattedEvents, nil)

	return nil
}
