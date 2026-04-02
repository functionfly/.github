package captcha

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPickAuthCaptchaProvider(t *testing.T) {
	cs := NewCaptchaService(nil)
	cs.RegisterProvider(NewTurnstileProvider("ts-site", "ts-secret", nil))
	cs.RegisterProvider(NewRecaptchaV3Provider("v3-site", "v3-secret", nil))

	assert.Equal(t, "turnstile", PickAuthCaptchaProvider(cs, ""))
	assert.Equal(t, "recaptcha_v3", PickAuthCaptchaProvider(cs, "recaptcha_v3"))

	empty := NewCaptchaService(nil)
	assert.Equal(t, "", PickAuthCaptchaProvider(empty, ""))
}
