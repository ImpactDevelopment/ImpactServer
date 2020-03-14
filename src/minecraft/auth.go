package minecraft

import (
	"errors"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"strings"
)

const urlHasJoined = "https://sessionserver.mojang.com/session/minecraft/hasJoined"

func HasJoinedServer(username, hash string) (*Profile, error) {
	req, err := util.GetRequest(urlHasJoined)
	if err != nil {
		return nil, err
	}

	req.SetQuery("username", username)
	req.SetQuery("serverId", "0"+hash)

	resp, err := req.Do()
	if err != nil {
		return nil, err
	}

	var profile Profile
	err = resp.JSON(&profile)
	if err != nil {
		return nil, err
	}

	match := strings.EqualFold(username, profile.Name)
	if !match {
		return nil, errors.New("invalid username")
	}

	return &profile, nil
}
