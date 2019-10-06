package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

type (
	role struct {
		// Role id, e.g. "developer"
		ID string `json:"id"`
		// Role rank, lower is better
		Rank int `json:"-"`
	}
	userInfo struct {
		// A list of roles applicable to this user
		Roles []role `json:"roles"`
		// Icon to display next to this user
		Icon string `json:"icon,omitempty"`
		// Cape this user should wear
		Cape string `json:"cape,omitempty"`
	}
)

// API Handler
func getUserInfo(c echo.Context) error {
	lists, err := getLegacyUuidLists()
	if err != nil {
		return err
	}

	legacyUsers := mapLegacyListsToUserInfoList(lists)

	return c.JSON(http.StatusOK, legacyUsers)
}

func hashUUID(uuid string) string {
	hash := sha256.Sum256([]byte(uuid))
	return hex.EncodeToString(hash[:])
}

// Get each of the legacy uuid lists as a map of role -> list
func getLegacyUuidLists() (lists map[string][]string, err error) {
	urls := map[string]string{
		"developer": "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/developer.txt",
		"staff":     "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/staff.txt",
		"pepsi":     "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/pepsi.txt",
		"premium":   "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/premium.txt",
	}

	// Make a map the same length as urls
	lists = make(map[string][]string, len(urls))

	// Dump the response from each url into the lists map
	for key, url := range urls {
		res, err := http.Get(url)
		if err != nil {
			// Hm, error getting one of the urls
			fmt.Println("Error getting", key, err.Error())
			continue
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			// wtf
			fmt.Println("Error getting", key, res.StatusCode)
			continue
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error reading response for", key, err.Error())
			continue
		}

		lists[key] = strings.Split(string(body), "\n")
	}
	return
}

// Convert a [roleID][]uuid map to a [hashedUUID]role map
func mapLegacyListsToUserInfoList(lists map[string][]string) (info map[string]*userInfo) {
	defaults := map[string]userInfo{
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

	info = make(map[string]*userInfo, sumLists(lists))
	for key := range lists {
		for _, line := range lists[key] {
			// Send a hash of the uuid, not the uuid itself
			// to make it harder to just bulk-ban users
			hash := hashUUID(line)

			if _, ok := info[hash]; !ok {
				// New user, copy the default info for this key
				defaultInfo := defaults[key]
				info[hash] = &defaultInfo
			} else {
				// Existing user, add roles
				for _, role := range defaults[key].Roles {
					err := info[hash].AddRole(role)
					if err != nil {
						continue
					}

					// If this role outranks the others, override capes and icons
					if info[hash].IsHighest(role) {
						cape := defaults[key].Cape
						icon := defaults[key].Icon
						if cape != "" {
							info[hash].SetCape(cape)
						}
						if icon != "" {
							info[hash].SetIcon(icon)
						}
					}
				}
			}
		}
	}
	return
}

func sumLists(m map[string][]string) (sum int) {
	sum = 0
	for key := range m {
		sum += len(m[key])
	}
	return
}

// True if the role, or a conflicting role, is in the slice
func containsMatch(s []role, e role) bool {
	for _, a := range s {
		if a == e {
			return true
		}
		if a.Rank == e.Rank {
			return true
		}
		if a.ID == e.ID {
			return true
		}
	}
	return false
}

// Add a role to a userInfo.
func (info *userInfo) AddRole(r role) error {
	if containsMatch(info.Roles, r) {
		return errors.New("userInfo.AddRole: conflicting role already exists")
	}

	info.Roles = append(info.Roles, r)
	return nil
}

func (info *userInfo) SetCape(cape string) {
	info.Cape = cape
}

func (info *userInfo) SetIcon(icon string) {
	info.Icon = icon
}

func (info userInfo) IsHighest(r role) bool {
	roles := info.Roles
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Rank < roles[j].Rank
	})
	return roles[0] == r
}
