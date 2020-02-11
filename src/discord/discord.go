package discord

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

var discord *discordgo.Session

var guildID string
var donatorRole string

func init() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		fmt.Println("WARNING: No discord bot token, will not be able to grant donator role!")
		return
	}
	guildID = os.Getenv("DISCORD_GUILD_ID")
	donatorRole = os.Getenv("DISCORD_DONATOR_ROLE_ID")
	if guildID == "" || donatorRole == "" {
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

func GiveDonator(discordID string) error {
	return discord.GuildMemberRoleAdd(guildID, discordID, donatorRole)
}

func CheckServerMembership(discordID string) bool {
	member, err := discord.GuildMember(guildID, discordID)
	return err == nil && member != nil
}
