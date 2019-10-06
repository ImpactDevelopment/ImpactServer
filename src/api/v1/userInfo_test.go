package v1

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserInfo_AddRole(t *testing.T) {
	info := userInfo{}
	dupRole := role{ID: "pretty cool", Rank: 1}

	assert.Equal(t, 0, len(info.Roles))

	if assert.NoError(t, info.AddRole(dupRole)) {
		assert.Equal(t, 1, len(info.Roles))
	}

	if assert.NoError(t, info.AddRole(role{ID: "higher than high", Rank: 0})) {
		assert.Equal(t, 2, len(info.Roles))
	}

	// Should not add duplicates
	if assert.Error(t, info.AddRole(dupRole)) {
		assert.Equal(t, 2, len(info.Roles))
	}
	// Or conflicting ranks
	if assert.Error(t, info.AddRole(role{Rank: dupRole.Rank})) {
		assert.Equal(t, 2, len(info.Roles))
	}
	// Or conflicting IDs
	if assert.Error(t, info.AddRole(role{ID: dupRole.ID})) {
		assert.Equal(t, 2, len(info.Roles))
	}

	if assert.NoError(t, info.AddRole(role{ID: "fairly average", Rank: 2})) {
		assert.Equal(t, 3, len(info.Roles))
	}
}

func TestUserInfo_IsHighest(t *testing.T) {
	highest := role{Rank: 1, ID: "high"}
	newHighest := role{Rank: 0, ID: "new"}
	info := userInfo{Roles: []role{
		{Rank: 5},
		{Rank: 4},
		{Rank: 2},
		highest,
		{Rank: 3},
	}}

	assert.True(t, info.IsHighest(highest))
	if assert.NoError(t, info.AddRole(newHighest)) {
		assert.True(t, info.IsHighest(newHighest))
	}
}

func TestUserInfo_SetCape(t *testing.T) {
	info := userInfo{}

	assert.Empty(t, info.Cape)

	info.SetCape("foo")
	assert.Equal(t, "foo", info.Cape)

	info.SetCape("bar")
	assert.Equal(t, "bar", info.Cape)

	info.SetCape("")
	assert.Empty(t, info.Cape)
}

func TestUserInfo_SetIcon(t *testing.T) {
	info := userInfo{}

	assert.Empty(t, info.Icon)

	info.SetIcon("foo")
	assert.Equal(t, "foo", info.Icon)

	info.SetIcon("bar")
	assert.Equal(t, "bar", info.Icon)

	info.SetIcon("")
	assert.Empty(t, info.Icon)
}
