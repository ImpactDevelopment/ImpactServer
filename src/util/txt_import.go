package util

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/google/uuid"
)

// cd Resources/data/users
// git blame --line-porcelain *.txt | grep -e "\t" -e "committer-time" | tail -r | tr "\n" " " | tr "\t" "\n" | sed -e '$a\' | tail -n +2 | tail -r | sed "s/ committer-time//g" > blame.txt

func ImportFromBlameAge() {
	data, err := ioutil.ReadFile("blame.txt")
	if err != nil {
		panic(err)
	}
	dates := make(map[uuid.UUID]int64) // there are duplicates (elmo)
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue // oh you silly last line
		}
		uuidStr := strings.Split(line, " ")[0]
		epochStr := strings.Split(line, " ")[1]
		uuidVal := uuid.MustParse(uuidStr)
		epochVal, err := strconv.ParseInt(epochStr, 10, 64) // base 10, 64 bit. rofl
		if err != nil {
			panic(err)
		}
		if dates[uuidVal] == 0 || epochVal < dates[uuidVal] {
			dates[uuidVal] = epochVal
		}
	}
	for uuidVal, epochVal := range dates {
		fmt.Println(uuidVal)
		_, err := database.DB.Exec("INSERT INTO users (mc_uuid, created_at) VALUES ($1, $2)", uuidVal, epochVal)
		if err != nil {
			panic(err)
		}
	}
}

func ImportFromRoles() {
	//importFromRole("pepsi")
	//importFromRole("developer")
	//importFromRole("staff")
	//importFromRole("premium")
	importFromRole("spawnmason")
}

func importFromRole(role string) {
	data, err := ioutil.ReadFile(role + ".txt")
	if err != nil {
		panic(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue // oh you silly last line
		}
		fmt.Println(line)
		// delibrately ignore duplicate errors lol
		database.DB.Exec(`INSERT INTO users(mc_uuid) VALUES ($1)`, line)
		_, err = database.DB.Exec(`UPDATE users SET roles = array_append(roles, $2) WHERE mc_uuid = $1`, line, role)
		if err != nil {
			panic(err)
		}
	}
}
