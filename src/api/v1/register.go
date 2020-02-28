package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/jwt"
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

	// Verify the registration token
	// TODO get roles from token
	body.Token = strings.TrimSpace(body.Token)
	var createdAt int64
	err = database.DB.QueryRow("SELECT created_at FROM pending_donations WHERE token = $1 AND NOT used", body.Token).Scan(&createdAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}

	email, err := verifyEmail(body.Email)
	if err != nil {
		return err
	}

	hashedPassword, err := hashPassword(body.Password)
	if err != nil {
		return err
	}

	// get discord user id
	discordID, err := getDiscordID(body.DiscordToken)
	if err != nil {
		return err
	}

	minecraftID, err := getMinecraftID(body.Minecraft)
	if err != nil {
		return err
	}

	// Insert the user, TODO check if they exist first
	var userID uuid.UUID
	err = database.DB.QueryRow("INSERT INTO users(legacy, email, password_hash, mc_uuid, discord_id) VALUES (false, $1, $2, $3, $4) RETURNING user_id", email, hashedPassword, minecraftID, discordID).Scan(&userID)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = database.DB.Exec("UPDATE pending_donations SET used = true, used_by = $1 WHERE token = $2", userID, body.Token)
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

	// Get the user so we can log them in
	user := database.LookupUserByID(userID)
	if user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "successfully registered, but can't find user")
	}
	token := jwt.CreateUserJWT(user)

	return c.String(http.StatusOK, token)
}

func verifyEmail(email string) (string, error) {
	return email, nil // TODO
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		err = echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password").SetInternal(err)
	}
	return string(hashedPassword), err

}

func getDiscordID(token string) (string, error) {
	discordID, err := discord.GetUserId(strings.TrimSpace(token))
	if err != nil {
		err = echo.NewHTTPError(http.StatusBadRequest, "invalid discord token").SetInternal(err)
	}
	return discordID, err
}

func getMinecraftID(minecraft string) (*uuid.UUID, error) {
	// Try parsing minecraft as a UUID, if that fails use it as a name to lookup the UUID
	minecraftID, err := uuid.Parse(strings.TrimSpace(minecraft))
	if err == nil && minecraftID.String() != "" {
		// minecraft is an id, verify it
		var bad = echo.NewHTTPError(http.StatusBadRequest, "bad minecraft uuid")

		req, err := util.GetRequest("https://api.mojang.com/user/profiles/" + url.PathEscape(strings.Replace(minecraftID.String(), "-", "", -1)) + "/names")
		if err != nil {
			return nil, bad
		}
		resp, err := req.Do()
		if err != nil {
			return nil, bad
		}
		if !resp.Ok() {
			return nil, bad
		}
	} else {
		// minecraft must be a name, look up the id
		var bad = echo.NewHTTPError(http.StatusBadRequest, "bad minecraft username")

		req, err := util.GetRequest("https://api.mojang.com/users/profiles/minecraft/" + url.PathEscape(strings.TrimSpace(minecraft)))
		if err != nil {
			return nil, bad
		}
		resp, err := req.Do()
		if err != nil {
			return nil, bad
		}
		if !resp.Ok() {
			return nil, bad
		}

		// Parse the response
		// https://wiki.vg/Mojang_API#Username_-.3E_UUID_at_time
		var respBody struct {
			Id   uuid.UUID `json:"id"`
			Name string    `json:"name"`
		}
		err = resp.JSON(&respBody)
		if err != nil || respBody.Id.String() == "" {
			return nil, bad
		}
		minecraftID = respBody.Id
	}

	return &minecraftID, nil
}
