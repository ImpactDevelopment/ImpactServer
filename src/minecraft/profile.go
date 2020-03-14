package minecraft

import (
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"strings"
)

// Profile includes both the minecraft user's ID and their name
type Profile struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// GetProfile returns the full Profile from either the name or ID
func GetProfile(minecraft string) (*Profile, error) {
	var ret = &Profile{}

	// Try parsing minecraft as a UUID, if that fails use it as a name to lookup the UUID
	minecraftID, err := uuid.Parse(strings.TrimSpace(minecraft))
	if err == nil && minecraftID.String() != "" {
		// minecraft is an id, verify it
		var bad = echo.NewHTTPError(http.StatusBadRequest, "bad minecraft uuid")

		// Lookup minecraft name
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

		// Parse response
		type name struct {
			Name string `json:"name"`
			At   int64  `json:"changedToAt"`
		}
		var body = make([]name, 5)
		err = resp.JSON(&body)
		if err != nil {
			return nil, err
		}
		if len(body) < 1 {
			return nil, err
		}

		// Find the most recent name, this is probably body[len(body)-1] but let's explicitly check changedToAt
		newest := body[0]
		for _, it := range body {
			if it.At > newest.At {
				newest = it
			}
		}

		ret = &Profile{
			ID:   minecraftID,
			Name: newest.Name,
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
		// Response happens to use the same format as our Profile struct
		err = resp.JSON(&ret)
		if err != nil || ret.ID.String() == "" {
			return nil, bad
		}
	}

	return ret, nil
}
