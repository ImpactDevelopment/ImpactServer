package users

import "github.com/google/uuid"

type User interface {
	MinecraftIDs() []uuid.UUID
	Roles() []Role
	UserInfo() UserInfo
	IsLegacy() bool
}
