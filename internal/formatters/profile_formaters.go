package formatters

import (
	"evo-bot-go/internal/config"
	"evo-bot-go/internal/database/repositories"
	"evo-bot-go/internal/utils"
	"fmt"
	"strconv"
	"strings"
)

// Format a readable view of a user profile
func FormatProfileView(user *repositories.User, profile *repositories.Profile, showScore bool) string {
	if profile == nil {
		return "Your profile was not found.\n\nCreate a profile using the \"Edit my profile\" button."
	}

	// Format username
	username := ""
	fullName := user.Firstname
	if user.Lastname != "" {
		fullName += " " + user.Lastname
	}
	fullName = "<b><a href=\"tg://user?id=" + strconv.FormatInt(user.TgID, 10) + "\">" + fullName + "</a></b>"

	if user.TgUsername != "" {
		username = "(@" + user.TgUsername + ")"
	}

	// Build profile text
	text := fmt.Sprintf("\U0001f590 %s %s\n", fullName, username)

	if profile.Bio != "" {
		profile.Bio = strings.ReplaceAll(profile.Bio, "<", "&lt;")
		profile.Bio = strings.ReplaceAll(profile.Bio, ">", "&gt;")
		text += fmt.Sprintf("\n<blockquote>About</blockquote>\n%s\n", profile.Bio)
	}

	if showScore && user.Score > 100 {
		text += fmt.Sprintf("\n<b>%d</b> <i>(what's this? hmm...)</i>\n", user.Score)
	}

	return text
}

// Format a readable view of a user profile for the admin manager
func FormatProfileManagerView(user *repositories.User, profile *repositories.Profile, hasCoffeeBan bool, config *config.Config) string {

	// Format username
	username := ""
	fullName := user.Firstname
	if user.Lastname != "" {
		fullName += " " + user.Lastname
	}
	fullName = "<b><a href=\"tg://user?id=" + strconv.FormatInt(user.TgID, 10) + "\">" + fullName + "</a></b>"

	if user.TgUsername != "" {
		username = "(@" + user.TgUsername + ")"
	}

	// Build profile text
	text := fmt.Sprintf("\U0001f590 %s %s\n", fullName, username)

	if profile.Bio != "" {
		text += "\n<i>About:</i>"
		profile.Bio = strings.ReplaceAll(profile.Bio, "<", "&lt;")
		profile.Bio = strings.ReplaceAll(profile.Bio, ">", "&gt;")
		text += fmt.Sprintf("<blockquote expandable>%s</blockquote>", profile.Bio)
	}
	text += fmt.Sprintf("\n\n<i>Score:</i> <b>%d</b>", user.Score)

	coffeeBanStatus := "\u2705 Allowed"
	if hasCoffeeBan {
		coffeeBanStatus = "\u274c Banned"
	}
	text += fmt.Sprintf("\n<i>Coffee meetings:</i> %s", coffeeBanStatus)
	text += fmt.Sprintf("\n<i>Telegram ID:</i> <code>%d</code>", user.TgID)
	if profile.PublishedMessageID.Valid {
		linkToPost := utils.GetIntroMessageLink(config, profile.PublishedMessageID.Int64)
		text += fmt.Sprintf("\n<i>Profile link:</i> %s", linkToPost)
	}
	return text
}

func FormatPublicProfileForMessage(user *repositories.User, profile *repositories.Profile, showScore bool) string {

	// Format username
	username := ""
	fullName := user.Firstname
	if user.Lastname != "" {
		fullName += " " + user.Lastname
	}
	fullName = "<b><a href=\"tg://user?id=" + strconv.FormatInt(user.TgID, 10) + "\">" + fullName + "</a></b>"

	if user.TgUsername != "" {
		username = "(@" + user.TgUsername + ")"
	}

	// Build profile text
	text := fmt.Sprintf("\U0001f590 %s %s\n", fullName, username)

	if profile.Bio != "" {

		profile.Bio = strings.ReplaceAll(profile.Bio, "<", "&lt;")
		profile.Bio = strings.ReplaceAll(profile.Bio, ">", "&gt;")
		text += fmt.Sprintf("\n<blockquote>About</blockquote>\n%s\n", profile.Bio)
	}

	return text
}
