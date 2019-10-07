package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapLegacyListsToUserInfoList(t *testing.T) {
	uuidToHash := map[string]string{
		"foo":  hashUUID("foo"),
		"bar":  hashUUID("bar"),
		"baz":  hashUUID("baz"),
		"fred": hashUUID("fred"),
	}
	testData := map[string][]string{
		"developer": {"foo"},
		"staff":     {"foo", "bar"},
		"pepsi":     {"fred"},
		"premium":   {"fred", "baz"},
	}

	result := mapLegacyListsToUserInfoList(testData)

	assert.Equal(t, len(uuidToHash), len(result))
	for _, hash := range uuidToHash {
		assert.Contains(t, result, hash)
	}

	foo := result[uuidToHash["foo"]]
	fooRoleIDs := []string{"developer", "staff"}
	assert.Equal(t, len(fooRoleIDs), len(foo.Roles))
	for _, role := range foo.Roles {
		assert.Contains(t, fooRoleIDs, role.ID)
	}

	bar := result[uuidToHash["bar"]]
	barRoleIDs := []string{"staff"}
	assert.Equal(t, len(barRoleIDs), len(bar.Roles))
	for _, role := range bar.Roles {
		assert.Contains(t, barRoleIDs, role.ID)
	}

	baz := result[uuidToHash["baz"]]
	bazRoleIDs := []string{"premium"}
	assert.Equal(t, len(bazRoleIDs), len(baz.Roles))
	for _, role := range baz.Roles {
		assert.Contains(t, bazRoleIDs, role.ID)
		println("Found role", role.ID)
	}

	fred := result[uuidToHash["fred"]]
	fredRoleIDs := []string{"premium", "pepsi"}
	assert.Equal(t, len(fredRoleIDs), len(fred.Roles))
	for _, role := range fred.Roles {
		assert.Contains(t, fredRoleIDs, role.ID)
	}

}

func hasRole(info *userInfo, id string) bool {
	for _, role := range info.Roles {
		if role.ID == id {
			return true
		}
	}
	return false
}
