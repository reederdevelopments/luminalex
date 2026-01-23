package email

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

//go:embed email_template.html
var templateFS embed.FS

// GmailSender implements the Sender interface using the Gmail SMTP server.
type GmailSender struct {
	l         *log.Logger
	fromEmail string
	password  string
	smtpHost  string
	smtpPort  string
}

// NewGmailSender creates a new GmailSender using SMTP with username and password.
// The password should be an "App Password" for Gmail.
func NewGmailSender(l *log.Logger, fromEmail, password string) (Sender, error) {
	if fromEmail == "" || password == "" {
		return nil, fmt.Errorf("email and password must be provided for GmailSender")
	}

	return &GmailSender{
		l:         l,
		fromEmail: fromEmail,
		password:  password,
		smtpHost:  "smtp.gmail.com",
		smtpPort:  "587",
	}, nil
}

// Send sends an email using the Gmail SMTP server.
func (s *GmailSender) Send(to, name, subject, body, surveyName string) error {
	templateBytes, err := templateFS.ReadFile("email_template.html")
	if err != nil {
		return fmt.Errorf("failed to read email template: %w", err)
	}

	htmlBody := string(templateBytes)
	htmlBody = strings.Replace(htmlBody, "{SURVEY_NAME}", surveyName, 1)
	htmlBody = strings.Replace(htmlBody, "{NAME}", name, 1)
	htmlBody = strings.Replace(htmlBody, "{EMAIL_BODY}", body, 1)
	htmlBody = strings.Replace(htmlBody, "{DATE}", time.Now().Format("2006/01/02"), 1)

	from := mail.Address{Name: "Maoni by Unifi", Address: s.fromEmail}
	toAddr := mail.Address{Address: to}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from.String()))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", toAddr.String()))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)

	auth := smtp.PlainAuth("", s.fromEmail, s.password, s.smtpHost)

	err = smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, s.fromEmail, []string{to}, buf.Bytes())
	if err != nil {
		s.l.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	s.l.Printf("Email sent successfully to %s", to)
	return nil
}
