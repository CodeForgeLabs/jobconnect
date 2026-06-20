package chapa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	SecretKey string
	BaseURL   string
}

type chapaInitResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		CheckoutURL string `json:"checkout_url"`
	} `json:"data"`
}

type PaymentRequest struct {
	Amount      float64
	TxRef       string
	Description string
	Email       string // Optional if Phone is provided
	PhoneNumber string // Optional if Email is provided
	CallbackURL string
	ReturnURL   string
}

func (c *Client) InitializePayment(ctx context.Context, req PaymentRequest) (string, error) {

	if req.Email == "" && req.PhoneNumber == "" {
		return "", fmt.Errorf("either email or phone number is required")
	}

	if req.PhoneNumber != "" {
		if !(strings.HasPrefix(req.PhoneNumber, "09") || strings.HasPrefix(req.PhoneNumber, "07")) {
			return "", fmt.Errorf("phone must start with 09 or 07")
		}
	}

	payload := map[string]any{
		"amount":       fmt.Sprintf("%.2f", req.Amount),
		"currency":     "ETB",
		"tx_ref":       req.TxRef,
		"callback_url": req.CallbackURL,
		"return_url":   req.ReturnURL,
	}

	if req.Email != "" {
		payload["email"] = req.Email
	}
	if req.PhoneNumber != "" {
		payload["phone_number"] = req.PhoneNumber
	}

	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx,
		"POST",
		c.BaseURL+"/v1/transaction/initialize",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.SecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("chapa http error: %d", resp.StatusCode)
	}

	var result chapaInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Status != "success" {
		return "", fmt.Errorf("chapa failed: %s", result.Message)
	}

	return result.Data.CheckoutURL, nil
}
