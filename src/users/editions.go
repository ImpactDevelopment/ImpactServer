package users

type Edition struct {
	// The edition text, will be followed by " Edition"
	Text string `json:"text,omitempty"`
	// Colour of the edition text
	TextColor string `json:"text_color,omitempty"`
}
