package community

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

var slugInvalidRe = regexp.MustCompile(`[^a-z0-9\s-]`)
var slugMultiDashRe = regexp.MustCompile(`-{2,}`)
var slugMultiSpaceRe = regexp.MustCompile(`\s+`)

// GenerateSlug converts a title into a URL-friendly slug.
// If suffix is true, appends a short random string for collision avoidance.
func GenerateSlug(title string, suffix bool) string {
	slug := strings.ToLower(title)
	slug = slugInvalidRe.ReplaceAllString(slug, "")
	slug = slugMultiSpaceRe.ReplaceAllString(slug, "-")
	slug = slugMultiDashRe.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 200 {
		slug = slug[:200]
		slug = strings.TrimRight(slug, "-")
	}
	if slug == "" {
		slug = "post"
	}
	if suffix {
		slug = fmt.Sprintf("%s-%04x", slug, rand.Intn(0xffff))
	}
	return slug
}
