package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContractRepo struct {
	pool *pgxpool.Pool
}

func NewContractRepo(pool *pgxpool.Pool) *ContractRepo {
	return &ContractRepo{pool: pool}
}

func (r *ContractRepo) Create(ctx context.Context, c domain.Contract) (int64, error) {
	milestonesRaw, err := json.Marshal(c.Milestones)
	if err != nil {
		return 0, err
	}
	var proposalID *int64
	if c.ProposalID > 0 {
		proposalID = &c.ProposalID
	}
	var id int64
	err = r.pool.QueryRow(ctx, `
		insert into contracts (
			client_id, freelancer_id, job_id, proposal_id, contract_type, status,
			title, description, hourly_rate, fixed_total, weekly_hour_limit,
			milestones, created_at, updated_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		returning id
	`,
		c.ClientID,
		c.FreelancerID,
		c.JobID,
		proposalID,
		c.ContractType,
		c.Status,
		c.Title,
		c.Description,
		c.HourlyRate,
		c.FixedTotal,
		c.WeeklyHourLimit,
		milestonesRaw,
		c.CreatedAt,
		c.UpdatedAt,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			switch uniqueViolationConstraint(err) {
			case "uq_contracts_proposal_id_nonzero":
				return 0, fmt.Errorf("contract already exists for proposal")
			case "uq_contracts_pending_offer_per_job_client":
				return 0, fmt.Errorf("job already has a pending offer")
			}
		}
		return 0, err
	}
	return id, nil
}

func (r *ContractRepo) GetByID(ctx context.Context, contractID int64) (domain.Contract, error) {
	row := r.pool.QueryRow(ctx, selectBase+` where id = $1`, contractID)
	c, err := scanContract(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Contract{}, ErrNotFound
		}
		return domain.Contract{}, err
	}
	return c, nil
}

func (r *ContractRepo) GetByIDForActor(ctx context.Context, contractID int64, actorID uuid.UUID) (domain.Contract, error) {
	row := r.pool.QueryRow(ctx, selectBase+` where id = $1 and (client_id = $2 or freelancer_id = $2)`, contractID, actorID)
	c, err := scanContract(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Contract{}, ErrNotFound
		}
		return domain.Contract{}, err
	}
	return c, nil
}

func (r *ContractRepo) GetByProposalID(ctx context.Context, proposalID int64) (domain.Contract, error) {
	row := r.pool.QueryRow(ctx, selectBase+` where proposal_id = $1`, proposalID)
	c, err := scanContract(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Contract{}, ErrNotFound
		}
		return domain.Contract{}, err
	}
	return c, nil
}

