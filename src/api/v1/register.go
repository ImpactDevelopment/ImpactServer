package v1

import (
	"log"
	"net/http"
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
	Token    string `json:"token" form:"token" query:"token"`
	Discord  string `json:"discord" form:"discord" query:"discord"`
	Mcuuid   string `json:"mcuuid" form:"mcuuid" query:"mcuuid"`
	Email    string `json:"email" form:"email" query:"email"`
	Password string `json:"password" form:"password" query:"password"`
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
	if body.Token == "" || body.Discord == "" || body.Mcuuid == "" || body.Email == "" || body.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty field(s)")
	}
	var createdAt int64
	err = database.DB.QueryRow("SELECT created_at FROM pending_donations WHERE token = $1 AND NOT used", body.Token).Scan(&createdAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}
	log.Println(body)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if !discord.CheckServerMembership(body.Discord) {
		return echo.NewHTTPError(http.StatusBadRequest, "join our discord first lol")
	}

	req, err := util.GetRequest("https://api.mojang.com/user/profiles/" + strings.Replace(body.Mcuuid, "-", "", -1) + "/names")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad mc uuid")
	}
	resp, err := req.Do()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad mc uuid")
	}
	if resp.Code() != 200 {
		return echo.NewHTTPError(http.StatusBadRequest, "bad mc uuid")
	}
	var userID uuid.UUID
	err = database.DB.QueryRow("INSERT INTO users(email, password_hash, mc_uuid, discord_id) VALUES ($1, $2, $3, $4) RETURNING user_id", body.Email, hashedPassword, body.Mcuuid, body.Discord).Scan(&userID)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = database.DB.Exec("UPDATE pending_donations SET used = true, used_by = $1 WHERE token = $2", userID, body.Token)
	if err != nil {
		log.Println(err)
		return err
	}
	err = discord.GiveDonator(body.Discord)
	if err != nil {
		log.Println(err)
		return err
	}
	return c.String(200, "SUCCESS")
}
