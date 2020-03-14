package users

import "strings"

type Edition struct {
	// The edition icon, will be drawn left of the text
	Icon string `json:"icon,omitempty"`
	// The edition text, will be followed by " Edition"
	Text string `json:"text,omitempty"`
	// Colour of the edition text
	TextColor string `json:"text_color,omitempty"`
}

func (user User) Edition() *Edition {
	// Start by building a list of editions
	var editions []Edition
	if user.MinecraftID != nil {
		if special, ok := specialCases[*user.MinecraftID]; ok {
			if e := special.edition; e != nil {
				editions = append(editions, *e)
			}
		}
	}
	for _, role := range getRolesSorted(user.Roles) {
		if template, ok := defaultRoleTemplates[role.ID]; ok {
			e := template.edition
			if e != nil {
				editions = append(editions, *e)
			}
		}
	}

	// Now we've built the slice, we can reduce it
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
