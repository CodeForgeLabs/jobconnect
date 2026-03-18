package email

import "context"

// NoopSender does not send emails (for dev or when email is disabled).
type NoopSender struct{}

func NewNoopSender() *NoopSender {
	return &NoopSender{}
}

func (NoopSender) SendVerifyEmailOTP(ctx context.Context, email, otp string) error {
	return nil
}

func (NoopSender) SendPasswordResetOTP(ctx context.Context, email, otp string) error {
	return nil
}
