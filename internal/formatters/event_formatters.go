package formatters

import (
	"fmt"
	"strings"
	"time"

	"evo-bot-go/internal/constants"
	"evo-bot-go/internal/database/repositories"
)

func GetTypeEmoji(eventType constants.EventType) string {
	switch eventType {
	case constants.EventTypeClubCall:
		return "\U0001f4ac"
	case constants.EventTypeMeetup:
		return "\U0001f399"
	case constants.EventTypeWorkshop:
		return "\u2699\ufe0f"
	case constants.EventTypeReadingClub:
		return "\U0001f4da"
	case constants.EventTypeConference:
		return "\U0001f465"
	default:
		return "\U0001f504"
	}
}

func GetTypeName(eventType constants.EventType) string {
	switch eventType {
	case constants.EventTypeClubCall:
		return "club call"
	case constants.EventTypeMeetup:
		return "meetup"
	case constants.EventTypeWorkshop:
		return "workshop"
	case constants.EventTypeReadingClub:
		return "reading club"
	case constants.EventTypeConference:
		return "conference"
	default:
		return string(eventType)
	}
}

func GetStatusEmoji(status constants.EventStatus) string {
	switch status {
	case constants.EventStatusFinished:
		return "\u2705"
	case constants.EventStatusActual:
		return "\U0001f504"
	default:
		return "\U0001f504"
	}
}

func FormatEventListForTopicsView(events []repositories.Event, title string) string {
	var response strings.Builder
	response.WriteString(fmt.Sprintf("%s:\n", title))

	for _, event := range events {
		// Handle optional started_at field
		startedAtStr := "not set"
		if event.StartedAt != nil && !event.StartedAt.IsZero() {
			startedAtStr = event.StartedAt.Format("02.01.2006 at 15:04")
		}

		typeEmoji := GetTypeEmoji(constants.EventType(event.Type))
		typeName := GetTypeName(constants.EventType(event.Type))

		response.WriteString(fmt.Sprintf("\n%s _%s_: *%s*\n", typeEmoji, typeName, event.Name))
		response.WriteString(fmt.Sprintf("\u2514   _ID_ /%d, _when_: %s\n",
			event.ID, startedAtStr))
	}

	return response.String()
}

func FormatEventListForEventsView(events []repositories.Event, title string) string {
	var response strings.Builder
	response.WriteString(fmt.Sprintf("%s:\n", title))

	for _, event := range events {
		// Handle optional started_at field
		startedAtStr := "not set"
		if event.StartedAt != nil && !event.StartedAt.IsZero() {
			startedAtStr = event.StartedAt.Format("02.01.2006 at 15:04 UTC")

			// Add time remaining if event is in the future
			utcNow := time.Now().UTC()
			if event.StartedAt.After(utcNow) {
				timeUntil := event.StartedAt.Sub(utcNow)

				switch {
				case timeUntil <= 24*time.Hour:
					// Less than 24 hours
					hours := int(timeUntil.Hours())
					mins := int(timeUntil.Minutes()) % 60
					if hours > 0 {
						startedAtStr += fmt.Sprintf(" _(in %dh %dmin)_", hours, mins)
					} else {
						startedAtStr += fmt.Sprintf(" _(in %dmin)_", mins)
					}
				case timeUntil <= 7*24*time.Hour:
					// Less than 7 days
					days := int(timeUntil.Hours() / 24)
					hours := int(timeUntil.Hours()) % 24
					startedAtStr += fmt.Sprintf(" _(in %dd %dh)_", days, hours)
				}
			}
		}

		typeEmoji := GetTypeEmoji(constants.EventType(event.Type))
		typeName := GetTypeName(constants.EventType(event.Type))

		response.WriteString(fmt.Sprintf("\n%s _%s_: *%s*\n", typeEmoji, typeName, event.Name))
		response.WriteString(fmt.Sprintf("\u2514   _when_: %s\n", startedAtStr))
	}

	return response.String()
}

func FormatEventListForAdmin(events []repositories.Event, title string, cancelCommand string, actionDescription string) string {
	var response strings.Builder
	response.WriteString(fmt.Sprintf("*%s*\n", title))

	for _, event := range events {
		// Handle optional started_at field
		startedAtStr := "not set"
		if event.StartedAt != nil && !event.StartedAt.IsZero() {
			startedAtStr = event.StartedAt.Format("02.01.2006 at 15:04 UTC")
		}

		statusEmoji := GetStatusEmoji(constants.EventStatus(event.Status))
		typeEmoji := GetTypeEmoji(constants.EventType(event.Type))

		response.WriteString(fmt.Sprintf("\n%s ID /%d: *%s*\n", typeEmoji, event.ID, event.Name))
		response.WriteString(fmt.Sprintf("\u2514 %s _when_: *%s*\n",
			statusEmoji, startedAtStr))
	}

	response.WriteString(fmt.Sprintf("\nPlease send the event ID to %s.", actionDescription))

	return response.String()
}

func FormatHtmlTopicListForUsers(topics []repositories.Topic, eventName string, eventType string) string {
	var response strings.Builder

	typeEmoji := GetTypeEmoji(constants.EventType(eventType))
	typeName := GetTypeName(constants.EventType(eventType))

	response.WriteString(fmt.Sprintf("\n %s Event (%s): <b>%s</b>\n", typeEmoji, typeName, eventName))

	if len(topics) == 0 {
		response.WriteString(
			fmt.Sprintf("\n\U0001f50d No topics or questions for this event yet.\n Use /%s to add one.", constants.TopicAddCommand))
	} else {
		topicCount := len(topics)
		response.WriteString(fmt.Sprintf("\U0001f4cb Topics and questions found: <b>%d</b>\n\n", topicCount))

		for i, topic := range topics {
			// Format date as DD.MM.YYYY for better readability
			dateFormatted := topic.CreatedAt.Format("02.01.2006")
			response.WriteString(fmt.Sprintf(
				"<i>%s</i> <blockquote expandable>%s</blockquote>\n",
				dateFormatted,
				topic.Topic,
			))

			// Don't add separator after the last item
			if i < topicCount-1 {
				response.WriteString("\n")
			}
		}

		response.WriteString(
			fmt.Sprintf(
				"\nUse /%s to add new topics and questions, or /%s to view topics for another event.",
				constants.TopicAddCommand,
				constants.TopicsCommand,
			),
		)
	}

	return response.String()
}

func FormatHtmlTopicListForAdmin(topics []repositories.Topic, eventName string, eventType string) string {
	var response strings.Builder

	typeEmoji := GetTypeEmoji(constants.EventType(eventType))
	typeName := GetTypeName(constants.EventType(eventType))

	response.WriteString(fmt.Sprintf("\n %s <i>Event (%s):</i> %s\n\n", typeEmoji, typeName, eventName))

	if len(topics) == 0 {
		response.WriteString("No topics or questions for this event yet.")
	} else {
		for _, topic := range topics {
			userNickname := "not specified"
			if topic.UserNickname != nil {
				userNickname = "@" + *topic.UserNickname
			}
			dateFormatted := topic.CreatedAt.Format("02.01.2006")
			response.WriteString(fmt.Sprintf(
				"ID:<code>%d</code> / <i>%s</i> / %s \n",
				topic.ID,
				dateFormatted,
				userNickname,
			))
			response.WriteString(fmt.Sprintf("<blockquote expandable>%s</blockquote> \n", topic.Topic))
		}
	}

	return response.String()
}
