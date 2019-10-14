package users

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type legacyGithubUser struct {
	uuid  uuid.UUID
	roles []Role
}

func (user legacyGithubUser) MinecraftIDs() []uuid.UUID {
	return []uuid.UUID{user.uuid}
}

func (user legacyGithubUser) Roles() []Role {
	return user.roles
}

func (user legacyGithubUser) UserInfo() (info UserInfo) {
	info = UserInfo{}
	for _, role := range getRolesSorted(user.Roles()) {
		role.applyDefaults(&info)
	}
	return
}

func (user legacyGithubUser) IsLegacy() bool {
	return true
}

var legacyGithubData map[uuid.UUID][]Role
var legacyGithubDataLock sync.Mutex

func GetAllUsers() []User {
	legacyGithubDataLock.Lock()
	defer legacyGithubDataLock.Unlock()

	ret := make([]User, 0)
	for uuid, roles := range legacyGithubData {
		ret = append(ret, &legacyGithubUser{uuid: uuid, roles: roles})
	}
	return ret
}

func LookupUserByUUID(uuid uuid.UUID) User {
	legacyGithubDataLock.Lock()
	defer legacyGithubDataLock.Unlock()

	roles, ok := legacyGithubData[uuid]
	if ok {
		return &legacyGithubUser{uuid: uuid, roles: roles}
	}
	return nil
}

func UpdateLegacyData() error {
	fmt.Println("Fetching legacy data from github")
	legacyGithubDataLock.Lock()
	defer legacyGithubDataLock.Unlock()

	data, err := generateLegacyData()
	if err != nil {
		return err
	}
	legacyGithubData = data
	fmt.Println("Fetched", len(data), "legacy users from github")
	return nil
}

func generateLegacyData() (map[uuid.UUID][]Role, error) {
	rolesData := map[string]Role{
		"developer": Role{ID: "developer", rank: 0},
		"pepsi":     Role{ID: "pepsi", rank: 1},
		"staff":     Role{ID: "staff", rank: 2},
		"premium":   Role{ID: "premium", rank: 3},
	}
	data := make(map[uuid.UUID][]Role)
	roleToUsers, err := getLegacyUUIDLists()
	if err != nil {
		return nil, err
	}
	for roleName, uuids := range roleToUsers {
		for _, uuid := range uuids {
			data[uuid] = append(data[uuid], rolesData[roleName]) // <-- unironically beautiful
		}
	}
	return data, nil
}

// Get each of the legacy uuid lists as a map of role -> list
func getLegacyUUIDLists() (map[string][]uuid.UUID, error) {
	urls := map[string]string{
		"developer": "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/developer.txt",
		"staff":     "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/staff.txt",
		"pepsi":     "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/pepsi.txt",
		"premium":   "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/data/users/premium.txt",
	}

	// Make a map the same length as urls
	lists := make(map[string][]uuid.UUID, len(urls))

	// Dump the response from each url into the lists map
	for key, url := range urls {
		res, err := http.Get(url)
		if err != nil {
			// Hm, error getting one of the urls
			fmt.Println("Error getting", key, err.Error())
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			// wtf
			fmt.Println("Error getting", key, res.StatusCode)
			return nil, err
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error reading response for", key, err.Error())
			return nil, err
		}

		uuidStrs := strings.Split(strings.TrimSpace(string(body)), "\n")
		uuids := make([]uuid.UUID, 0, len(uuidStrs))
		for _, uuidStr := range uuidStrs {
			uuid, err := uuid.Parse(strings.TrimSpace(uuidStr))
			if err != nil {
				fmt.Println("Invalid line from github, ignoring!")
				fmt.Println(uuidStr)
				fmt.Println(err)
				continue
			}
			uuids = append(uuids, uuid)
		}
		lists[key] = uuids
	}
	return lists, nil
}
