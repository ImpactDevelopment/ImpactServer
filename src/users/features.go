package users

type Features struct {
	Public []string `json:"public,omitempty"`
	All    []string `json:"all,omitempty"`
}

func (user User) Features() *Features {
	var cape, nametag, icon, edition, ignite, nightlies bool
	ignite = user.HasRoleWithID("premium", "staff", "developer") // TODO double check which roles can ignite
	nightlies = user.HasRoleWithID("premium", "pepsi", "staff", "developer")
	for _, template := range user.templates() {
		if cape && nametag && icon && edition {
			break // No need to keep looping if we have matched every feature
		}
		if info := template.info; info != nil {
			if !cape && info.Cape != "" {
				cape = true
			}
			if !nametag && (info.TextColor != "" || info.BackgroundColor != "" || info.BorderColor != "" || info.Icon != "") {
				nametag = true
			}
			if !icon && info.Icon != "" {
				icon = true
			}
		}
		if !edition && template.edition != nil {
			edition = true
		}
	}

	// bool -> string (yes, this is horrible)
	var public []string
	if cape {
		public = append(public, "cape")
	}
	if nametag {
		public = append(public, "nametag")
	}
	if icon {
		public = append(public, "icon")
	}

	all := public
	if ignite {
		all = append(all, "ignite")
	}
	if edition {
		all = append(all, "edition")
	}
	if nightlies {
		all = append(all, "nightlies")
	}

	if len(all) < 1 {
		return nil
	}

	return &Features{
		Public: public,
		All:    all,
	}
}
