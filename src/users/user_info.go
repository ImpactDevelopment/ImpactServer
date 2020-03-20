package users

import (
	"github.com/google/uuid"
)

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
}

// NewUserInfo creates a UserInfo based on a User's roles and any special cases that apply to them
func NewUserInfo(user User) *UserInfo {
	var info UserInfo
	templates := user.templates()

	if len(templates) < 1 {
		return nil
	}

	// Reverse loop so that higher priority templates overwrite lower ones
	for i := len(templates) - 1; i > 0; i-- {
		if template := templates[i].info; template != nil {
			if template.Icon != "" {
				info.Icon = template.Icon
			}
			if template.Cape != "" {
				info.Cape = template.Cape
			}
			if template.TextColor != "" {
				info.TextColor = template.TextColor
			}
			if template.BackgroundColor != "" {
				info.BackgroundColor = template.BackgroundColor
			}
			if template.BorderColor != "" {
				info.BorderColor = template.BorderColor
			}
		}
	}

	return &info
}

var specialCases = map[uuid.UUID]roleTemplate{ // TODO this should basically just be a SELECT * FROM customizations;
	// catgorl
	uuid.MustParse("2c3174fc-0c6b-4cfb-bb2b-0069bf7294d1"): {
		info: &UserInfo{
			TextColor: "LIGHT_PURPLE",
		},
	},
	// leijurv
	uuid.MustParse("51dcd870-d33b-40e9-9fc1-aecdcff96081"): {
		info: &UserInfo{
			TextColor: "RED",
			Icon:      "https://files.impactclient.net/img/texture/speckles128.png",
		},
		edition: &Edition{
			Icon: "https://files.impactclient.net/img/texture/speckles128.png",
		},
	},
	// triibu popstonia
	uuid.MustParse("8e563236-c7f5-4c82-aa27-c95bf3f4c322"): {
		info: &UserInfo{
			Icon: "https://files.impactclient.net/img/texture/popstonia.png",
		},
	},
	// popstonia (rebane)
	uuid.MustParse("342fc44b-1fd1-4272-a4c3-a98a2df98abc"): {
		info: &UserInfo{
			Icon: "https://files.impactclient.net/img/texture/popstonia.png",
		},
	},
}
