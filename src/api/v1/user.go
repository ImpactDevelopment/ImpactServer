package v1

import (
	"database/sql"
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/discord"
	"github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type mcuser struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type resultDiscord struct {
	Discord *discord.User
	Error   error
}

type resultMC struct {
	User  *mcuser
	Error error
}

func getUser(c echo.Context) error {
	if user := middleware.GetUser(c); user != nil {
		type resp struct {
			Email         string          `json:"email"`
			Minecraft     *mcuser         `json:"minecraft,omitempty"`
			Discord       *discord.User   `json:"discord,omitempty"`
			LegacyEnabled bool            `json:"legacy_enabled"`
			Incognito     bool            `json:"incognito"`
			Roles         []users.Role    `json:"roles,omitempty"`
			Info          *users.UserInfo `json:"info,omitempty"`
		}

		// Lookup minecraft and discord in parallel
		minecraftCh := make(chan resultMC)
		discordCh := make(chan resultDiscord)
		go func() { minecraftCh <- lookupMinecraftInfo(user.MinecraftID) }()
		go func() { discordCh <- lookupDiscordInfo(user.DiscordID) }()
		var (
			minecraftResult = <-minecraftCh
			discordResult   = <-discordCh
		)
		if minecraftResult.Error != nil {
			return minecraftResult.Error
		}
		if discordResult.Error != nil {
			return discordResult.Error
		}

		return c.JSON(http.StatusOK, resp{
			Email:         user.Email,
			Minecraft:     minecraftResult.User,
			Discord:       discordResult.Discord,
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
				_, err = tx.Exec(`UPDATE users SET discord_id = $2 WHERE user_id = $1`, user.ID, discordID)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
				}
			}
		}

		if body.Minecraft != nil {
			minecraftID, err := getMinecraftID(*body.Minecraft)
			if err != nil {
				return err
			}
			if minecraftID != user.MinecraftID {
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
	} else {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
}

func lookupDiscordInfo(id string) resultDiscord {
	if id == "" {
		return resultDiscord{}
	}
	dcUser, err := discord.GetUser(id)
	return resultDiscord{
		Discord: dcUser,
		Error:   err,
	}
}

func lookupMinecraftInfo(id *uuid.UUID) resultMC {
	if id == nil {
		return resultMC{}
	}

	// Construct the error in advance lol
	e := echo.NewHTTPError(http.StatusInternalServerError, "Failed to lookup minecraft info for id "+id.String())

	// Lookup minecraft name
	req, err := util.GetRequest("https://api.mojang.com/user/profiles/" + url.PathEscape(strings.Replace(id.String(), "-", "", -1)) + "/names")
	if err != nil {
		return resultMC{
			Error: e.SetInternal(err),
		}
	}
	response, err := req.Do()
	if err != nil {
		return resultMC{
			Error: e.SetInternal(err),
		}
	}
	if !response.Ok() {
		return resultMC{
			Error: e,
		}
	}

	// Parse response
	type name struct {
		Name string `json:"name"`
		At   int64  `json:"changedToAt"`
	}
	var body = make([]name, 5)
	err = response.JSON(&body)
	if err != nil {
		return resultMC{
			Error: e.SetInternal(err),
		}
	}
	if len(body) < 1 {
		return resultMC{
			Error: e,
		}
	}

	// Find the most recent name, this is probably body[len(body)-1] but this is safer
	var newest name
	for _, it := range body {
		if it.At > newest.At {
			newest = it
		}
	}

	return resultMC{
		User: &mcuser{
			ID:   *id,
			Name: newest.Name,
		},
	}
}
