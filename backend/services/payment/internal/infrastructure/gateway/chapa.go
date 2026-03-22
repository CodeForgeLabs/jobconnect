package gateway

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"jobconnect/payment/internal/application"
)

const chapaBaseURL = "https://api.chapa.co/v1"

type ChapaGateway struct {
	secretKey  string
	httpClient *http.Client
}

func NewChapaGateway(secretKey string) *ChapaGateway {
	return &ChapaGateway{
		secretKey:  secretKey,
		httpClient: &http.Client{},
	}
}

// Ensure ChapaGateway implements PaymentGateway
var _ application.PaymentGateway = (*ChapaGateway)(nil)

func (g *ChapaGateway) InitiateCheckout(ctx context.Context, in application.CheckoutInput) (application.CheckoutResult, error) {
	url := fmt.Sprintf("%s/transaction/initialize", chapaBaseURL)
	
	// Convert AmountMinor to major currency (e.g. 100 ETB cents = 1 ETB)
	// We assume AmountMinor is in the smallest unit. Chapa expects major unit for amount?
	// Actually, typically if it's ETB, you pass the full amount in ETB. Let's assume minor is cents, so /100.
	amountMajor := float64(in.AmountMinor) / 100.0

	payload := map[string]interface{}{
		"amount":       amountMajor,
		"currency":     in.Currency,
		"email":        in.Email,
		"first_name":   in.FirstName,
		"last_name":    in.LastName,
		"phone_number": in.PhoneNumber,
		"tx_ref":       in.TxRef,
		"callback_url": in.CallbackURL,
		"return_url":   in.ReturnURL,
		"custom_title": "JobConnect Escrow Deposit",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return application.CheckoutResult{}, fmt.Errorf("failed to marshal chapa payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return application.CheckoutResult{}, fmt.Errorf("failed to create chapa request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return application.CheckoutResult{}, fmt.Errorf("chapa request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return application.CheckoutResult{}, fmt.Errorf("chapa returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			CheckoutURL string `json:"checkout_url"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return application.CheckoutResult{}, fmt.Errorf("failed to parse chapa response: %w", err)
	}

	if result.Status != "success" {
		return application.CheckoutResult{}, fmt.Errorf("chapa init failed: %s", result.Message)
	}

	return application.CheckoutResult{
		CheckoutURL: result.Data.CheckoutURL,
		ExternalRef: in.TxRef,
	}, nil
}

func (g *ChapaGateway) VerifyPayment(ctx context.Context, externalRef string) (application.VerifyResult, error) {
	url := fmt.Sprintf("%s/transaction/verify/%s", chapaBaseURL, externalRef)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return application.VerifyResult{}, fmt.Errorf("failed to create chapa verify request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.secretKey)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return application.VerifyResult{}, fmt.Errorf("chapa verify request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Status   string  `json:"status"`
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
			TxRef    string  `json:"tx_ref"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return application.VerifyResult{}, fmt.Errorf("failed to parse chapa verify response: %w (%s)", err, string(respBody))
	}

	isVerified := result.Status == "success" && result.Data.Status == "success"
	amountMinor := int64(result.Data.Amount * 100)

	return application.VerifyResult{
		Verified:    isVerified,
		AmountMinor: amountMinor,
		Currency:    result.Data.Currency,
		ExternalRef: result.Data.TxRef,
	}, nil
}

func (g *ChapaGateway) InitiateTransfer(ctx context.Context, in application.TransferInput) (application.TransferResult, error) {
	url := fmt.Sprintf("%s/transfers", chapaBaseURL)

	amountMajor := float64(in.AmountMinor) / 100.0
	payload := map[string]interface{}{
		"account_name":   in.AccountHolderName,
		"account_number": in.AccountNumber,
		"amount":         amountMajor,
		"currency":       in.Currency,
		"reference":      in.TxRef,
		"bank_code":      in.BankCode,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return application.TransferResult{}, fmt.Errorf("failed to marshal chapa transfer payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return application.TransferResult{}, fmt.Errorf("failed to create chapa transfer request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return application.TransferResult{}, fmt.Errorf("chapa transfer request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return application.TransferResult{}, fmt.Errorf("chapa transfer returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    string `json:"data"` // typically a transfer ID
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return application.TransferResult{}, fmt.Errorf("failed to parse chapa transfer response: %w", err)
	}

	if result.Status != "success" {
		return application.TransferResult{}, fmt.Errorf("chapa transfer failed: %s", result.Message)
	}

	return application.TransferResult{
		TransferID:  result.Data,
		ExternalRef: in.TxRef,
	}, nil
}

func (g *ChapaGateway) VerifyWebhookSignature(payload []byte, signature string) error {
	mac := hmac.New(sha256.New, []byte(g.secretKey))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		return fmt.Errorf("invalid chapa webhook signature")
	}

	return nil
}
