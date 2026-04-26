package handlers

import (
	"context"
	"net/http"
	"strings"

	disputev1 "jobconnect/dispute/gen/dispute/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type disputeClient interface {
	OpenDispute(ctx context.Context, in *disputev1.OpenDisputeRequest, opts ...grpc.CallOption) (*disputev1.OpenDisputeResponse, error)
	GetDispute(ctx context.Context, in *disputev1.GetDisputeRequest, opts ...grpc.CallOption) (*disputev1.GetDisputeResponse, error)
	ListDisputes(ctx context.Context, in *disputev1.ListDisputesRequest, opts ...grpc.CallOption) (*disputev1.ListDisputesResponse, error)
	ResolveDispute(ctx context.Context, in *disputev1.ResolveDisputeRequest, opts ...grpc.CallOption) (*disputev1.ResolveDisputeResponse, error)
}

type DisputeHandler struct {
	client disputeClient
}

func NewDisputeHandler(client disputeClient) *DisputeHandler {
	return &DisputeHandler{client: client}
}

func (h *DisputeHandler) OpenDispute(c *gin.Context) {
	var body struct {
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
		Reason        string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.OpenDispute(withAuthContext(c), &disputev1.OpenDisputeRequest{
		ReferenceType: body.ReferenceType,
		ReferenceId:   body.ReferenceID,
		Reason:        body.Reason,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "dispute", resp.GetDispute())
}

func (h *DisputeHandler) GetDispute(c *gin.Context) {
	disputeID, ok := parseInt64Param(c, "disputeId")
	if !ok {
		return
	}
	resp, err := h.client.GetDispute(withAuthContext(c), &disputev1.GetDisputeRequest{DisputeId: disputeID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "dispute", resp.GetDispute())
}

func (h *DisputeHandler) ListDisputes(c *gin.Context) {
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	resp, err := h.client.ListDisputes(withAuthContext(c), &disputev1.ListDisputesRequest{
		ReferenceType: strings.TrimSpace(c.Query("reference_type")),
		ReferenceId:   strings.TrimSpace(c.Query("reference_id")),
		Status:        disputev1.DisputeStatus(disputev1.DisputeStatus_value[status]),
		PageSize:      int32(parseIntQuery(c, "page_size", 20)),
		PageToken:     strings.TrimSpace(c.Query("page_token")),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	payload, convErr := protoSliceToAny(resp.GetDisputes())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"disputes": payload, "next_page_token": resp.GetNextPageToken()})
}

func (h *DisputeHandler) ResolveDispute(c *gin.Context) {
	disputeID, ok := parseInt64Param(c, "disputeId")
	if !ok {
		return
	}
	var body struct {
		Decision string `json:"decision"`
		Note     string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.ResolveDispute(withAuthContext(c), &disputev1.ResolveDisputeRequest{
		DisputeId: disputeID,
		Decision:  disputev1.DisputeDecision(disputev1.DisputeDecision_value[strings.ToUpper(strings.TrimSpace(body.Decision))]),
		Note:      body.Note,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	writeProtoEnvelope(c, http.StatusOK, "dispute", resp.GetDispute())
}
