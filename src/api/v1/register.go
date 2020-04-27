package v1

import (
	"database/sql"
	"github.com/ImpactDevelopment/ImpactServer/src/jwt"
	"github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/ImpactDevelopment/ImpactServer/src/minecraft"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/discord"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func checkToken(c echo.Context) error {
	var body struct {
		Token string `json:"token" form:"token" query:"token"`
	}

	err := c.Bind(&body)
	if err != nil {
		return err
	}
	if body.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token missing")
	}
	token, err := getToken(body.Token)
	if err != nil {
		return err
	}

	var createdAt int64
	var roles pq.StringArray
	err = database.DB.QueryRow(`
		SELECT
			created_at,
			STRING_TO_ARRAY(
				CONCAT_WS(',',
					CASE WHEN premium THEN 'premium' END,
					CASE WHEN pepsi THEN 'pepsi' END,
					CASE WHEN spawnmason THEN 'spawnmason' END,
					CASE WHEN staff THEN 'staff' END
				),
				','
			) AS roles
		FROM pending_donations
		WHERE token = $1 AND NOT used`, token).Scan(&createdAt, &roles)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid token").SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error()).SetInternal(err)
	}

	return c.JSON(http.StatusOK, struct {
		CreatedAt string   `json:"created_at"`
		Roles     []string `json:"roles"`
	}{
		CreatedAt: time.Unix(createdAt, 0).UTC().Format(time.RFC3339),
		Roles:     roles,
	})
}

func registerWithToken(c echo.Context) error {
	var body struct {
		Token        string `json:"token" form:"token" query:"token"`
		DiscordToken string `json:"discord" form:"discord" query:"discord"`
		Minecraft    string `json:"minecraft" form:"minecraft" query:"minecraft"`
		Email        string `json:"email" form:"email" query:"email"`
		Password     string `json:"password" form:"password" query:"password"`
	}

	authedUser := middleware.GetUser(c)
	err := c.Bind(&body)
	if err != nil {
		return err
	}
	// Allow creating account without discord or minecraft
	if (authedUser == nil && body.Token == "") || body.Email == "" || body.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty field(s)")
	}

	// Verify the registration token
	var (
		token      *uuid.UUID
		createdAt  int64
		amount     int64
		used       bool
		logID      sql.NullString
		premium    bool
		pepsi      bool
		spawnmason bool
		staff      bool
	)
	// token can be omitted if logged in
	if body.Token != "" {
		token, err = getToken(body.Token)
		if err != nil {
			return err
		}
		err = database.DB.QueryRow("SELECT created_at, amount, used, log_msg_id, premium, pepsi, spawnmason, staff FROM pending_donations WHERE token = $1", token).Scan(&createdAt, &amount, &used, &logID, &premium, &pepsi, &spawnmason, &staff)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
		}
		if used {
			return echo.NewHTTPError(http.StatusConflict, "token already used")
		}
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
	} else if authedUser != nil {
		discordID = authedUser.DiscordID
	}

	var minecraftProfile *minecraft.Profile
	if body.Minecraft != "" || authedUser != nil {
		var mc = body.Minecraft
		if mc == "" && authedUser != nil {
			mc = authedUser.MinecraftID.String()
		}

		minecraftProfile, err = minecraft.GetProfile(mc)
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
	if user, err := findAccountFromIDs(email, discordID, minecraftProfile); err == nil && user == nil {
		// no error, but user is nil, so create a new user
		err = tx.QueryRow("INSERT INTO users(legacy) VALUES (false) RETURNING user_id").Scan(&userID)
		if err != nil {
			log.Print(err.Error())
			return err
		}
	} else if err == nil && user != nil {
		if authedUser == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "email, minecraft or discord is in use by an existing user")
		}
		if authedUser.ID != user.ID {
			return echo.NewHTTPError(http.StatusConflict, "something is being used by a different user (email, minecraft or discord)")
		}
		// no error and found a user, so use their id
		userID = &user.ID
	} else {
		return err
	}

	// Grant roles based on token
	_, err = tx.Exec(`UPDATE users
							SET premium    = $2 OR premium,
							    pepsi      = $3 OR pepsi,
							    spawnmason = $4 OR spawnmason,
							    staff      = $5 OR staff
							WHERE user_id = $1`,
		userID, premium, pepsi, spawnmason, staff)
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
	if minecraftProfile != nil {
		_, err := tx.Exec(`UPDATE users SET mc_uuid=$2 WHERE user_id = $1`, userID, minecraftProfile.ID)
		if err != nil {
			log.Print(err.Error())
			return err
		}
	}

	// TODO should we just DELETE the token?
	if token != nil {
		_, err = tx.Exec("UPDATE pending_donations SET used = true, used_by = $1 WHERE token = $2", userID, token)
		if err != nil {
			log.Print(err.Error())
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Print(err.Error())
		return err
	}

	go func() {
		var isMember bool
		if discordID != "" {
			// TODO grant donator status based on token roles
			if isMember = discord.CheckServerMembership(discordID); isMember {
				err = discord.GiveDonator(discordID)
			} else {
				err = discord.JoinOurServer(body.DiscordToken, discordID, true)
			}
			if err != nil {
				discord.LogDonationEvent(logID.String, "Error adding donator to discord: "+err.Error(), discordID, minecraftProfile, amount)
				return
			}
		}

		var msg strings.Builder
		msg.WriteString("Someone just")
		if premium && logID.String != "" {
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
		if authedUser == nil {
			msg.WriteString(" registered an")
		} else {
			msg.WriteString(" upgraded their")
		}
		msg.WriteString(" Impact Account")
		_, _ = discord.LogDonationEvent(logID.String, msg.String(), discordID, minecraftProfile, amount)
	}()

	// Get the user so we can log them in
	user := database.LookupUserByID(*userID)
	if user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "successfully registered, but can't find user")
	}

	return c.String(http.StatusOK, jwt.CreateUserJWT(user))
}

func getToken(token string) (*uuid.UUID, error) {
	tokenID, err := uuid.Parse(strings.TrimSpace(token))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid token format").SetInternal(err)
	}
	return &tokenID, nil
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

func findAccountFromIDs(email string, discordID string, minecraft *minecraft.Profile) (*users.User, error) {
	var (
		emailUser     *users.User
		discordUser   *users.User
		minecraftUser *users.User
	)

	// Lookup the users TODO do these lookups concurrently?
	if minecraft != nil {
		minecraftUser = database.LookupUserByMinecraftID(minecraft.ID)
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
