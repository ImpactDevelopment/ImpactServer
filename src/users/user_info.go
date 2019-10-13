package users

type UserInfo struct {
	// Icon to display next to this user
	Icon string `json:"icon,omitempty"`
	// Cape this user should wear
	Cape string `json:"cape,omitempty"`
}

func (info *UserInfo) SetCape(cape string) {
	info.Cape = cape
}

func (info *UserInfo) SetIcon(icon string) {
	info.Icon = icon
}
