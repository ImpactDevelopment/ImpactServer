package v1

import (
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/jwt"

	"github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/labstack/echo/v4"
)

// API configures the Group to implement v1 of the API
func API(api *echo.Group) {
	// TODO API Doc

	api.GET("/thealtening/info", getTheAlteningInfo, middleware.CacheUntilPurge())
	api.GET("/motd", getMotd, middleware.CacheUntilPurge())
	api.GET("/themes", getThemes, middleware.CacheUntilPurge())
	api.GET("/minecraft/user/info", getUserInfo, middleware.CacheUntilPurge())
	api.GET("/minecraft/user/:role/list", getRoleMembers, middleware.CacheUntilPurge())
	api.GET("/dbtest", dbTest, middleware.NoCache())
	api.Match([]string{http.MethodGet, http.MethodPost}, "/login/password", jwt.PasswordLoginHandler, middleware.NoCache())
	api.Match([]string{http.MethodGet, http.MethodPost}, "/login/minecraft", jwt.MinecraftLoginHandler, middleware.NoCache())
	api.Match([]string{http.MethodGet, http.MethodPost}, "/login/discord", jwt.DiscordLoginHandler, middleware.NoCache())
	api.Match([]string{http.MethodGet, http.MethodPost}, "/paypal/afterpayment", afterDonation, middleware.NoCache())
	api.Match([]string{http.MethodGet, http.MethodPost}, "/checktoken", checkToken, middleware.NoCache())
	api.Match([]string{http.MethodGet, http.MethodPost}, "/register/token", registerWithToken, middleware.NoCache())
	api.GET("/emailtest", emailTest, middleware.NoCache())
	api.GET("/premiumcheck", premiumCheck, middleware.NoCache())
	api.GET("/integration/futureclient/masonlist", futureIntegration, middleware.NoCache())
}
