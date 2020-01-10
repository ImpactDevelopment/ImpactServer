package users

import (
	"database/sql"
	"fmt"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/google/uuid"
)

type User struct {
	email         sql.NullString
	mcUUID        database.NullUUID
	dcID          sql.NullString
	passwdHash    sql.NullString
	legacyPremium bool
	capeEnabled   bool
	premium       bool
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
	"developer": {ID: "developer", rank: 0},
	"pepsi":     {ID: "pepsi", rank: 1},
	"staff":     {ID: "staff", rank: 2},
	"premium":   {ID: "premium", rank: 3},
}

func (user User) Email() *string {
	if user.email.Valid {
		return &user.email.String
	} else {
		return nil
	}
}

func (user User) MinecraftID() *uuid.UUID {
	if user.mcUUID.Valid {
		return &user.mcUUID.UUID
	} else {
		return nil
	}
}

func (user User) DiscordID() *string {
	if user.dcID.Valid {
		return &user.dcID.String
	} else {
		return nil
	}
}

func (user User) Roles() []Role {
	roles := []Role{}
	if user.premium {
		roles = append(roles, RolesData["premium"])
	}
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

func (user User) UserInfo() UserInfo {
	info := UserInfo{}

	if special, ok := specialCases[user.mcUUID.UUID]; ok {
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

func (user User) IsLegacy() bool {
	return user.legacyPremium
}

func (user User) CheckPassword(password string) bool {
	if !user.HasPassword() {
		return false
	}
	hash := password // TODO actually hash passwords
	return user.passwdHash.String == hash
}

func (user User) HasPassword() bool {
	return user.passwdHash.Valid
}

func GetAllUsers() []User {
	if database.DB == nil {
		fmt.Println("Database not connected!")
		return nil
	}
	rows, err := database.DB.Query(selectString())
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	ret := make([]User, 0)
	for rows.Next() {
		var user User
		err = rows.Scan(scanDest(&user)...)
		if err != nil {
			panic(err)
		}
		ret = append(ret, user)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return ret
}

func LookupUserByMinecraftID(uuid uuid.UUID) *User {
	if database.DB == nil {
		fmt.Println("Database not connected!")
		return nil
	}
	var user User
	err := database.DB.QueryRow(selectWhereString(`mc_uuid = $1`), uuid).Scan(scanDest(&user)...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // no match
		}
		panic(err) // any other error
	}
	return &user
}

func selectString() string {
	return `SELECT email, mc_uuid, discord_id, password_hash, legacy_premium, cape_enabled, premium, pepsi, staff, developer FROM users`
}
func selectWhereString(where string) string {
	return selectString() + ` WHERE ` + where
}
func scanDest(user *User) []interface{} {
	return []interface{}{&user.email, &user.mcUUID, &user.dcID, &user.passwdHash, &user.legacyPremium, &user.capeEnabled, &user.premium, &user.pepsi, &user.staff, &user.developer}
}
