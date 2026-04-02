package captcha

import (
	"os"
)

// PickAuthCaptchaProvider chooses which registered provider verifies login/signup tokens.
// If AUTH_CAPTCHA_PROVIDER is set and registered, it wins; otherwise preference is
// Turnstile, then reCAPTCHA v3, v2, then hCaptcha.
func PickAuthCaptchaProvider(cs *CaptchaService, explicit string) string {
	if cs == nil {
		return ""
	}
	avail := make(map[string]struct{})
	for _, p := range cs.GetAvailableProviders() {
		avail[p] = struct{}{}
	}
	if explicit == "" {
		explicit = os.Getenv("AUTH_CAPTCHA_PROVIDER")
	}
	if explicit != "" {
		if _, ok := avail[explicit]; ok {
			return explicit
		}
	}
	priority := []string{"turnstile", "recaptcha_v3", "recaptcha_v2", "hcaptcha"}
	for _, p := range priority {
		if _, ok := avail[p]; ok {
			return p
		}
	}
	return ""
}

// PublicSiteKey returns the site key for the chosen auth captcha provider (for browser widgets).
func PublicSiteKey(cs *CaptchaService, explicitProvider string) (providerName, siteKey string, ok bool) {
	name := PickAuthCaptchaProvider(cs, explicitProvider)
	if name == "" {
		return "", "", false
	}
	ch, err := cs.GenerateChallenge(name)
	if err != nil || ch == nil || ch.SiteKey == "" {
		return "", "", false
	}
	return name, ch.SiteKey, true
}
