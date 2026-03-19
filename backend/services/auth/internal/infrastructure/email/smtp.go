package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strconv"
	"strings"
)

// SMTPSender sends OTP emails using an SMTP server.
type SMTPSender struct {
	host         string
	port         int
	tlsMode      string
	username     string
	password     string
	fromAddress  string
	fromName     string
	sendMail     func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
	sendImplicit func(addr, host, username, password, from string, to []string, msg []byte) error
}

func NewSMTPSender(host string, port int, tlsMode, username, password, fromAddress, fromName string) *SMTPSender {
	return &SMTPSender{
		host:         strings.TrimSpace(host),
		port:         port,
		tlsMode:      strings.ToLower(strings.TrimSpace(tlsMode)),
		username:     strings.TrimSpace(username),
		password:     password,
		fromAddress:  strings.TrimSpace(fromAddress),
		fromName:     strings.TrimSpace(fromName),
		sendMail:     smtp.SendMail,
		sendImplicit: sendMailImplicitTLS,
	}
}

func (s *SMTPSender) SendVerifyEmailOTP(ctx context.Context, toEmail, otp string) error {
	_ = ctx
	body := fmt.Sprintf("Your JobConnect verification code is: %s\n\nIf you did not request this, you can ignore this email.", otp)
	return s.send(toEmail, "Verify your JobConnect email", body)
}

func (s *SMTPSender) SendPasswordResetOTP(ctx context.Context, toEmail, otp string) error {
	_ = ctx
	body := fmt.Sprintf("Your JobConnect password reset code is: %s\n\nIf you did not request this, you can ignore this email.", otp)
	return s.send(toEmail, "Reset your JobConnect password", body)
}

func (s *SMTPSender) SendEmailChangeOTP(ctx context.Context, toEmail, otp string) error {
	_ = ctx
	body := fmt.Sprintf("Your JobConnect email change code is: %s\n\nIf you did not request this, you can ignore this email.", otp)
	return s.send(toEmail, "Confirm your new JobConnect email", body)
}

func (s *SMTPSender) send(toEmail, subject, body string) error {
	toEmail = strings.TrimSpace(toEmail)
	if toEmail == "" {
		return fmt.Errorf("recipient email is required")
	}
	if s.host == "" {
		return fmt.Errorf("smtp host is required")
	}
	if s.port <= 0 {
		return fmt.Errorf("smtp port must be positive")
	}
	if s.fromAddress == "" {
		return fmt.Errorf("smtp from address is required")
	}

	addr := s.host + ":" + strconv.Itoa(s.port)
	fromHeader := s.fromAddress
	if s.fromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", s.fromName, s.fromAddress)
	}

	msg := []byte(strings.Join([]string{
		"From: " + fromHeader,
		"To: " + toEmail,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n"))

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}
	switch s.tlsMode {
	case "", "starttls", "none":
		if err := s.sendMail(addr, auth, s.fromAddress, []string{toEmail}, msg); err != nil {
			return fmt.Errorf("smtp send failed: %w", err)
		}
	case "implicit":
		if err := s.sendImplicit(addr, s.host, s.username, s.password, s.fromAddress, []string{toEmail}, msg); err != nil {
			return fmt.Errorf("smtp send failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported smtp tls mode: %s", s.tlsMode)
	}
	return nil
}

func sendMailImplicitTLS(addr, host, username, password, from string, to []string, msg []byte) error {
	tlsConn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return err
	}
	defer tlsConn.Close()

	client, err := smtp.NewClient(tlsConn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if username != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}

	wc, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := io.Copy(wc, strings.NewReader(string(msg))); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	if err := client.Quit(); err != nil && !isNetClosed(err) {
		return fmt.Errorf("smtp send failed: %w", err)
	}
	return nil
}

func isNetClosed(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "use of closed network connection") || err == net.ErrClosed
}
