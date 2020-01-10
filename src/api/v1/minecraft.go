package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/cloudflare"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
)

var userData map[string]users.UserInfo

var legacyRoles map[string]string

// API Handler
func getUserInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, userData)
}

// /minecraft/user/:role/list
func getRoleMembers(c echo.Context) error {
	ret := legacyRoles[c.Param("role")]
	if ret == "" {
		return c.NoContent(http.StatusNotFound)
	}
	return c.String(http.StatusOK, ret)
}

func userHasRole(user users.User, roleName string) bool {
	for _, role := range user.Roles() {
		if role.ID == roleName {
			return true
		}
	}
	return false
}

func init() {
	usersList := users.GetAllUsers()
	updatedData(usersList)
	updatedLegacyRoles(usersList)
	util.DoRepeatedly(5*time.Minute, func() {
		usersList := users.GetAllUsers()
		if updatedData(usersList) {
			log.Println("MC UPDATE: Updated user info")
			cloudflare.PurgeURLs([]string{
				"https://api.impactclient.net/v1/minecraft/user/info",
			})
		}
		if updatedLegacyRoles(usersList) {
			log.Println("MC UPDATE: Updated user legacy data")
			cloudflare.PurgeURLs([]string{
				"https://api.impactclient.net/v1/minecraft/user/staff/list",
				"https://api.impactclient.net/v1/minecraft/user/developer/list",
				"https://api.impactclient.net/v1/minecraft/user/pepsi/list",
				"https://api.impactclient.net/v1/minecraft/user/premium/list",
			})
		}
	})
}

func updatedData(usersList []users.User) bool {
	newUserData := generateMap(usersList)
	// reflect.DeepEqual is slow, especially since this map is big
	if userData == nil || !reflect.DeepEqual(newUserData, userData) {
		userData = newUserData
		return true
	}
	return false
}

func updatedLegacyRoles(usersList []users.User) bool {
	newLegacyRoles := generateLegacy(usersList)
	// reflect.DeepEqual is slow, especially since this map is big
	if legacyRoles == nil || !reflect.DeepEqual(newLegacyRoles, legacyRoles) {
		legacyRoles = newLegacyRoles
		return true
	}
	return false
}

func generateLegacy(usersList []users.User) map[string]string {
	m := make(map[string]string)
	for role, _ := range users.RolesData {
		ret := ""
		for _, user := range usersList {
			if !userHasRole(user, role) {
				continue
			}
			if !user.IsLegacy() {
				continue
			}
			if uuid := user.MinecraftID(); uuid != nil {
				ret += uuid.String() + "\n"
			}
		}
		m[role] = ret
	}
	return m
}

func generateMap(usersList []users.User) map[string]users.UserInfo {
	data := make(map[string]users.UserInfo)
	for _, user := range usersList {
		if uuid := user.MinecraftID(); uuid != nil {
			userInfo := user.UserInfo()
			var empty users.UserInfo
			if userInfo == empty {
				// if a user has cape disabled, they are trying to be incognito. we should send no entry at all. not good enough to send "HASH123":{}.
				continue
			}
			data[hashUUID(*uuid)] = userInfo
		}
	}
	return data
}

func hashUUID(uuid uuid.UUID) string {
	hash := sha256.Sum256([]byte(uuid.String()))
	return hex.EncodeToString(hash[:])
}
