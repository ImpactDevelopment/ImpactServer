package users

import "github.com/google/uuid"

type UserInfo struct {
	// Icon to display next to this user
	Icon string `json:"icon,omitempty"`
	// Cape this user should wear
	Cape string `json:"cape,omitempty"`
	// Color code of the text for nametags. e.g. LIGHT_PURPLE or BLUE
	TextColor string `json:"text_color,omitempty"`
	// Numeric ARGB color of the nametag background. Empty string for default. e.g. 1358954495 for pepsi's light gray
	BackgroundColor string `json:"bg_color,omitempty"`
	// Numeric ARGB color of the nametag border. Empty string for default. e.g. -1761673216 for pepsi's red
	BorderColor string `json:"border_color,omitempty"`
	// A list of editions for use in e.g. watermarks
	Editions []Edition `json:"editions,omitempty"`
}

// NewUserInfo creates a UserInfo based on a User's roles and any special cases that apply to them
func NewUserInfo(user User) *UserInfo {
	var info UserInfo

	if user.MinecraftID != nil {
		if special, ok := specialCases[*user.MinecraftID]; ok {
			info = special
		}
	}

	for _, role := range getRolesSorted(user.Roles) { // go in order from highest priority to least (aka numerically lowest to highest)
		role.applyDefaults(&info)
	}

	return &info
}

var specialCases = map[uuid.UUID]UserInfo{ // TODO this should basically just be a SELECT * FROM customizations;
	// catgorl
	uuid.MustParse("2c3174fc-0c6b-4cfb-bb2b-0069bf7294d1"): {
		TextColor: "LIGHT_PURPLE",
	},
	// leijurv
	uuid.MustParse("51dcd870-d33b-40e9-9fc1-aecdcff96081"): {
		TextColor: "RED",
		Icon:      "https://files.impactclient.net/img/texture/speckles128.png",
	},
	// triibu popstonia
	uuid.MustParse("8e563236-c7f5-4c82-aa27-c95bf3f4c322"): {
		Icon: "https://files.impactclient.net/img/texture/popstonia.png",
	},
	// popstonia (rebane)
	uuid.MustParse("342fc44b-1fd1-4272-a4c3-a98a2df98abc"): {
		Icon: "https://files.impactclient.net/img/texture/popstonia.png",
	},
}
