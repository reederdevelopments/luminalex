package email

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"log"
	"net/mail"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

//go:embed email_template.html
var templateFS embed.FS

// GmailSender implements the Sender interface using the Gmail API.
type GmailSender struct {
	l         *log.Logger
	service   *gmail.Service
	fromEmail string
}

// NewGmailSender creates a new GmailSender.
func NewGmailSender(l *log.Logger, serviceAccountJSON []byte, impersonateUser string) (Sender, error) {
	config, err := google.JWTConfigFromJSON(serviceAccountJSON, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse service account key file: %w", err)
	}
	config.Subject = impersonateUser

	ctx := context.Background()
	ts := config.TokenSource(ctx)

	srv, err := gmail.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %w", err)
	}

	return &GmailSender{
		l:         l,
		service:   srv,
		fromEmail: impersonateUser,
	}, nil
}

// Send sends an email using the Gmail API.
func (s *GmailSender) Send(to, name, subject, body, surveyName string) error {
	templateBytes, err := templateFS.ReadFile("email_template.html")
	if err != nil {
		return fmt.Errorf("failed to read email template: %w", err)
	}

	htmlBody := string(templateBytes)
	htmlBody = strings.Replace(htmlBody, "{SURVEY_NAME}", surveyName, 1)
	htmlBody = strings.Replace(htmlBody, "{NAME}", name, 1)
	htmlBody = strings.Replace(htmlBody, "{EMAIL_BODY}", body, 1)

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

	msg := gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(buf.Bytes()),
	}

	_, err = s.service.Users.Messages.Send("me", &msg).Do()
	if err != nil {
		s.l.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	s.l.Printf("Email sent successfully to %s", to)
	return nil
}
