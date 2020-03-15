package v1

import (
	"database/sql"
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/discord"
	"github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/ImpactDevelopment/ImpactServer/src/minecraft"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"strings"
)

type (
	resultDiscord struct {
		Discord *discord.User
		Error   error
	}
	resultMC struct {
		Profile *minecraft.Profile
		Error   error
	}
	resultEdition struct {
		Edition *users.Edition
	}
)

func getUser(c echo.Context) error {
	if user := middleware.GetUser(c); user != nil {
		type resp struct {
			Email         string             `json:"email"`
			Minecraft     *minecraft.Profile `json:"minecraft,omitempty"`
			Discord       *discord.User      `json:"discord,omitempty"`
			Edition       *users.Edition     `json:"edition,omitempty"`
			LegacyEnabled bool               `json:"legacy_enabled"`
			Incognito     bool               `json:"incognito"`
			Roles         []users.Role       `json:"roles,omitempty"`
			Info          *users.UserInfo    `json:"info,omitempty"`
		}

		// Lookup minecraft and discord in parallel
		minecraftCh := make(chan resultMC)
		discordCh := make(chan resultDiscord)
		editionCh := make(chan resultEdition)
		go func() {
			if user.MinecraftID == nil {
				minecraftCh <- resultMC{}
				return
			}
			profile, err := minecraft.GetProfile(user.MinecraftID.String())
			minecraftCh <- resultMC{
				Profile: profile,
				Error:   err,
			}
		}()
		go func() {
			if user.DiscordID == "" {
				discordCh <- resultDiscord{}
				return
			}
			discordUser, err := discord.GetUser(user.DiscordID)
			discordCh <- resultDiscord{
				Discord: discordUser,
				Error:   err,
			}
		}()
		go func() {
			editionCh <- resultEdition{
				Edition: user.Edition(),
			}
		}()
		var (
			minecraftResult = <-minecraftCh
			discordResult   = <-discordCh
			editionResult   = <-editionCh
		)
		if minecraftResult.Error != nil {
			return minecraftResult.Error
		}
		if discordResult.Error != nil {
			return discordResult.Error
		}

		return c.JSON(http.StatusOK, resp{
			Email:         user.Email,
			Minecraft:     minecraftResult.Profile,
			Discord:       discordResult.Discord,
			Edition:       editionResult.Edition,
			LegacyEnabled: user.LegacyEnabled,
			Incognito:     user.Incognito,
			Roles:         user.Roles,
			Info:          user.UserInfo,
		})
	} else {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
}

func patchUser(c echo.Context) error {
	if user := middleware.GetUser(c); user != nil {
		// Everything is a pointer so we can check what was present in the request
		// e.g. an unset field defaulting to false might be bad
		var body struct {
			Email         *string `json:"email"`
			Minecraft     *string `json:"minecraft"`
			DiscordToken  *string `json:"discord"`
			Password      *string `json:"password"`
			LegacyEnabled *bool   `json:"legacy_enabled"`
			Incognito     *bool   `json:"incognito"`
		}
		err := c.Bind(&body)
		if err != nil {
			return err
		}

		// Use a transaction so that the DB can maybe optimise or at least rollback if any one step fails
		tx, err := database.DB.Begin()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "error starting database transaction").SetInternal(err)
		}
		defer tx.Rollback()

		if body.Email != nil && *body.Email != user.Email {
			email, err := verifyEmail(*body.Email)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`UPDATE users SET email = $2 WHERE user_id = $1`, user.ID, email)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		if body.Password != nil {
			hashedPassword, err := hashPassword(*body.Password)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`UPDATE users SET password_hash = $2 WHERE user_id = $1`, user.ID, hashedPassword)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		if body.DiscordToken != nil {
			// Don't lookup id if the token is empty or falsy; instead set discord_id to NullString's default: NULL
			var discordID sql.NullString
			if token := strings.TrimSpace(strings.ToLower(*body.DiscordToken)); token != "" && token != "false" && token != "null" && token != "0" {
				id, err := getDiscordID(strings.TrimSpace(*body.DiscordToken))
				if err != nil {
					return err
				}
				discordID = sql.NullString{
					String: id,
					Valid:  true,
				}
			}
			if discordID.String != user.DiscordID {
				// Revoke roles from current linked discord
				if user.DiscordID != "" && discord.CheckServerMembership(user.DiscordID) {
					err := discord.SetDonator(user.DiscordID, false)
					if err != nil {
						return echo.NewHTTPError(http.StatusInternalServerError, "Unable to remove roles from current Discord account").SetInternal(err)
					}
				}
				// Update the database
				_, err = tx.Exec(`UPDATE users SET discord_id = $2 WHERE user_id = $1`, user.ID, discordID)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
				}
				// Grant roles to the new discord user
				if discordID.Valid && user.HasRoleWithID("premium") {
					if discord.CheckServerMembership(discordID.String) {
						err := discord.SetDonator(discordID.String, true)
						if err != nil {
							return echo.NewHTTPError(http.StatusInternalServerError, "Unable to grant roles to new Discord account").SetInternal(err)
						}
					} else {
						// TODO join guild? Or maybe include "not joined" in response so the client can know to show a "join" button?
					}
				}
			}
		}

		if body.Minecraft != nil {
			var minecraftID *uuid.UUID
			if *body.Minecraft != "" {
				profile, err := minecraft.GetProfile(*body.Minecraft)
				if err != nil {
					return err
				}
				minecraftID = &profile.ID
			}

			var changed bool
			if minecraftID == nil {
				changed = user.MinecraftID != nil
			} else if user.MinecraftID == nil {
				changed = true
			} else {
				changed = minecraftID.String() != user.MinecraftID.String()
			}

			if changed {
				_, err = tx.Exec(`UPDATE users SET mc_uuid = $2 WHERE user_id = $1`, user.ID, minecraftID)
				if err != nil {
					log.Println(err)
					return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
				}
			}
		}

		if body.Incognito != nil && *body.Incognito != user.Incognito {
			var capeEnabled = !*body.Incognito // we store this inverted lol
			_, err = tx.Exec(`UPDATE users SET cape_enabled = $2 WHERE user_id = $1`, user.ID, capeEnabled)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		if body.LegacyEnabled != nil && *body.LegacyEnabled != user.LegacyEnabled {
			_, err = tx.Exec(`UPDATE users SET legacy_enabled = $2 WHERE user_id = $1`, user.ID, *body.LegacyEnabled)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		// Update the DB
		err = tx.Commit()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "error committing changes to the database").SetInternal(err)
		}

		// update context and then defer to getUser
		c.Set("user", database.LookupUserByID(user.ID))
		return getUser(c)
	}
	return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
}
