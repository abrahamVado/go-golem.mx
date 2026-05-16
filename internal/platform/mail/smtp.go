package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/abrahamVado/go-paladin.mx/internal/platform/config"
)

type Sender interface {
	Send(to []string, subject string, htmlBody string, textBody string) error
	Enabled() bool
}

type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	from     string
	enabled  bool
}

func NewSMTPSender(cfg config.Config) *SMTPSender {
	return &SMTPSender{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		username: cfg.SMTPUsername,
		password: cfg.SMTPPassword,
		from:     cfg.MailFrom,
		enabled:  cfg.MailEnabled,
	}
}

func (s *SMTPSender) Enabled() bool {
	return s != nil && s.enabled
}

func (s *SMTPSender) Send(to []string, subject string, htmlBody string, textBody string) error {
	if !s.Enabled() {
		return nil
	}
	if len(to) == 0 {
		return fmt.Errorf("missing recipient")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	if htmlBody != "" {
		msg.WriteString(htmlBody)
	} else {
		msg.WriteString("<pre>" + textBody + "</pre>")
	}

	client, err := s.dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if s.port != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{
				ServerName: s.host,
				MinVersion: tls.VersionTLS12,
			}); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("smtp server does not support STARTTLS")
		}
	}

	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(s.from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(strings.TrimSpace(recipient)); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg.Bytes()); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func (s *SMTPSender) dial(addr string) (*smtp.Client, error) {
	if s.port == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{
			ServerName: s.host,
			MinVersion: tls.VersionTLS12,
		})
		if err != nil {
			return nil, err
		}
		return smtp.NewClient(conn, s.host)
	}

	return smtp.Dial(addr)
}
