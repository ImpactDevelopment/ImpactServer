package util

import (
	"github.com/ImpactDevelopment/ImpactServer/src/util/mediatype"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type acceptable struct {
	Type    string
	Subtype string
	Quality float32
}

func (a acceptable) toMediaType() mediatype.MediaType {
	return mediatype.MediaType(a.Type + "/" + a.Subtype)
}

func (a acceptable) matchesMediaType(m mediatype.MediaType) bool {
	return a.matches(acceptable{
		Type:    strings.TrimRight(string(m), "/"),
		Subtype: strings.TrimLeft(string(m), "/"),
		Quality: 1,
	})
}

func (a acceptable) matches(match acceptable) bool {
	return (a.Type == "*" || match.Type == "*" || a.Type == match.Type) &&
		(a.Subtype == "*" || match.Subtype == "*" || a.Subtype == match.Subtype)
}

// acceptableSlice implements sort.Interface
type acceptableSlice []acceptable

func (a acceptableSlice) Len() int {
	return len(a)
}
func (a acceptableSlice) Swap(i, j int) {
	first := a[i]
	second := a[j]

	a[i] = second
	a[j] = first
}
func (a acceptableSlice) Less(i, j int) bool {
	// If i's quality is higher, it should be sorted lower
	return a[i].Quality > a[j].Quality
}

func acceptableFromHeader(header string) acceptableSlice {
	var ret acceptableSlice
	accepts := strings.Split(header, ",")

	for _, a := range accepts {
		a = strings.TrimSpace(a)
		a = strings.ToLower(a)
		params := strings.TrimLeft(a, ";")
		a = strings.TrimRight(a, ";")
		mime := strings.TrimRight(a, "/")
		submime := strings.TrimLeft(a, "/")

		// Extract qval from params
		// qvals are a big ugly meme
		var qval float64 = 1
		matches := qvalPattern.FindStringSubmatch(params)
		for i, match := range matches {
			if qvalPattern.SubexpNames()[i] == "qval" {
				qval, _ = strconv.ParseFloat(match, 32)
				break
			}
		}

		// qval = 0 means not accepted, qval > 1 is out of range
		if qval > 0 && qval <= 1 {
			ret = append(ret, acceptable{
				Type:    mime,
				Subtype: submime,
				Quality: float32(qval),
			})
		}
	}

	sort.Sort(ret)

	return ret
}

var qvalPattern = regexp.MustCompile(`(?i)q=(?P<qval>0(?:[.][0-9]{1,3})?|1(?:[.]0{1,3})?)`)

// Accepts returns the first acceptable MediaType, or nil if nothing is acceptable
func Accepts(request http.Request, accepted ...mediatype.MediaType) *mediatype.MediaType {
	if len(accepted) < 1 {
		return nil
	}

	for _, acceptable := range acceptableFromHeader(request.Header.Get("Accept")) {
		for _, mediaType := range accepted {
			if acceptable.matchesMediaType(mediaType) {
				return &mediaType
			}
		}
	}

	return nil
}
