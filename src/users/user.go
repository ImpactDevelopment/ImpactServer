package users

import (
	"github.com/google/uuid"
)

type User struct {
	Email         string
	MinecraftID   *uuid.UUID
	DiscordID     string
	PasswordHash  string
	LegacyEnabled bool
	Incognito     bool
	Legacy        bool
	Roles         []Role
	UserInfo      *UserInfo
}

func (user User) RoleIDs() []string {
	roles := user.Roles
	arr := make([]string, len(roles))
	for i, role := range roles {
		arr[i] = role.ID
	}
	return arr
}

func (user User) HasRoleWithID(roleID string) bool {
	for _, role := range user.Roles {
		if role.ID == roleID {
			return true
		}
	}
	return false
}

// IsFullAccount returns true if the user is a full Impact Account
func (user User) IsFullAccount() bool {
	return user.Email != ""
}

// CheckPassword returns true if the password is correct
func (user User) CheckPassword(password string) bool {
	if !user.IsFullAccount() {
		return false
	}
	hash := password // TODO actually hash passwords
	return user.PasswordHash == hash
}
