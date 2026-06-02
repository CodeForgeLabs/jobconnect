package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
)

type ContractRepo struct {
	mu     sync.RWMutex
	nextID int64
	items  map[int64]domain.Contract
}

func NewContractRepo() *ContractRepo {
	return &ContractRepo{
		nextID: 1,
		items:  make(map[int64]domain.Contract),
	}
}

func (r *ContractRepo) Create(_ context.Context, c domain.Contract) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	c.ID = id
	r.items[id] = c
	return id, nil
}

func (r *ContractRepo) GetByID(_ context.Context, contractID int64) (domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.items[contractID]
	if !ok {
		return domain.Contract{}, fmt.Errorf("contract not found")
	}
	return c, nil
}

func (r *ContractRepo) GetByIDForActor(_ context.Context, contractID int64, actorID uuid.UUID) (domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.items[contractID]
	if !ok {
		return domain.Contract{}, fmt.Errorf("contract not found")
	}
	if c.ClientID != actorID && c.FreelancerID != actorID {
		return domain.Contract{}, fmt.Errorf("contract not found")
	}
	return c, nil
}

func (r *ContractRepo) ListByActor(_ context.Context, actorID uuid.UUID, status string, limit, offset int) ([]domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if limit <= 0 {
		limit = 20
	}
	status = strings.ToLower(strings.TrimSpace(status))
	all := make([]domain.Contract, 0, len(r.items))
	for _, c := range r.items {
		if c.ClientID != actorID && c.FreelancerID != actorID {
			continue
		}
		if status != "" && !strings.EqualFold(c.Status, status) {
			continue
		}
		all = append(all, c)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	if offset >= len(all) {
		return []domain.Contract{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *ContractRepo) SetStatusForFreelancer(_ context.Context, contractID int64, freelancerID uuid.UUID, status string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.items[contractID]
	if !ok {
		return fmt.Errorf("contract not found")
	}
	if c.FreelancerID != freelancerID {
		return fmt.Errorf("contract not found")
	}
	c.Status = status
	c.UpdatedAt = at
	switch strings.ToLower(strings.TrimSpace(status)) {
	case domain.StatusActive:
		activated := at
		c.ActivatedAt = &activated
	case domain.StatusDeclined:
		declined := at
		c.DeclinedAt = &declined
	}
	r.items[contractID] = c
	return nil
}
