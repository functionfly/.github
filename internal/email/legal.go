package email

import (
	"fmt"
	"time"
)

// TransactionalEmailCopyrightHTML returns the standard HTML footer line for transactional emails.
// Uses Velocity Orange (#f97316) for brand consistency.
func TransactionalEmailCopyrightHTML() string {
	return fmt.Sprintf(`© %d FunctionFly. All rights reserved.<br>
<a href="https://functionfly.com/privacy" style="color:#f97316;text-decoration:none;">Privacy Policy</a> · 
<a href="https://functionfly.com/terms" style="color:#f97316;text-decoration:none;">Terms of Service</a>`, time.Now().Year())
}

// TransactionalEmailCopyrightPlain returns the plain-text copyright line for email footers.
func TransactionalEmailCopyrightPlain() string {
	return fmt.Sprintf("© %d FunctionFly. All rights reserved.\nhttps://functionfly.com/privacy | https://functionfly.com/terms", time.Now().Year())
}
