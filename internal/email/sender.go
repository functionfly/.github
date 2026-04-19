package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
)

// buildMessage constructs a RFC-compliant multipart email message
func buildMessage(to, subject, fromName, fromEmail string, textBody, htmlBody string) []byte {
	var msg bytes.Buffer

	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, fromEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	msg.WriteString("\r\n")

	// Plain text part
	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n\r\n")

	// HTML part
	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n\r\n")

	msg.WriteString("--boundary123--\r\n")

	return msg.Bytes()
}

// buildMessageMultiple constructs a multipart email message for multiple recipients
func buildMessageMultiple(to []string, subject, fromName, fromEmail string, textBody, htmlBody string) []byte {
	var msg bytes.Buffer

	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, fromEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	msg.WriteString("\r\n")

	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n\r\n")

	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n\r\n")

	msg.WriteString("--boundary123--\r\n")

	return msg.Bytes()
}

// sendSMTP sends an email via SMTP, authenticating if credentials are provided
func sendSMTP(addr, username, password, fromEmail string, to []string, message []byte) error {
	var err error
	if username != "" || password != "" {
		auth := smtp.PlainAuth("", username, password, extractHost(addr))
		err = smtp.SendMail(addr, auth, fromEmail, to, message)
	} else {
		err = smtp.SendMail(addr, nil, fromEmail, to, message)
	}
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// extractHost returns the hostname portion of a host:port address
func extractHost(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
