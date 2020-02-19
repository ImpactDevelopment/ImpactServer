package database

import (
	"database/sql"
	"fmt"

	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/google/uuid"
)

type userRow struct {
	id            uuid.UUID
	email         sql.NullString
	minecraft     NullUUID
	discord       sql.NullString
	passwdHash    sql.NullString
	legacyEnabled bool
	capeEnabled   bool
	premium       bool
	pepsi         bool
	spawnmason    bool
	staff         bool
	developer     bool
	legacy        bool
}

// rowScanner is implemented by sql.Row and sql.Rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanUsersView takes a sql.Row or sql.Rows and scans it into the user.
// It is assumed the row is has the same column order as `users_view`
func (user *userRow) scanUsersView(row rowScanner) error {
	return row.Scan(&user.id, &user.email, &user.minecraft, &user.discord, &user.passwdHash, &user.capeEnabled, &user.legacyEnabled, &user.legacy, &user.premium, &user.pepsi, &user.spawnmason, &user.staff, &user.developer)
}

// makeUser converts a userRow into a users.User
func (user *userRow) makeUser() users.User {
	ret := users.User{
		LegacyEnabled: user.legacyEnabled,
		Incognito:     !user.capeEnabled,
		Legacy:        user.legacy,
		Roles:         user.roles(),
	}
	if user.email.Valid {
		ret.Email = user.email.String
	}
	if user.minecraft.Valid {
		ret.MinecraftID = &user.minecraft.UUID
	}
	if user.discord.Valid {
		ret.DiscordID = user.discord.String
	}
	if user.passwdHash.Valid {
		ret.PasswordHash = user.passwdHash.String
	}
	ret.UserInfo = users.NewUserInfo(ret)
	ret.ID = user.id
	return ret
}

func (user userRow) roles() []users.Role {
	var roles []users.Role
	if user.premium {
		roles = append(roles, users.Roles["premium"])
	}
	if user.staff {
		roles = append(roles, users.Roles["staff"])
	}
	if user.pepsi {
		roles = append(roles, users.Roles["pepsi"])
	}
	if user.spawnmason {
		roles = append(roles, users.Roles["spawnmason"])
	}
	if user.developer {
		roles = append(roles, users.Roles["developer"])
	}
	return roles
}

// GetAllUsers returns... all the users
func GetAllUsers() []users.User {
	if DB == nil {
		fmt.Println("Database not connected!")
		return nil
	}

	rows, err := DB.Query(`SELECT * FROM users_view`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	ret := make([]users.User, 0)
	for rows.Next() {
		var r userRow
		err = r.scanUsersView(rows)
		if err != nil {
			panic(err)
		}
		ret = append(ret, r.makeUser())
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return ret
}

// LookupUserByMinecraftID returns the matching user, or nil if not found
func LookupUserByID(id uuid.UUID) *users.User {
	return lookupUserByField("user_id", id)
}

func LookupUserByEmail(email string) *users.User {
	return lookupUserByField("email", email)
}

// LookupUserByMinecraftID returns the matching user, or nil if not found
func LookupUserByMinecraftID(minecraftID uuid.UUID) *users.User {
	return lookupUserByField("mc_uuid", minecraftID)
}

// LookupUserByDiscordID returns the matching user, or nil if not found
func LookupUserByDiscordID(discordID string) *users.User {
	return lookupUserByField("discord_id", discordID)
}

func lookupUserByField(field string, value interface{}) *users.User {
	if DB == nil {
		fmt.Println("Database not connected!")
		return nil
	}

	var r userRow
	err := r.scanUsersView(DB.QueryRow(`SELECT * FROM users_view WHERE `+field+` = $1`, value))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // no match
		}
		panic(err) // any other error
	}

	user := r.makeUser()
	return &user
}
