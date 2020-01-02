package mime

type MimeType string

const (
	JSON MimeType = "application/json"
	Form MimeType = "application/x-www-form-urlencoded"
)

func (t MimeType) String() string {
	return string(t)
}
