package events

import (
	"context"
	"encoding/json"
	"log"

	"jobconnect/proposal/internal/application"
	shared "jobconnect/events"

	"github.com/google/uuid"
)

type proposalCommandPayload struct {
	ProposalID int64  `json:"proposal_id"`
	ClientID   string `json:"client_id"`
	Reason     string `json:"reason"`
	Stage      string `json:"stage"`
}

func StartCommandConsumer(ctx context.Context, brokers []string, contractTopic, jobTopic string, setStatusUC *application.SetProposalStatus, markOfferUC *application.InternalMarkProposalOfferSent, hireUC *application.InternalHireProposal, releaseUC *application.ReleaseHiredProposal) []*shared.Consumer {
	contractConsumer := shared.NewConsumer(brokers, contractTopic, "proposal-service-contract")
	contractConsumer.On("contract.proposal.mark_offer_sent.requested", func(ctx context.Context, env shared.Envelope) error {
		return handleContractEvent(ctx, env, markOfferUC, hireUC, releaseUC)
	})
	contractConsumer.On("contract.proposal.set_hired.requested", func(ctx context.Context, env shared.Envelope) error {
		return handleContractEvent(ctx, env, markOfferUC, hireUC, releaseUC)
	})
	contractConsumer.On("contract.proposal.release_offer.requested", func(ctx context.Context, env shared.Envelope) error {
		return handleContractEvent(ctx, env, markOfferUC, hireUC, releaseUC)
	})
	jobConsumer := shared.NewConsumer(brokers, jobTopic, "proposal-service-job")
	jobConsumer.On("job.proposal.set_status.requested", func(ctx context.Context, env shared.Envelope) error {
		_ = setStatusUC
		log.Printf("job.proposal.set_status.requested ignored until client_id is included in payload")
		return nil
	})
	jobConsumer.On("job.proposal.release_hired.requested", func(ctx context.Context, env shared.Envelope) error {
		return handleContractEvent(ctx, env, markOfferUC, hireUC, releaseUC)
	})
	go func() { _ = contractConsumer.Run(ctx) }()
	go func() { _ = jobConsumer.Run(ctx) }()
	return []*shared.Consumer{contractConsumer, jobConsumer}
}

func handleContractEvent(ctx context.Context, env shared.Envelope, markOfferUC *application.InternalMarkProposalOfferSent, hireUC *application.InternalHireProposal, releaseUC *application.ReleaseHiredProposal) error {
	var p proposalCommandPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return nil
	}
	clientID, err := uuid.Parse(p.ClientID)
	if err != nil {
		return nil
	}
	switch env.EventType {
	case "contract.proposal.mark_offer_sent.requested":
		_, err = markOfferUC.Execute(ctx, application.InternalMarkProposalOfferSentInput{ProposalID: p.ProposalID, ClientID: clientID, Reason: p.Reason})
	case "contract.proposal.set_hired.requested":
		_, err = hireUC.Execute(ctx, application.InternalHireProposalInput{ProposalID: p.ProposalID, ClientID: clientID, RequestID: env.IdempotencyKey, Reason: p.Reason})
	case "contract.proposal.release_offer.requested", "job.proposal.release_hired.requested":
		_, err = releaseUC.Execute(ctx, application.ReleaseHiredProposalInput{ProposalID: p.ProposalID, ClientID: clientID, Reason: p.Reason})
	}
	if err != nil {
		log.Printf("%s failed: %v", env.EventType, err)
	}
	return nil
}
