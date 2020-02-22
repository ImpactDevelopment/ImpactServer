package v1

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/discord"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type registrationCheck struct {
	Token string `json:"token" form:"token" query:"token"`
}

type registration struct {
	Token        string `json:"token" form:"token" query:"token"`
	DiscordToken string `json:"discord" form:"discord" query:"discord"`
	Minecraft    string `json:"minecraft" form:"minecraft" query:"minecraft"`
	Email        string `json:"email" form:"email" query:"email"`
	Password     string `json:"password" form:"password" query:"password"`
}

// https://wiki.vg/Mojang_API#Username_-.3E_UUID_at_time
type uuidLookupResponse struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func checkToken(c echo.Context) error {
	body := &registrationCheck{}
	err := c.Bind(body)
	if err != nil {
		return err
	}
	if body.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token missing")
	}
	var createdAt int64
	err = database.DB.QueryRow("SELECT created_at FROM pending_donations WHERE token = $1 AND NOT used", body.Token).Scan(&createdAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}
	return c.String(200, "true")
}

func registerWithToken(c echo.Context) error {
	body := &registration{}
	err := c.Bind(body)
	if err != nil {
		return err
	}
	// TODO allow creating account without discord or minecraft
	if body.Token == "" || body.DiscordToken == "" || body.Minecraft == "" || body.Email == "" || body.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty field(s)")
	}

	// get discord user id
	body.Token = strings.TrimSpace(body.Token)
	discordID, err := discord.GetUserId(body.DiscordToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid discord token")
	}

	var createdAt int64
	err = database.DB.QueryRow("SELECT created_at FROM pending_donations WHERE token = $1 AND NOT used", body.Token).Scan(&createdAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Try parsing minecraft as a UUID, if that fails use it as a name to lookup the UUID
	minecraftID, err := uuid.Parse(strings.TrimSpace(body.Minecraft))
	if err == nil && minecraftID.String() != "" {
		// Verify provided minecraft id
		req, err := util.GetRequest("https://api.mojang.com/user/profiles/" + url.PathEscape(strings.Replace(minecraftID.String(), "-", "", -1)) + "/names")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "bad mc uuid")
		}
		resp, err := req.Do()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "bad mc uuid")
		}
		if !resp.Ok() {
			return echo.NewHTTPError(http.StatusBadRequest, "bad mc uuid")
		}
	} else {
		// minecraft isn't an ID, it must be a name
		req, err := util.GetRequest("https://api.mojang.com/users/profiles/minecraft/" + url.PathEscape(strings.TrimSpace(body.Minecraft)))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "bad mc name")
		}
		resp, err := req.Do()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "bad mc name")
		}
		if !resp.Ok() {
			return echo.NewHTTPError(http.StatusBadRequest, "bad mc name")
		}
		var respBody uuidLookupResponse
		err = resp.JSON(&respBody)
		if err != nil || respBody.Id.String() == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "bad mc name")
		}

		minecraftID = respBody.Id
	}

	// TODO check for existing rows first
	// TODO check which roles should be assigned based on the token
	var userID uuid.UUID
	err = database.DB.QueryRow(`INSERT INTO users(legacy, email, password_hash, mc_uuid, discord_id, roles) VALUES (false, $1, $2, $3, $4, '{"premium"}') RETURNING user_id`, body.Email, hashedPassword, minecraftID, discordID).Scan(&userID)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = database.DB.Exec(`UPDATE pending_donations SET used = true, used_by = $1 WHERE token = $2`, userID, body.Token)
	if err != nil {
		log.Println(err)
		return err
	}

	if discord.CheckServerMembership(discordID) {
		err = discord.GiveDonator(discordID)
	} else {
		err = discord.JoinOurServer(body.DiscordToken, discordID, true)
	}
	if err != nil {
		log.Printf("Error adding donator to discord: %s\n", err.Error())
		return err
	}

	// TODO redirect to dashboard
	const donatorInfo = "https://discordapp.com/channels/208753003996512258/613478149669388298"
	return c.Redirect(http.StatusFound, donatorInfo)
}
