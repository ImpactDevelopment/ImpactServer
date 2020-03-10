package discord

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

var discord *discordgo.Session

var guildID string
var donatorRole string
var verifiedRole string

// Discord's OAuth tokens are alphanumeric
var discordOAuthToken = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// Where to notify donations
const donationMsgChannel = "678230156091064330"

// User is a wrapper around discordgo.User so that we can feel ok about exporting it,
// we also set Avatar to a full url instead of just the id for json reasons
type User struct {
	discordgo.User
	Avatar string `json:"avatar"`
}

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

// GetUserId returns the discord user id using the user's access token
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

// GetUser returns the user object matching the given user id
func GetUser(id string) (user *User, err error) {
	discordUser, err := discord.User(id)
	if err != nil {
		var restErr *discordgo.RESTError
		if errors.As(err, &restErr) {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf(`error authenticating with discord "%s"`, restErr.Message.Message))
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "error authenticating with discord").SetInternal(err)
	}

	return &User{*discordUser, discordUser.AvatarURL("")}, nil

}

// JoinOurServer adds the user matching discordId to our discord server. The user's access token must be provided and it
// must have the guilds.join scope
func JoinOurServer(accessToken string, discordID string, donator bool) error {
	roles := []string{verifiedRole}
	if donator {
		roles = append(roles, donatorRole)
	}
	return discord.GuildMemberAdd(accessToken, guildID, discordID, "", roles, false, false)
}

// SetDonator updates the roles for the given discord user without treating it like a new donation
func SetDonator(discordID string, donator bool) error {
	if donator {
		return discord.GuildMemberRoleAdd(guildID, discordID, donatorRole)
	} else {
		return discord.GuildMemberRoleRemove(guildID, discordID, donatorRole)
	}
}

// GiveDonator grants the donator role to the user and verifies them
func GiveDonator(discordID string) error {
	GiveVerified(discordID)
	// dont return early & fail to give donator role if we cant give verified
	return discord.GuildMemberRoleAdd(guildID, discordID, donatorRole)
}

// GiveVerified grants the user the verified role which allows them to see channels and talk in them.
func GiveVerified(discordID string) error {
	return discord.GuildMemberRoleAdd(guildID, discordID, verifiedRole)
}

// CheckServerMembership is true if the user is a member of our guild
func CheckServerMembership(discordID string) bool {
	member, err := discord.GuildMember(guildID, discordID)
	return err == nil && member != nil
}

func LogDonationEvent(msg string, discordID string, minecraftID string, amount int64) error {
	m := discordgo.MessageSend{Content: msg}
	if discordID != "" || minecraftID != "" || amount > 0 {
		m.Embed = &discordgo.MessageEmbed{}
	}

	if amount > 0 {
		m.Embed.Fields = append(m.Embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Donation",
			Value:  fmt.Sprintf("$%01d.%02d", amount/100, amount%100),
			Inline: false,
		})
	}

	if discordID != "" {
		m.Embed.Fields = append(m.Embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Discord",
			Value:  discordgo.User{ID: discordID}.Mention(),
			Inline: true,
		})
	}

	if minecraftID != "" {
		m.Embed.Fields = append(m.Embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Minecraft",
			Value:  fmt.Sprintf(`[%s](https://namemc.com/profile/%s)`, minecraftID, minecraftID),
			Inline: true,
		})
		m.Embed.Image = &discordgo.MessageEmbedImage{
			URL: "https://crafatar.com/avatars/" + minecraftID,
		}
	}

	_, err := discord.ChannelMessageSendComplex(donationMsgChannel, &m)
	return err
}
