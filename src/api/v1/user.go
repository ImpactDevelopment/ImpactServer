package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func getUser(c echo.Context) error {
	if user := middleware.GetUser(c); user != nil {
		return c.JSON(http.StatusOK, user)
	} else {
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to cast user")
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
			discordID, err := getDiscordID(*body.DiscordToken)
			if err != nil {
				return err
			}
			if discordID != user.DiscordID {
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

		return c.JSONBlob(http.StatusOK, []byte(`{"message":"success"}`))
	} else {
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to cast user")
	}
}
