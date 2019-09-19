package util

import "strings"

// GetSubdomains tries to return the subdomain part from a host or hostname using ugly hacks.
func GetSubdomains(host string) string {
	domains := []string{"localhost"}
	ogTLD := []string{"com", "org", "net", "int", "edu", "gov", "mil", "biz"}
	ccTLD := []string{"ac", "ad", "ae", "af", "ag", "ai", "al", "am", "ao", "aq", "ar", "as", "at", "au", "aw", "ax", "az", "ba", "bb", "bd", "be", "bf", "bg", "bh", "bi", "bj", "bm", "bn", "bo", "br", "bs", "bt", "bw", "by", "bz", "ca", "cc", "cd", "cf", "cg", "ch", "ci", "ck", "cl", "cm", "cn", "co", "cr", "cu", "cv", "cw", "cx", "cy", "cz", "de", "dj", "dk", "dm", "do", "dz", "ec", "ee", "eg", "er", "es", "et", "eu", "fi", "fj", "fk", "fm", "fo", "fr", "ga", "gd", "ge", "gf", "gg", "gh", "gi", "gl", "gm", "gn", "gp", "gq", "gr", "gs", "gt", "gu", "gw", "gy", "hk", "hm", "hn", "hr", "ht", "hu", "id", "ie", "il", "im", "in", "io", "iq", "ir", "is", "it", "je", "jm", "jo", "jp", "ke", "kg", "kh", "ki", "km", "kn", "kp", "kr", "kw", "ky", "kz", "la", "lb", "lc", "li", "lk", "lr", "ls", "lt", "lu", "lv", "ly", "ma", "mc", "md", "me", "mg", "mh", "mk", "ml", "mm", "mn", "mo", "mp", "mq", "mr", "ms", "mt", "mu", "mv", "mw", "mx", "my", "mz", "na", "nc", "ne", "nf", "ng", "ni", "nl", "no", "np", "nr", "nu", "nz", "om", "pa", "pe", "pf", "pg", "ph", "pk", "pl", "pm", "pn", "pr", "ps", "pt", "pw", "py", "qa", "re", "ro", "rs", "ru", "rw", "sa", "sb", "sc", "sd", "se", "sg", "sh", "si", "sk", "sl", "sm", "sn", "so", "sr", "ss", "st", "su", "sv", "sx", "sy", "sz", "tc", "td", "tf", "tg", "th", "tj", "tk", "tl", "tm", "tn", "to", "tr", "tt", "tv", "tw", "tz", "ua", "ug", "uk", "us", "uy", "uz", "va", "vc", "ve", "vg", "vi", "vn", "vu", "wf", "ws", "ye", "yt", "za", "zm", "zw", "xn--lgbbat1ad8j", "xn--y9a3aq", "xn--mgbcpq6gpa1a", "xn--54b7fta0cc", "xn--90ais", "xn--90ae", "xn--fiqs8s", "xn--fiqz9s", "xn--wgbh1c", "xn--e1a4c", "xn--qxa6a", "xn--node", "xn--qxam", "xn--j6w193g", "xn--h2brj9c", "xn--mgbbh1a71e", "xn--fpcrj9c3d", "xn--gecrj9c", "xn--s9brj9c", "xn--xkc2dl3a5ee0h", "xn--45brj9c", "xn--2scrj9c", "xn--rvc1e0am3e", "xn--45br5cyl", "xn--3hcrj9c", "xn--mgbbh1a", "xn--h2breg3eve", "xn--h2brj9c8c", "xn--mgbgu82a", "xn--mgba3a4f16a", "xn--mgbtx2b", "xn--mgbayh7gpa", "xn--80ao21a", "xn--q7ce6a", "xn--mix082f", "xn--mix891f", "xn--mgbx4cd0ab", "xn--mgbah1a3hjkrd", "xn--l1acc", "xn--mgbc0a9azcg", "xn--d1alf", "xn--mgb9awbf", "xn--mgbai9azgqp6j", "xn--ygbi2ammx", "xn--wgbl6a", "xn--p1ai", "xn--mgberp4a5d4ar", "xn--90a3ac", "xn--yfro4i67o", "xn--clchc0ea0b2g2a9gcd", "xn--3e0b707e", "xn--fzc2c9e2c", "xn--xkc2al3hye2a", "xn--mgbpl2fh", "xn--ogbpf8fl", "xn--kprw13d", "xn--kpry57d", "xn--o3cw4h", "xn--pgbs0dh", "xn--j1amh", "xn--mgbaam7a8h", "xn--mgb2ddes"}
	localTLD := []string{"example", "invalid", "local", "test"}
	sLD := []string{"com", "co", "me", "net", "org", "sch", "edu"}

	host = stripPort(host)

	tld := getTLD(host)
	host = stripTLD(host)

	if contains(domains, tld) {
		return host
	}

	if contains(localTLD, tld) {
		return stripTLD(host)
	}

	if contains(ogTLD, tld) {
		return stripTLD(host)
	}

	if contains(ccTLD, tld) {
		tld2 := getTLD(host)
		host2 := stripTLD(host)
		// check 2nd level domain on ccTLDs
		if contains(sLD, tld2) {
			return stripTLD(host2)
		}
		return host2
	}
	return ""
}

func stripPort(host string) string {
	i := strings.LastIndex(host, ":")
	if i < 0 {
		return host
	}
	return host[:i]
}

func getTLD(domain string) string {
	i := strings.LastIndex(domain, ".")
	if i < 0 {
		return domain
	}
	return domain[i+1:]
}

func stripTLD(domain string) string {
	i := strings.LastIndex(domain, ".")
	if i < 0 {
		return ""
	}
	return domain[:i]
}

func contains(slice []string, entry string) bool {
	for _, value := range slice {
		if value == entry {
			return true
		}
	}
	return false
}
