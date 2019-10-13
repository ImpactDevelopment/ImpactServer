package users

import "github.com/google/uuid"

type LegacyUser struct {
	uuids []uuid.UUID
	roles []Role
}

func (user LegacyUser) MinecraftIDs() []uuid.UUID {
	return user.uuids
}

func (user LegacyUser) Roles() []Role {
	return user.roles
}

func (user LegacyUser) UserInfo() (info *UserInfo) {
	info = &UserInfo{}
	for _, role := range getRolesSorted(user.Roles()) {
		role.applyDefaults(info)
	}
	return
}

func (user LegacyUser) IsLegacy() bool {
	return true
}
