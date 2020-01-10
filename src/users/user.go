package users

import "github.com/google/uuid"

type User interface {
	Email() *string
	MinecraftID() *uuid.UUID
	DiscordID() *string
	Roles() []Role
	UserInfo() UserInfo
	IsLegacy() bool
	CheckPassword(password string) bool
	HasPassword() bool
}
