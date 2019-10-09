package util

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const urlHasJoined = "https://sessionserver.mojang.com/session/minecraft/hasJoined"

type ResponseHasJoined struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func HasJoinedServer(username, hash string) (string, error) {
	req, err := http.NewRequest("GET", urlHasJoined, nil)
	if err != nil {
		return "", err
	}

	query := req.URL.Query()
	query.Add("username", username)
	query.Add("serverId", "0"+hash)
	req.URL.RawQuery = query.Encode()

	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	data := ResponseHasJoined{}
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	match := strings.EqualFold(username, data.Name)
	if !match {
		return "", errors.New("invalid username")
	}

	return data.Id, nil
}
