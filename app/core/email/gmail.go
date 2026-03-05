package email

import (
	"bytes"
	"crypto/tls"
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

type GmailSender struct {
	l         *log.Logger
	fromEmail string
	password  string
	smtpHost  string
	smtpPort  string
}

func NewGmailSender(l *log.Logger, fromEmail, password string) (Sender, error) {
	if fromEmail == "" {
		return nil, fmt.Errorf("email must be provided for GmailSender")
	}

	return &GmailSender{
		l:         l,
		fromEmail: fromEmail,
		password:  password,
		smtpHost:  "smtp.gmail.com", // Keeping the standard host for local dev
		smtpPort:  "587",
	}, nil
}

// buildMessageBytes is a helper to generate the raw email payload
func (s *GmailSender) buildMessageBytes(msg Message) ([]byte, error) {
	templateBytes, err := templateFS.ReadFile("email_template.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read email template: %w", err)
	}

	htmlBody := string(templateBytes)
	htmlBody = strings.Replace(htmlBody, "{SURVEY_NAME}", msg.SurveyName, 1)
	htmlBody = strings.Replace(htmlBody, "{NAME}", msg.Name, 1)
	htmlBody = strings.Replace(htmlBody, "{EMAIL_BODY}", msg.Body, 1)
	htmlBody = strings.Replace(htmlBody, "{DATE}", time.Now().Format("2006/01/02"), 1)

	from := mail.Address{Name: "Community Feedback by Unifi", Address: s.fromEmail}
	toAddr := mail.Address{Address: msg.To}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from.String()))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", toAddr.String()))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)

	return buf.Bytes(), nil
}

// Send handles a single email (legacy support)
func (s *GmailSender) Send(to, name, subject, body, surveyName string) error {
	msg := Message{To: to, Name: name, Subject: subject, Body: body, SurveyName: surveyName}
	msgBytes, err := s.buildMessageBytes(msg)
	if err != nil {
		return err
	}

	var auth smtp.Auth
	if s.password != "" {
		auth = smtp.PlainAuth("", s.fromEmail, s.password, s.smtpHost)
	}

	err = smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, s.fromEmail, []string{to}, msgBytes)
	if err != nil {
		s.l.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	s.l.Printf("Email sent successfully to %s", to)
	return nil
}

// SendBatch handles multiple emails over a single connection
func (s *GmailSender) SendBatch(messages []Message) error {
	if len(messages) == 0 {
		return nil
	}

	// 1. Establish the connection manually
	client, err := smtp.Dial(s.smtpHost + ":" + s.smtpPort)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer client.Close()

	// 2. Upgrade to TLS (Required for port 587)
	if err := client.StartTLS(&tls.Config{ServerName: s.smtpHost}); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// 3. Authenticate once
	if s.password != "" {
		auth := smtp.PlainAuth("", s.fromEmail, s.password, s.smtpHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// 4. Stream emails through the open connection
	for _, msg := range messages {
		msgBytes, err := s.buildMessageBytes(msg)
		if err != nil {
			s.l.Printf("Skipping message to %s (build error): %v", msg.To, err)
			continue
		}

		if err := client.Mail(s.fromEmail); err != nil {
			s.l.Printf("Failed MAIL command for %s: %v", msg.To, err)
			continue
		}
		if err := client.Rcpt(msg.To); err != nil {
			s.l.Printf("Failed RCPT command for %s: %v", msg.To, err)
			continue
		}

		w, err := client.Data()
		if err != nil {
			s.l.Printf("Failed DATA command for %s: %v", msg.To, err)
			continue
		}

		if _, err = w.Write(msgBytes); err != nil {
			s.l.Printf("Failed to write message body for %s: %v", msg.To, err)
		}

		if err = w.Close(); err != nil {
			s.l.Printf("Failed to close DATA writer for %s: %v", msg.To, err)
		} else {
			s.l.Printf("Successfully sent batched email to %s", msg.To)
		}

		// Optional: add a tiny 50ms sleep here if Google still complains about throughput rate
	}

	// 5. Cleanly terminate the connection
	return client.Quit()
}
