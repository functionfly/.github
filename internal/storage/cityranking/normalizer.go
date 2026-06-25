package cityranking

import (
	"regexp"
	"strings"
)

// MinPopulationForActiveMetric ensures we don't pollute the leaderboard with
// ghost cities (e.g. one user self-reports into a brand new metro).
const MinPopulationForActiveMetric = 1

// MinActiveUsersForPublic is the privacy threshold below which a metro is
// suppressed from the public leaderboard. The plan calls for k≥5 (k-anonymity).
const MinActiveUsersForPublic = 5

// MatchSource indicates how a normalized location was resolved.
type MatchSource string

const (
	MatchAlias        MatchSource = "alias"
	MatchCityAndState MatchSource = "city_state"
	MatchCityOnly     MatchSource = "city_only"
	MatchFallback     MatchSource = "fallback"
)

// NormalizedLocation is the result of attempting to resolve a user-typed
// "Location" string to a known city.
type NormalizedLocation struct {
	CityID    int64
	CitySlug  string
	Source    MatchSource
	Ambiguous bool
}

// stateAbbreviations maps USPS-style abbreviations to full names used in the
// seed data. Used only as a last-resort hint when the user types a 2-letter
// state code.
var stateAbbreviations = map[string]string{
	"al": "Alabama", "ak": "Alaska", "az": "Arizona", "ar": "Arkansas",
	"ca": "California", "co": "Colorado", "ct": "Connecticut", "de": "Delaware",
	"fl": "Florida", "ga": "Georgia", "hi": "Hawaii", "id": "Idaho",
	"il": "Illinois", "in": "Indiana", "ia": "Iowa", "ks": "Kansas",
	"ky": "Kentucky", "la": "Louisiana", "me": "Maine", "md": "Maryland",
	"ma": "Massachusetts", "mi": "Michigan", "mn": "Minnesota", "ms": "Mississippi",
	"mo": "Missouri", "mt": "Montana", "ne": "Nebraska", "nv": "Nevada",
	"nh": "New Hampshire", "nj": "New Jersey", "nm": "New Mexico", "ny": "New York",
	"nc": "North Carolina", "nd": "North Dakota", "oh": "Ohio", "ok": "Oklahoma",
	"or": "Oregon", "pa": "Pennsylvania", "ri": "Rhode Island", "sc": "South Carolina",
	"sd": "South Dakota", "tn": "Tennessee", "tx": "Texas", "ut": "Utah",
	"vt": "Vermont", "va": "Virginia", "wa": "Washington", "wv": "West Virginia",
	"wi": "Wisconsin", "wy": "Wyoming", "dc": "District of Columbia",
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9\s]+`)
var multiSpace = regexp.MustCompile(`\s+`)

// NormalizeInput returns a canonicalized lookup key for a user-typed location.
// Examples:
//   "Austin, TX"          -> "austin tx"
//   "  austin  TX  "      -> "austin tx"
//   "São Paulo"           -> "sao paulo"
//   "Zürich"              -> "zurich"
func NormalizeInput(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	if s == "" {
		return ""
	}
	// Strip stray diacritics first so the next regex pass doesn't drop the
	// underlying letter (e.g. "ã" would otherwise be replaced with a space).
	s = deaccent(s)
	// Strip common punctuation but keep spaces and alphanumerics.
	s = nonAlphaNum.ReplaceAllString(s, " ")
	s = multiSpace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	return s
}

func deaccent(s string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a", "å", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o", "ø", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ñ", "n", "ç", "c",
	)
	return replacer.Replace(s)
}

// ExpandStateAbbreviations splits "austin tx" into city="austin" and state="tx"
// (or full state name). Returns empty strings when no state hint is present.
func ExpandStateAbbreviations(input string) (city, state string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", ""
	}
	last := parts[len(parts)-1]
	if full, ok := stateAbbreviations[last]; ok {
		return strings.Join(parts[:len(parts)-1], " "), full
	}
	return input, ""
}
