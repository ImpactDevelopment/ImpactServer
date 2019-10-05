package v1

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

type role struct {
	// Role id, e.g. "developer"
	ID string `json:"id"`
	// Role rank, lower is better
	Rank int
}

type userinfo struct {
	// A list of roles applicable to this user
	Roles []role `json:"roles"`
	// Icon to display next to this user
	Icon string `json:"icon,omitempty"`
	// Cape this user should wear
	Cape string `json:"cape,omitempty"`
}

// API Handler
func userInfo(c echo.Context) error {
	res, err := getFromLegacy()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func hashUUID(uuid string) string {
	hash := sha256.Sum256([]byte(uuid))
	return hex.EncodeToString(hash[:])
}

func getFromLegacy() (info map[string]*userinfo, err error) {
	urls := map[string]string{
		"developer": "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/developer.txt",
		"staff":     "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/staff.txt",
		"pepsi":     "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/pepsi.txt",
		"premium":   "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/premium.txt",
	}
	defaults := map[string]userinfo{
		"developer": {
			Roles: []role{{ID: "developer", Rank: 0}},
			Cape:  "http://i.imgur.com/X9NYKct.png",
		},
		"staff": {
			Roles: []role{{ID: "staff", Rank: 2}},
			Cape:  "http://i.imgur.com/uh6QcuF.png",
		},
		"pepsi": {
			Roles: []role{{ID: "pepsi", Rank: 1}},
			Icon:  "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/textures/Pepsi_32.png",
			Cape:  "http://i.imgur.com/SKjRGbH.png",
		},
		"premium": {
			Roles: []role{{ID: "premium", Rank: 3}},
			Cape:  "http://i.imgur.com/fc8gsyN.png",
		},
	}
	lists := make(map[string]string, len(urls))
	for key, url := range urls {
		res, err := http.Get(url)
		if err != nil {
			// Hm, error getting one of the urls
			fmt.Println("Error getting", key, err.Error())
			continue
		}
		if res.StatusCode != http.StatusOK {
			// wtf
			fmt.Println("Error getting", key, res.StatusCode)
			continue
		}
		body, _ := ioutil.ReadAll(res.Body)
		lists[key] = string(body)
	}

	info = make(map[string]*userinfo, sumLines(lists))

	for key := range lists {
		forEachLine(lists[key], func(line string) {
			// Send a hash of the uuid, not the uuid itself
			// to make it harder to just bulk-ban users
			hash := hashUUID(line)

			if _, ok := info[hash]; ok {
				// Existing user, add role and override cape/icon
				role := defaults[key].Roles[0]
				cape := defaults[key].Cape
				icon := defaults[key].Icon
				info[hash].AddRole(role, cape, icon)
			} else {
				// New user, copy the default info for this key
				defaultInfo := defaults[key]
				info[hash] = &defaultInfo
			}

		})
	}
	return
}

func sumLines(m map[string]string) (sum int) {
	sum = 0
	for key := range m {
		forEachLine(m[key], func(line string) {
			sum++
		})
	}
	return
}

func contains(s []role, e role) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func forEachLine(lines string, f func(line string)) {
	scanner := bufio.NewScanner(strings.NewReader(lines))
	for scanner.Scan() {
		f(scanner.Text())
	}
}

// Add a role to a userinfo.
// If the role is the highest, also set the cape and icon.
// Cape and icon will each only be set if not empty
func (info *userinfo) AddRole(r role, cape, icon string) {
	if contains(info.Roles, r) {
		fmt.Println("Warning tried adding role", r.ID, "to user twice")
		return
	}
	info.Roles = append(info.Roles, r)

	// Sort roles by rank
	sort.Slice(info.Roles, func(i, j int) bool {
		return info.Roles[i].Rank < info.Roles[j].Rank
	})

	// If we just added the highest role, also set cape and icon
	if info.Roles[0] == r {
		if cape != "" {
			info.Cape = cape
		}
		if icon != "" {
			info.Icon = icon
		}
	}
}
