package email

import (
	"fmt"
	"net/smtp"
)

type EmailSender interface {
	SendEmail(to string, subject string, body string) error
}

type smtpEmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPSender(host, port, username, password, from string) *smtpEmailSender {
	return &smtpEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *smtpEmailSender) SendEmail(to string, subject string, body string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		s.from, to, subject, body,
	))

	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	err := smtp.SendMail(addr, auth, s.from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email via smtp: %w", err)
	}

	return nil
}
