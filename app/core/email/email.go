package email

import "log"

// Sender defines an interface for sending emails.
type Sender interface {
	Send(to, name, subject, body, surveyName string) error
}

// LogSender implements the Sender interface by logging emails to the console.
type LogSender struct {
	l *log.Logger
}

// NewLogSender creates a new LogSender.
func NewLogSender(l *log.Logger) *LogSender {
	return &LogSender{l: l}
}

// Send logs the email details instead of sending a real email.
func (s *LogSender) Send(to, name, subject, body, surveyName string) error {
	s.l.Printf("--- EMAIL ---")
	s.l.Printf("To: %s (%s)", to, name)
	s.l.Printf("Subject: %s", subject)
	s.l.Printf("Survey Name: %s", surveyName)
	s.l.Printf("Body: \n%s", body)
	s.l.Printf("-------------")
	return nil
}
