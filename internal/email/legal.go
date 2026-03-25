package email

import (
	"fmt"
	"time"
)

// TransactionalEmailCopyrightHTML returns the standard HTML footer line for transactional emails.
func TransactionalEmailCopyrightHTML() string {
	return fmt.Sprintf(`<p>&copy; %d FunctionFly LLC. All rights reserved.</p>`, time.Now().Year())
}

// TransactionalEmailCopyrightPlain returns the plain-text copyright line for email footers.
func TransactionalEmailCopyrightPlain() string {
	return fmt.Sprintf("© %d FunctionFly LLC. All rights reserved.", time.Now().Year())
}
