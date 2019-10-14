package users

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
