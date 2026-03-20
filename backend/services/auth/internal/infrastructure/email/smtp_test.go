package email

import (
	"context"
	"errors"
	"net/smtp"
	"strings"
	"testing"
)

func TestSMTPSender_SendVerifyEmailOTP_Success(t *testing.T) {
	sender := NewSMTPSender("smtp.example.com", 587, "starttls", "user", "pass", "no-reply@example.com", "JobConnect")

	called := false
	sender.sendMail = func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
		called = true
		if addr != "smtp.example.com:587" {
			t.Fatalf("unexpected addr: %s", addr)
		}
		if from != "no-reply@example.com" {
			t.Fatalf("unexpected from: %s", from)
		}
		if len(to) != 1 || to[0] != "user@example.com" {
			t.Fatalf("unexpected to: %v", to)
		}
		message := string(msg)
		if !strings.Contains(message, "Subject: Verify your JobConnect email") {
			t.Fatalf("missing subject in message")
		}
		if !strings.Contains(message, "Your JobConnect verification code is: 123456") {
			t.Fatalf("missing otp in message")
		}
		return nil
	}

	err := sender.SendVerifyEmailOTP(context.Background(), "user@example.com", "123456")
	if err != nil {
		t.Fatalf("SendVerifyEmailOTP error: %v", err)
	}
	if !called {
		t.Fatalf("expected sendMail to be called")
	}
}

func TestSMTPSender_SendPasswordResetOTP_SendError(t *testing.T) {
	sender := NewSMTPSender("smtp.example.com", 587, "starttls", "", "", "no-reply@example.com", "")
	sender.sendMail = func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("dial failed")
	}

	err := sender.SendPasswordResetOTP(context.Background(), "user@example.com", "654321")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "smtp send failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSMTPSender_SendEmailChangeOTP_ValidatesRecipient(t *testing.T) {
	sender := NewSMTPSender("smtp.example.com", 587, "starttls", "", "", "no-reply@example.com", "")
	err := sender.SendEmailChangeOTP(context.Background(), "", "999999")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "recipient email is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSMTPSender_ImplicitTLSBranch(t *testing.T) {
	sender := NewSMTPSender("smtp.example.com", 465, "implicit", "user", "pass", "no-reply@example.com", "JobConnect")

	called := false
	sender.sendImplicit = func(addr, host, username, password, from string, to []string, msg []byte) error {
		called = true
		if addr != "smtp.example.com:465" {
			t.Fatalf("unexpected addr: %s", addr)
		}
		if host != "smtp.example.com" {
			t.Fatalf("unexpected host: %s", host)
		}
		if from != "no-reply@example.com" {
			t.Fatalf("unexpected from: %s", from)
		}
		if len(to) != 1 || to[0] != "user@example.com" {
			t.Fatalf("unexpected to: %v", to)
		}
		return nil
	}

	err := sender.SendVerifyEmailOTP(context.Background(), "user@example.com", "111111")
	if err != nil {
		t.Fatalf("SendVerifyEmailOTP error: %v", err)
	}
	if !called {
		t.Fatalf("expected implicit sender branch to be called")
	}
}
