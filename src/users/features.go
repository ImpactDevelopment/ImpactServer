package users

type Features struct {
	// Public features are listed on the public /users/info endpoint unless the user is incognito
	Public []string `json:"public,omitempty"`

	// Private features are only exposed to the user
	Private []string `json:"private,omitempty"`
}

func (user *User) Features() *Features {
	private := user.privateFeatures()
	public := user.publicFeatures()

	if len(private) > 0 || len(public) > 0 {
		return &Features{
			Public:  public,
			Private: private,
		}
	}
	return nil
}

func (user *User) privateFeatures() (features []string) {
	edition := user.Edition()
	if edition != nil {
		features = append(features, "edition")
	}
	return
}

func (user *User) publicFeatures() (features []string) {
	info := user.UserInfo
	if info != nil {
		if info.BackgroundColor != "" || info.BorderColor != "" || info.TextColor != "" {
			features = append(features, "nametag")
		}
		if info.Cape != "" {
			features = append(features, "cape")
		}
		if info.Icon != "" {
			features = append(features, "icon")
		}
	}
	return
}
