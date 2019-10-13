package users

import "github.com/google/uuid"

type User interface {
	MinecraftIDs() []uuid.UUID
	Roles() []Role
	UserInfo() *UserInfo
	IsLegacy() bool
}

type NormalUser struct {
	uuids []uuid.UUID
	roles []Role
	info  *UserInfo
}

func (user NormalUser) MinecraftIDs() []uuid.UUID {
	return user.uuids
}

func (user NormalUser) Roles() []Role {
	return user.roles
}

func (user NormalUser) UserInfo() *UserInfo {
	return user.info
}

func (user NormalUser) IsLegacy() bool {
	return false
}
