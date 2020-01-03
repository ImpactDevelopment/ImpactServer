package mediatype

// MediaType is a type as used by Accept, Content-Type and similar HTTP headers
type MediaType string

// Default MediaType constants
const (
	JSON MediaType = "application/json"
	XML  MediaType = "application/xml"
	Form MediaType = "application/x-www-form-urlencoded"
)

func (t MediaType) String() string {
	return string(t)
}
