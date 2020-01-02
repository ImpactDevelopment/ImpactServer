package util

import (
	"errors"
	"strings"
)

const urlHasJoined = "https://sessionserver.mojang.com/session/minecraft/hasJoined"

type ResponseHasJoined struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func HasJoinedServer(username, hash string) (string, error) {
	req, err := GetRequest(urlHasJoined)
	if err != nil {
		return "", err
	}

	req.SetQuery("username", username)
	req.SetQuery("serverId", "0"+hash)

	response, err := req.Do()
	if err != nil {
		return "", err
	}

	data := ResponseHasJoined{}
	err = response.JSON(data)
	if err != nil {
		return "", err
	}

	match := strings.EqualFold(username, data.Name)
	if !match {
		return "", errors.New("invalid username")
	}

	return data.Id, nil
}
