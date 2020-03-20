package users

import "strings"

type Edition struct {
	// The edition icon, will be drawn left of the text
	Icon string `json:"icon,omitempty"`
	// The edition text, User.Edition() concatenates these into a list and appends "Edition"
	// e.g. "Pepsi Premium Edition"
	Text string `json:"text,omitempty"`
	// Colour of the edition text
	TextColor string `json:"text_color,omitempty"`
}

func (user User) Edition() *Edition {
	// map from generic roleTemplate to edition template
	var editions []Edition
	for _, template := range user.templates() {
		if e := template.edition; e != nil {
			editions = append(editions, *e)
		}
	}

	// reduce to a single edition struct
	if len(editions) > 0 {
		var ret Edition

		// Set first icon
		for _, e := range editions {
			if e.Icon != "" {
				ret.Icon = e.Icon
				break
			}
		}

		// Set first text_color
		for _, e := range editions {
			if e.TextColor != "" {
				ret.TextColor = e.TextColor
				break
			}
		}

		// Concatenate the text
		var text strings.Builder
		for _, e := range editions {
			if text.Len() > 0 && e.Text != "" {
				text.WriteString(" ")
			}
			text.WriteString(e.Text)
		}
		if text.Len() > 0 {
			text.WriteString(" Edition")
		}
		ret.Text = text.String()

		return &ret
	}

	return nil
}
