package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"jobconnect/payment/internal/application"
)

type WebhookHandler struct {
	verifyDeposit *application.VerifyDeposit
	chapaGateway  application.PaymentGateway
	telebirrGW    application.PaymentGateway
}

func NewWebhookHandler(
	verifyDeposit *application.VerifyDeposit,
	chapaGateway application.PaymentGateway,
	telebirrGW application.PaymentGateway,
) *WebhookHandler {
	return &WebhookHandler{
		verifyDeposit: verifyDeposit,
		chapaGateway:  chapaGateway,
		telebirrGW:    telebirrGW,
	}
}

func (h *WebhookHandler) HandleChapaWebhook(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get("Chapa-Signature")
	if signature == "" {
		http.Error(w, "missing signature", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	err = h.chapaGateway.VerifyWebhookSignature(body, signature)
	if err != nil {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Event string `json:"event"`
		TxRef string `json:"tx_ref"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if payload.Event == "charge.success" {
		log.Printf("Received Chapa success webhook for tx_ref: %s", payload.TxRef)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) HandleTelebirrWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Telebirr webhook")
	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/webhooks/chapa", h.HandleChapaWebhook)
	mux.HandleFunc("/webhooks/telebirr", h.HandleTelebirrWebhook)
}
