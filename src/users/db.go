package users

import (
	"database/sql"
	"fmt"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/google/uuid"
)

type databaseUser struct {
	uuid          uuid.UUID
	legacyPremium bool
	capeEnabled   bool
	pepsi         bool
	staff         bool
	developer     bool
}

var specialCases = map[uuid.UUID]UserInfo{ // TODO this should basically just be a SELECT * FROM customizations;
	// catgorl
	uuid.MustParse("2c3174fc-0c6b-4cfb-bb2b-0069bf7294d1"): {
		TextColor: "LIGHT_PURPLE",
	},
	// leijurv
	uuid.MustParse("51dcd870-d33b-40e9-9fc1-aecdcff96081"): {
		TextColor: "RED",
		Icon:      "https://files.impactclient.net/img/texture/speckles32.png",
	},
}

var RolesData = map[string]Role{
	"developer": Role{ID: "developer", rank: 0},
	"pepsi":     Role{ID: "pepsi", rank: 1},
	"staff":     Role{ID: "staff", rank: 2},
	"premium":   Role{ID: "premium", rank: 3},
}

func (user databaseUser) MinecraftIDs() []uuid.UUID {
	return []uuid.UUID{user.uuid}
}

func (user databaseUser) Roles() []Role {
	if user.uuid.String() == "2c3174fc-0c6b-4cfb-bb2b-0069bf7294d1" {
		// stupid edge case: catgorl has a custom name, but no premium cape
		// this breaks the assumption of "everyone in the database has premium"

		// i'm only adding this check so that the new database-backed api is ExAcTlY iDeNtIcAl to the old github-backed one

		return nil
	}
	roles := []Role{RolesData["premium"]}
	if user.staff {
		roles = append(roles, RolesData["staff"])
	}
	if user.pepsi {
		roles = append(roles, RolesData["pepsi"])
	}
	if user.developer {
		roles = append(roles, RolesData["developer"])
	}
	return roles
}

func (user databaseUser) UserInfo() UserInfo {
	info := UserInfo{}

	if special, ok := specialCases[user.uuid]; ok {
		info = special
	}

	for _, role := range getRolesSorted(user.Roles()) { // go in order from highest priority to least (aka numerically lowest to highest)
		role.applyDefaults(&info)
	}

	if !user.capeEnabled {
		info.Cape = ""
	}

	return info
}

func (user databaseUser) IsLegacy() bool {
	return user.legacyPremium
}

func GetAllUsers() []User {
	if database.DB == nil {
		fmt.Println("Database not connected!")
		return nil
	}
	rows, err := database.DB.Query("SELECT mc_uuid, legacy_premium, cape_enabled, pepsi, staff, developer FROM users WHERE mc_uuid IS NOT NULL")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	ret := make([]User, 0)
	for rows.Next() {
		var user databaseUser
		err = rows.Scan(&user.uuid, &user.legacyPremium, &user.capeEnabled, &user.pepsi, &user.staff, &user.developer)
		if err != nil {
			panic(err)
		}
		ret = append(ret, &user)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return ret
}

func LookupUserByUUID(uuid uuid.UUID) User {
	if database.DB == nil {
		fmt.Println("Database not connected!")
		return nil
	}
	var user databaseUser
	err := database.DB.QueryRow("SELECT mc_uuid, legacy_premium, cape_enabled, pepsi, staff, developer FROM users WHERE mc_uuid = $1", uuid).Scan(&user.uuid, &user.legacyPremium, &user.capeEnabled, &user.pepsi, &user.staff, &user.developer)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // no match
		}
		panic(err) // any other error
	}
	return &user
}
