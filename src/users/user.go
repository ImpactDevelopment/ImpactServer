package users

import (
	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `json:"-"`
	Email         string     `json:"email"`
	MinecraftID   *uuid.UUID `json:"minecraft"`
	DiscordID     string     `json:"discord"`
	PasswordHash  string     `json:"-"`
	LegacyEnabled bool       `json:"legacy_enabled"`
	Incognito     bool       `json:"incognito"`
	Legacy        bool       `json:"legacy"`
	Roles         []Role     `json:"roles"`
	UserInfo      *UserInfo  `json:"user_info"`
}

func (user User) RoleIDs(legacyOnly bool) []string {
	roles := user.Roles
	arr := make([]string, 0)
	for _, role := range roles {
		if !role.LegacyList && legacyOnly {
			continue
		}
		arr = append(arr, role.ID)
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
