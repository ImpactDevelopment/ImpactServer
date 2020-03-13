package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/jwt"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
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
	// Allow creating account without discord or minecraft
	if body.Token == "" || body.Email == "" || body.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty field(s)")
	}

	// Verify the registration token
	// TODO get roles from token
	body.Token = strings.TrimSpace(body.Token)
	var (
		createdAt int64
		amount    int64
		used      bool
		logID     string
	)
	err = database.DB.QueryRow("SELECT created_at, amount, used, log_msg_id FROM pending_donations WHERE token = $1", body.Token).Scan(&createdAt, &amount, &used, &logID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}
	if used {
		return echo.NewHTTPError(http.StatusConflict, "token already used")
	}

	email, err := verifyEmail(body.Email)
	if err != nil {
		return err
	}

	hashedPassword, err := hashPassword(body.Password)
	if err != nil {
		return err
	}

	var discordID string
	if body.DiscordToken != "" {
		discordID, err = getDiscordID(body.DiscordToken)
		if err != nil {
			return err
		}
	}

	var minecraftID *uuid.UUID
	if body.Minecraft != "" {
		minecraftID, err = getMinecraftID(body.Minecraft)
		if err != nil {
			return err
		}
	}

	// Make DB changes in a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Find or create the user
	var userID *uuid.UUID
	if user, err := findAccountFromIDs(email, discordID, minecraftID); err == nil && user == nil {
		// no error, but user is nil, so create a new user
		err = tx.QueryRow("INSERT INTO users(legacy) VALUES (false) RETURNING user_id").Scan(&userID)
		if err != nil {
			log.Print(err.Error())
			return err
		}
	} else if err == nil && user != nil {
		// no error and found a user, so use their id
		userID = &user.ID
	} else {
		return err
	}

	// TODO set roles based on token roles array
	premium := true
	_, err = tx.Exec(`UPDATE users SET premium=$2 WHERE user_id = $1`, userID, premium)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	_, err = tx.Exec(`UPDATE users SET email=$2, password_hash=$3 WHERE user_id = $1`, userID, email, hashedPassword)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	if discordID != "" {
		_, err := tx.Exec(`UPDATE users SET discord_id=$2 WHERE user_id = $1`, userID, discordID)
		if err != nil {
			log.Print(err.Error())
			return err
		}
	}
	if minecraftID != nil {
		_, err := tx.Exec(`UPDATE users SET mc_uuid=$2 WHERE user_id = $1`, userID, minecraftID)
		if err != nil {
			log.Print(err.Error())
			return err
		}
	}

	// TODO should we just DELETE the token?
	_, err = tx.Exec("UPDATE pending_donations SET used = true, used_by = $1 WHERE token = $2", userID, body.Token)
	if err != nil {
		log.Print(err.Error())
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Print(err.Error())
		return err
	}

	go func() {
		var isMember bool
		var mcID string
		if minecraftID != nil {
			mcID = minecraftID.String()
		}

		if discordID != "" {
			// TODO grant donator status based on token roles
			if isMember = discord.CheckServerMembership(discordID); isMember {
				err = discord.GiveDonator(discordID)
			} else {
				err = discord.JoinOurServer(body.DiscordToken, discordID, true)
			}
			if err != nil {
				discord.LogDonationEvent(logID, "Error adding donator to discord: "+err.Error(), discordID, mcID, amount)
				return
			}
		}

		var msg strings.Builder
		msg.WriteString("Someone just")
		if logID != "" {
			// TODO get this bit _from_ the previous log msg?
			msg.WriteString(" donated")
		}
		if discordID != "" && !isMember {
			if msg.String() != "Someone just" {
				msg.WriteString(",")
			}
			msg.WriteString(" joined the server")
		}
		if msg.String() != "Someone just" {
			msg.WriteString(" and")
		}
		msg.WriteString(" registered an Impact Account")
		_, _ = discord.LogDonationEvent(logID, msg.String(), discordID, mcID, amount)
	}()

	// Get the user so we can log them in
	user := database.LookupUserByID(*userID)
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

func findAccountFromIDs(email string, discordID string, minecraftID *uuid.UUID) (*users.User, error) {
	var (
		emailUser     *users.User
		discordUser   *users.User
		minecraftUser *users.User
	)

	// Lookup the users TODO do these lookups concurrently?
	if minecraftID != nil {
		minecraftUser = database.LookupUserByMinecraftID(*minecraftID)
	}
	if discordID != "" {
		discordUser = database.LookupUserByDiscordID(discordID)
	}
	if email != "" {
		emailUser = database.LookupUserByEmail(email)
	}

	// Validation; ensure all matched users are the same as each other
	// yes, this is horrible
	if discordUser != nil {
		if emailUser != nil && discordUser.ID != emailUser.ID {
			return nil, echo.NewHTTPError(http.StatusConflict, "discord belongs to a different account to email")
		}
		if minecraftUser != nil && discordUser.ID != minecraftUser.ID {
			return nil, echo.NewHTTPError(http.StatusConflict, "discord belongs to a different account to minecraft")
		}
		if discordUser.Email != "" && email != discordUser.Email {
			return nil, echo.NewHTTPError(http.StatusForbidden, "cannot modify account email")
		}
	}

	if minecraftUser != nil {
		if emailUser != nil && minecraftUser.ID != emailUser.ID {
			return nil, echo.NewHTTPError(http.StatusConflict, "minecraft belongs to a different account to email")
		}
		if discordUser != nil && minecraftUser.ID != discordUser.ID {
			return nil, echo.NewHTTPError(http.StatusConflict, "minecraft belongs to a different account to discord")
		}
		if minecraftUser.Email != "" && email != minecraftUser.Email {
			return nil, echo.NewHTTPError(http.StatusForbidden, "cannot modify account email")
		}
	}

	if emailUser != nil {
		if discordUser != nil && emailUser.ID != discordUser.ID {
			return nil, echo.NewHTTPError(http.StatusConflict, "email belongs to a different account to discord")
		}
		if minecraftUser != nil && emailUser.ID != minecraftUser.ID {
			return nil, echo.NewHTTPError(http.StatusConflict, "email belongs to a different account to minecraft")
		}
		// If the user has a password (and they didn't also auth with the matching discord account) we should treat this as an attempt to hijack their account
		// TODO should we also compare the provided password with the password hash?
		// TODO should we provide an alternative method for existing (non-donator) users to donate?
		//      e.g. if logged in, just associate donations with the logged in user?
		//      or allow users to enter a token in their account dashboard?
		if emailUser.PasswordHash != "" && discordUser == nil {
			return nil, echo.NewHTTPError(http.StatusConflict, "email belongs to a user with a password set")
		}
	}

	// Validation done, return whichever user isn't nil
	if emailUser != nil {
		return emailUser, nil
	}
	if discordUser != nil {
		return discordUser, nil
	}
	if minecraftUser != nil {
		return minecraftUser, nil
	}

	// No user found, but also no error
	return nil, nil
}
