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

func (user User) HasRoleWithID(roleID ...string) bool {
	for _, role := range user.Roles {
		for _, id := range roleID {
			if role.ID == id {
				return true
			}
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

// templates returns all the user's role templates, including any special cases, in order
// TODO consider if low-index-high-rank makes sense or not for templates;
//     this only seems to matter in NewUserInfo(User) and User.Editions(), so
//     if we are happy for editions to be reversed too, that'd simplify the for
//     loop in NewUserInfo a little bit.
func (user User) templates() (templates []roleTemplate) {
	if user.MinecraftID != nil {
		if special, ok := specialCases[*user.MinecraftID]; ok {
			templates = append(templates, special)
		}
	}
	for _, role := range getRolesSorted(user.Roles) {
		if template, ok := defaultRoleTemplates[role.ID]; ok {
			templates = append(templates, template)
		}
	}
	return
}