func (r *ContractRepo) GetJobOfferState(ctx context.Context, jobID int64, clientID uuid.UUID) (domain.JobOfferState, error) {
	state := domain.JobOfferState{JobID: jobID}
	rows, err := r.pool.Query(ctx, `
		select id, status
		from contracts
		where job_id = $1 and client_id = $2
		order by created_at desc
	`, jobID, clientID)
	if err != nil {
		return domain.JobOfferState{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var contractID int64
		var status string
		if err := rows.Scan(&contractID, &status); err != nil {
			return domain.JobOfferState{}, err
		}
		switch strings.ToLower(strings.TrimSpace(status)) {
		case domain.StatusPendingAcceptance:
			if !state.HasPendingOffer {
				state.HasPendingOffer = true
				state.PendingContractID = contractID
			}
		case domain.StatusActive:
			if !state.HasActiveContract {
				state.HasActiveContract = true
				state.ActiveContractID = contractID
			}
		}
	}
	if err := rows.Err(); err != nil {
		return domain.JobOfferState{}, err
	}
	return state, nil
}

func (r *ContractRepo) ListByActor(ctx context.Context, actorID uuid.UUID, status string, limit, offset int) ([]domain.Contract, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		rows, err := r.pool.Query(ctx, selectBase+` where (client_id = $1 or freelancer_id = $1) order by created_at desc limit $2 offset $3`, actorID, limit, offset)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanContracts(rows)
	}

	rows, err := r.pool.Query(ctx, selectBase+` where (client_id = $1 or freelancer_id = $1) and status = $2 order by created_at desc limit $3 offset $4`, actorID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContracts(rows)
}

func (r *ContractRepo) UpdateOfferForClient(ctx context.Context, c domain.Contract) error {
	milestonesRaw, err := json.Marshal(c.Milestones)
	if err != nil {
		return err
	}
	res, err := r.pool.Exec(ctx, `
		update contracts
		set contract_type = $3,
			status = $4,
			title = $5,
			description = $6,
			hourly_rate = $7,
			fixed_total = $8,
			weekly_hour_limit = $9,
			milestones = $10,
			updated_at = $11,
			activated_at = null,
			declined_at = null,
			revoked_at = null
		where id = $1 and client_id = $2
	`, c.ID, c.ClientID, c.ContractType, c.Status, c.Title, c.Description, c.HourlyRate, c.FixedTotal, c.WeeklyHourLimit, milestonesRaw, c.UpdatedAt)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContractRepo) SetStatusForFreelancer(ctx context.Context, contractID int64, freelancerID uuid.UUID, status string, at time.Time) error {
	status = strings.ToLower(strings.TrimSpace(status))
	var tag int64
	var err error
	switch status {
	case domain.StatusActive:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, activated_at = $4
			where id = $1 and freelancer_id = $2
		`, contractID, freelancerID, status, at)
		err = execErr
		tag = res.RowsAffected()
	case domain.StatusDeclined:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, declined_at = $4
			where id = $1 and freelancer_id = $2
		`, contractID, freelancerID, status, at)
		err = execErr
		tag = res.RowsAffected()
	case domain.StatusRevoked:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, revoked_at = $4
			where id = $1 and freelancer_id = $2
		`, contractID, freelancerID, status, at)
		err = execErr
		tag = res.RowsAffected()
	case domain.StatusPendingAcceptance:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4, activated_at = null, declined_at = null, revoked_at = null
			where id = $1 and freelancer_id = $2
		`, contractID, freelancerID, status, at)
		err = execErr
		tag = res.RowsAffected()
	default:
		res, execErr := r.pool.Exec(ctx, `
			update contracts
			set status = $3, updated_at = $4
			where id = $1 and freelancer_id = $2
		`, contractID, freelancerID, status, at)
		err = execErr
		tag = res.RowsAffected()
	}
	if err != nil {
		return err
	}
	if tag == 0 {
		return ErrNotFound
	}
	return nil
}

const selectBase = `
	select id, client_id, freelancer_id, job_id, proposal_id, contract_type, status,
		title, description, hourly_rate, fixed_total, weekly_hour_limit,
		milestones, created_at, updated_at, activated_at, declined_at, revoked_at, paused_at, ended_at
	from contracts
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanContract(scanner rowScanner) (domain.Contract, error) {
	var c domain.Contract
	var proposalID *int64
	var milestonesRaw []byte
	err := scanner.Scan(
		&c.ID,
		&c.ClientID,
		&c.FreelancerID,
		&c.JobID,
		&proposalID,
		&c.ContractType,
		&c.Status,
		&c.Title,
		&c.Description,
		&c.HourlyRate,
		&c.FixedTotal,
		&c.WeeklyHourLimit,
		&milestonesRaw,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.ActivatedAt,
		&c.DeclinedAt,
		&c.RevokedAt,
		&c.PausedAt,
		&c.EndedAt,
	)
	if err != nil {
		return domain.Contract{}, err
	}
	if proposalID != nil {
		c.ProposalID = *proposalID
	}
	if len(milestonesRaw) > 0 {
		if err := json.Unmarshal(milestonesRaw, &c.Milestones); err != nil {
			return domain.Contract{}, fmt.Errorf("unmarshal milestones: %w", err)
		}
	}
	return c, nil
}

func scanContracts(rows rowSet) ([]domain.Contract, error) {
	out := make([]domain.Contract, 0)
	for rows.Next() {
		c, err := scanContract(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type rowSet interface {
	rowScanner
	Next() bool
	Err() error
}
