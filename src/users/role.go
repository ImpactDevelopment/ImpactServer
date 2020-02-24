package users

import (
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

var Roles = map[string]Role{
	"pepsi":      {ID: "pepsi", rank: 1, LegacyList: true},
	"spawnmason": {ID: "spawnmason", rank: 0, LegacyList: false},
	"developer":  {ID: "developer", rank: 2, LegacyList: true},
	"staff":      {ID: "staff", rank: 3, LegacyList: true},
	"premium":    {ID: "premium", rank: 4, LegacyList: true},
}

var defaultRoleTemplates = map[string]UserInfo{
	"developer": {
		Cape: "https://files.impactclient.net/img/texture/developer_cape.png",
	},
	"staff": {
		Cape: "https://files.impactclient.net/img/texture/staff_cape.png",
		Editions: []Edition{{
			Text: "Staff",
		}},
	},
	"pepsi": {
		Icon:            "https://files.impactclient.net/img/texture/pepsi32.png",
		Cape:            "https://files.impactclient.net/img/texture/pepsi_cape.png",
		TextColor:       "BLUE", // #FF004B93 is the official pepsi blue
		BackgroundColor: "#50FFFFFF",
		BorderColor:     "#FFC9002B",
		Editions: []Edition{{
			Text:      "Pepsi",
			TextColor: "#FFC9002B",
		}},
	},
	"spawnmason": {
		Icon:            "https://files.impactclient.net/img/texture/spawnmason128.png",
		Cape:            "https://files.impactclient.net/img/texture/spawnmason_cape_elytra.png",
		TextColor:       "GOLD",
		BackgroundColor: "#90404040",
		BorderColor:     "RED",
	},
	"premium": {
		Cape: "https://files.impactclient.net/img/texture/premium_cape.png",
		Editions: []Edition{{
			Text:      "Premium",
			TextColor: "GOLD",
		}},
	},
}

func (role Role) applyDefaults(info *UserInfo) {
	template, ok := defaultRoleTemplates[role.ID]
	if !ok {
		fmt.Println("ERROR idk how to apply", role.ID)
		// No default template to apply
		return
	}
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
	if len(template.Editions) > 0 {
		info.Editions = append(info.Editions, template.Editions...)
	}
}

func getRolesSorted(roles []Role) (sorted []Role) {
	// needed so that higher priority roles set cape and icon instead of lower priority ones
	// copying slices via = is by reference, so this mutates no matter what
	sorted = roles
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].rank < sorted[j].rank
	})
	return
}
