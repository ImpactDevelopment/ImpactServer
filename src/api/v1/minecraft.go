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

// API Handler
func getUserInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, userData)
}

func init() {
	_, err := updateData()
	if err != nil {
		panic(err)
	}
	util.DoRepeatedly(5*time.Minute, func() {
		updated, err := updateData()
		if err != nil {
			log.Println("MC ERROR", err)
			return
		}
		if updated {
			log.Println("MC UPDATE: Updated user info")
			cloudflare.PurgeURLs([]string{"https://api.impactclient.net/v1/minecraft/user/info"})
		}
	})
}

func updateData() (bool, error) {
	err := users.UpdateLegacyData()
	if err != nil {
		return false, err
	}
	return updatedData(), nil
}

func updatedData() bool {
	newUserData := generateMap()
	// reflect.DeepEqual is slow, especially since this map is big
	if userData == nil || !reflect.DeepEqual(newUserData, userData) {
		userData = newUserData
		return true
	}
	return false
}

func generateMap() map[string]users.UserInfo {
	data := make(map[string]users.UserInfo)
	for _, user := range users.GetAllUsers() {
		for _, uuid := range user.MinecraftIDs() {
			data[hashUUID(uuid)] = user.UserInfo()
		}
	}
	return data
}

func hashUUID(uuid uuid.UUID) string {
	hash := sha256.Sum256([]byte(uuid.String()))
	return hex.EncodeToString(hash[:])
}
