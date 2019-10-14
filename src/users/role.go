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
}

var defaultRoleTemplates = map[string]UserInfo{
	"developer": {
		Cape: "http://i.imgur.com/X9NYKct.png",
	},
	"staff": {
		Cape: "http://i.imgur.com/uh6QcuF.png",
	},
	"pepsi": {
		Icon: "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/textures/Pepsi_32.png",
		Cape: "http://i.imgur.com/SKjRGbH.png",
	},
	"premium": {
		Cape: "http://i.imgur.com/fc8gsyN.png",
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
		info.SetIcon(template.Icon)
	}
	if template.Cape != "" && info.Cape == "" {
		info.SetCape(template.Cape)
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
