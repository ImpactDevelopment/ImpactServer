package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func getUser(c echo.Context) error {
	if user, ok := c.Get("user").(*users.User); ok {
		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "no user found")
		}
		return c.JSON(http.StatusOK, user)
	} else {
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to cast user")
	}
}

func patchUser(c echo.Context) error {
	if user, ok := c.Get("user").(*users.User); ok {
		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "no user found")
		}

		// Bind the request body to a User struct
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

		// TODO use a transaction or compile a query or something
		if body.Email != nil {
			email, err := verifyEmail(*body.Email)
			if err != nil {
				return err
			}

			// Update the DB
			_, err = database.DB.Exec("UPDATE users SET email = $2 WHERE user_id = $1", user.ID, email)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		if body.Password != nil {
			hashedPassword, err := hashPassword(*body.Password)
			if err != nil {
				return err
			}

			// Update the DB
			_, err = database.DB.Exec("UPDATE users SET password_hash = $2 WHERE user_id = $1", user.ID, hashedPassword)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		if body.DiscordToken != nil {
			discordID, err := getDiscordID(*body.DiscordToken)
			if err != nil {
				return err
			}

			// Update the DB
			_, err = database.DB.Exec("UPDATE users SET discord_id = $2 WHERE user_id = $1", user.ID, discordID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		if body.Minecraft != nil {
			minecraftID, err := getMinecraftID(*body.Minecraft)
			if err != nil {
				return err
			}

			// Update the DB
			_, err = database.DB.Exec("UPDATE users SET mc_uuid = $2 WHERE user_id = $1", user.ID, minecraftID)
			if err != nil {
				log.Println(err)
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		if body.Incognito != nil {
			// Update the DB
			// TODO convert capes_enabled to incognito on the db side too
			_, err = database.DB.Exec("UPDATE users SET capes_enabled = $2 WHERE user_id = $1", user.ID, !*body.Incognito)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		if body.LegacyEnabled != nil {
			// Update the DB
			_, err = database.DB.Exec("UPDATE users SET legacy_enabled = $2 WHERE user_id = $1", user.ID, *body.LegacyEnabled)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}

		return c.JSONBlob(http.StatusOK, []byte(`{"message":"success"}`))
	} else {
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to cast user")
	}
}
