package discord

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/bwmarrin/discordgo"
)

var discord *discordgo.Session

var guildID string
var donatorRole string
var verifiedRole string

// Discord's OAuth tokens are alphanumeric
var discordOAuthToken = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func init() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		fmt.Println("WARNING: No discord bot token, will not be able to grant donator role!")
		return
	}
	guildID = os.Getenv("DISCORD_GUILD_ID")
	donatorRole = os.Getenv("DISCORD_DONATOR_ROLE_ID")
	verifiedRole = os.Getenv("DISCORD_VERIFIED_ROLE_ID")
	if guildID == "" || donatorRole == "" || verifiedRole == "" {
		fmt.Println("WARNING: Discord info is bad")
		return
	}
	var err error
	discord, err = discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	user, err := discord.User("@me")
	if err != nil {
		panic(err)
	}

	myselfID := user.ID
	fmt.Println("I am", myselfID)
}

func GetUserId(accessToken string) (userId string, err error) {
	// Validate the token, prevent trying to auth with discord using some completely invalid token
	if !discordOAuthToken.MatchString(accessToken) {
		return "", echo.NewHTTPError(http.StatusBadRequest, "invalid access_token "+accessToken)
	}

	// Create a discord session using the provided token. Does not verify the token is valid in any way.
	// Using discordgo here is massively overkill, but who cares
	// This won't use websockets unless we call session.Open(), so there's no need to call Close() either.
	session, err := discordgo.New("Bearer " + accessToken)
	if err != nil {
		return "", echo.NewHTTPError(http.StatusInternalServerError, "error setting up discord session").SetInternal(err)
	}

	// Get the user's identity
	discordUser, err := session.User("@me")
	if err != nil {
		var restErr *discordgo.RESTError
		if errors.As(err, &restErr) {
			return "", echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf(`error authenticating with discord "%s"`, restErr.Message.Message))
		}
		return "", echo.NewHTTPError(http.StatusInternalServerError, "error authenticating with discord").SetInternal(err)
	}
	if discordUser.ID == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "no discord user found")
	}

	return discordUser.ID, nil
}

func JoinOurServer(accessToken string, discordID string, donator bool) error {
	roles := []string{verifiedRole}
	if donator {
		roles = append(roles, donatorRole)
	}
	return discord.GuildMemberAdd(accessToken, guildID, discordID, "", roles, false, false)
}

func GiveDonator(discordID string) error {
	discordID = strings.TrimSpace(discordID)
	go discord.ChannelMessageSend("556229210491060226", "<@"+discordID+"> just donated and received Impact premium!")
	go discord.ChannelMessageSend("308653317834145802", "<@"+discordID+"> just donated and received Impact premium!")
	return discord.GuildMemberRoleAdd(guildID, discordID, donatorRole)
}

func GiveVerified(discordID string) error {
	return discord.GuildMemberRoleAdd(guildID, discordID, verifiedRole)
}

func CheckServerMembership(discordID string) bool {
	member, err := discord.GuildMember(guildID, discordID)
	return err == nil && member != nil
}
