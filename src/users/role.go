package users

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Role struct {
	// Role id, e.g. "developer"
	ID string `json:"id"`
	// Role rank, lower is better
	rank int
	// Is there a legacy list of pure UUIDs that old clients rely on?
	LegacyList bool
}

type roleTemplate struct {
	info    *UserInfo
	edition *Edition
}

var Roles = map[string]Role{
	"pepsi":      {ID: "pepsi", rank: 1, LegacyList: true},
	"spawnmason": {ID: "spawnmason", rank: 0, LegacyList: false},
	"developer":  {ID: "developer", rank: 2, LegacyList: true},
	"staff":      {ID: "staff", rank: 3, LegacyList: true},
	"premium":    {ID: "premium", rank: 4, LegacyList: true},
}

var defaultRoleTemplates = map[string]roleTemplate{
	"developer": {
		info: &UserInfo{
			Cape: "https://files.impactclient.net/img/texture/developer_cape.png",
		},
	},
	"staff": {
		info: &UserInfo{
			Cape: "https://files.impactclient.net/img/texture/staff_cape.png",
		},
		edition: &Edition{
			Text:      "Staff",
			TextColor: "#FF7734EB",
		},
	},
	"pepsi": {
		info: &UserInfo{
			Icon:            "https://files.impactclient.net/img/texture/pepsi32.png",
			Cape:            "https://files.impactclient.net/img/texture/pepsi_cape.png",
			TextColor:       "BLUE", // #FF004B93 is the official pepsi blue
			BackgroundColor: "#50FFFFFF",
			BorderColor:     "#FFC9002B",
		},
		edition: &Edition{
			Icon:      "https://files.impactclient.net/img/texture/pepsi32.png",
			Text:      "Pepsi",
			TextColor: "#FFC9002B",
		},
	},
	"spawnmason": {
		info: &UserInfo{
			Icon:            "https://files.impactclient.net/img/texture/spawnmason128.png",
			Cape:            "https://files.impactclient.net/img/texture/spawnmason_cape_elytra.png",
			TextColor:       "GOLD",
			BackgroundColor: "#90404040",
			BorderColor:     "RED",
		},
	},
	"premium": {
		info: &UserInfo{
			Cape: "https://files.impactclient.net/img/texture/premium_cape.png",
		},
		edition: &Edition{
			Text:      "Premium",
			TextColor: "GOLD",
		},
	},
}

func (role Role) applyDefaults(info *UserInfo) {
	t, ok := defaultRoleTemplates[role.ID]
	if !ok {
		fmt.Println("ERROR idk how to apply", role.ID)
		// No default template to apply
		return
	}
	if t.info == nil {
		return
	}

	template := t.info
	if template.Icon != "" && info.Icon == "" {
		info.Icon = template.Icon
	}
	if template.Cape != "" && info.Cape == "" {
		info.Cape = template.Cape
	}
	if template.TextColor != "" && info.TextColor == "" {
		info.TextColor = template.TextColor
	}
	if template.BackgroundColor != "" && info.BackgroundColor == "" {
		info.BackgroundColor = template.BackgroundColor
	}
	if template.BorderColor != "" && info.BorderColor == "" {
		info.BorderColor = template.BorderColor
	}
}

func getRolesSorted(roles []Role) (sorted []Role) {
	// needed so that higher priority roles set cape and icon instead of lower priority ones
	// copying slices via = is by reference, so use append instead
	// https://github.com/go101/go101/wiki/How-to-perfectly-clone-a-slice%3F
	sorted = append(roles[:0:0], roles...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].rank < sorted[j].rank
	})
	return
}

// MarshalJSON implements the json.Marshaler interface
// it marshals the role to just the role id as a json string
func (role Role) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, role.ID)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
// it allows unmarshalling a full role from just the role id json string
func (role *Role) UnmarshalJSON(bytes []byte) error {
	var id string
	err := json.Unmarshal(bytes, &id)
	if err != nil {
		return err
	}
	if r, ok := Roles[id]; ok {
		role.ID = r.ID
		role.rank = r.rank
		role.LegacyList = r.LegacyList
		return nil
	}
	return fmt.Errorf("unable to find role with id %s", string(id))
}
