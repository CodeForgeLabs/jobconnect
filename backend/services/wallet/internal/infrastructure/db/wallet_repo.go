package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"jobconnect/wallet/internal/application"
	"jobconnect/wallet/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepo struct {
	pool *pgxpool.Pool
}

func NewWalletRepo(pool *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{pool: pool}
}

func (r *WalletRepo) CreateWallet(ctx context.Context, ownerID uuid.UUID, currency string) (domain.WalletAccount, error) {
	const q = `
		INSERT INTO wallet_accounts (owner_id, currency, status, available_minor, held_minor)
		VALUES ($1, $2, $3, 0, 0)
		RETURNING id, owner_id, currency, status, available_minor, held_minor, created_at, updated_at
	`
	var w domain.WalletAccount
	err := r.pool.QueryRow(ctx, q, ownerID, domain.NormalizeCurrency(currency), domain.WalletStatusActive).
		Scan(&w.ID, &w.OwnerID, &w.Currency, &w.Status, &w.AvailableMinor, &w.HeldMinor, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.WalletAccount{}, fmt.Errorf("%w: wallet already exists for owner and currency", domain.ErrAlreadyExists)
		}
		return domain.WalletAccount{}, err
	}
	return w, nil
}

func (r *WalletRepo) GetWalletByID(ctx context.Context, walletID int64) (domain.WalletAccount, error) {
	const q = `
		SELECT id, owner_id, currency, status, available_minor, held_minor, created_at, updated_at
		FROM wallet_accounts
		WHERE id = $1
	`
	var w domain.WalletAccount
	err := r.pool.QueryRow(ctx, q, walletID).
		Scan(&w.ID, &w.OwnerID, &w.Currency, &w.Status, &w.AvailableMinor, &w.HeldMinor, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WalletAccount{}, fmt.Errorf("%w: wallet not found", domain.ErrNotFound)
	}
	if err != nil {
		return domain.WalletAccount{}, err
	}
	return w, nil
}

func (r *WalletRepo) GetWalletByOwnerCurrency(ctx context.Context, ownerID uuid.UUID, currency string) (domain.WalletAccount, error) {
	const q = `
		SELECT id, owner_id, currency, status, available_minor, held_minor, created_at, updated_at
		FROM wallet_accounts
		WHERE owner_id = $1 AND currency = $2
	`
	var w domain.WalletAccount
	err := r.pool.QueryRow(ctx, q, ownerID, domain.NormalizeCurrency(currency)).
		Scan(&w.ID, &w.OwnerID, &w.Currency, &w.Status, &w.AvailableMinor, &w.HeldMinor, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WalletAccount{}, fmt.Errorf("%w: wallet not found", domain.ErrNotFound)
	}
	if err != nil {
		return domain.WalletAccount{}, err
	}
	return w, nil
}

func (r *WalletRepo) GetBalance(ctx context.Context, walletID int64) (domain.BalanceSnapshot, error) {
	w, err := r.GetWalletByID(ctx, walletID)
	if err != nil {
		return domain.BalanceSnapshot{}, err
	}
	return domain.BalanceSnapshot{
		WalletID:       w.ID,
		Currency:       w.Currency,
		AvailableMinor: w.AvailableMinor,
		HeldMinor:      w.HeldMinor,
	}, nil
}

func (r *WalletRepo) CreditInternal(ctx context.Context, in application.CreditInput) (application.MutationResult, error) {
	return r.applySimpleMutation(ctx, in.WalletID, in.AmountMinor, in.IdempotencyKey, in.ReferenceType, in.ReferenceID, in.Note, domain.LedgerTypeCreditInternal, false)
}

func (r *WalletRepo) DebitInternal(ctx context.Context, in application.DebitInput) (application.MutationResult, error) {
	return r.applySimpleMutation(ctx, in.WalletID, in.AmountMinor, in.IdempotencyKey, in.ReferenceType, in.ReferenceID, in.Note, domain.LedgerTypeDebitInternal, true)
}

func (r *WalletRepo) applySimpleMutation(
	ctx context.Context,
	walletID int64,
	amountMinor int64,
	idempotencyKey, referenceType, referenceID, note, entryType string,
	isDebit bool,
) (application.MutationResult, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return application.MutationResult{}, fmt.Errorf("%w: idempotency_key is required", domain.ErrInvalidArgument)
	}
	if amountMinor <= 0 {
		return application.MutationResult{}, fmt.Errorf("%w: amount_minor must be greater than zero", domain.ErrInvalidArgument)
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return application.MutationResult{}, err
	}
	defer tx.Rollback(ctx)

	wallet, err := getWalletForUpdate(ctx, tx, walletID)
	if err != nil {
		return application.MutationResult{}, err
	}
	replayed, ok, err := getLedgerByIdempotency(ctx, tx, walletID, idempotencyKey)
	if err != nil {
		return application.MutationResult{}, err
	}
	if ok {
		if replayed.EntryType != entryType {
			return application.MutationResult{}, fmt.Errorf("%w: idempotency key used for another operation", domain.ErrConflict)
		}
		if err := tx.Commit(ctx); err != nil {
			return application.MutationResult{}, err
		}
		return application.MutationResult{Wallet: wallet, Entry: replayed}, nil
	}

	newAvailable := wallet.AvailableMinor
	newHeld := wallet.HeldMinor
	if isDebit {
		if wallet.AvailableMinor < amountMinor {
			return application.MutationResult{}, fmt.Errorf("%w: available balance too low", domain.ErrInsufficientFunds)
		}
		newAvailable -= amountMinor
	} else {
		newAvailable += amountMinor
	}

	wallet, err = updateWalletBalances(ctx, tx, walletID, newAvailable, newHeld)
	if err != nil {
		return application.MutationResult{}, err
	}
	entry, err := insertLedgerEntry(ctx, tx, ledgerInsertInput{
		WalletID:            walletID,
		EntryType:           entryType,
		AmountMinor:         amountMinor,
		IdempotencyKey:      idempotencyKey,
		ReferenceType:       referenceType,
		ReferenceID:         referenceID,
		Note:                note,
		AvailableAfterMinor: wallet.AvailableMinor,
		HeldAfterMinor:      wallet.HeldMinor,
	})
	if err != nil {
		if isUniqueViolation(err) {
			replayed, ok, replayErr := getLedgerByIdempotency(ctx, tx, walletID, idempotencyKey)
			if replayErr != nil {
				return application.MutationResult{}, replayErr
			}
			if ok {
				_ = tx.Commit(ctx)
				return application.MutationResult{Wallet: wallet, Entry: replayed}, nil
			}
		}
		return application.MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return application.MutationResult{}, err
	}
	return application.MutationResult{Wallet: wallet, Entry: entry}, nil
}

func (r *WalletRepo) PlaceHold(ctx context.Context, in application.PlaceHoldInput) (application.HoldMutationResult, error) {
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return application.HoldMutationResult{}, fmt.Errorf("%w: idempotency_key is required", domain.ErrInvalidArgument)
	}
	if in.AmountMinor <= 0 {
		return application.HoldMutationResult{}, fmt.Errorf("%w: amount_minor must be greater than zero", domain.ErrInvalidArgument)
	}
	if strings.TrimSpace(in.ReferenceType) == "" || strings.TrimSpace(in.ReferenceID) == "" {
		return application.HoldMutationResult{}, fmt.Errorf("%w: reference_type and reference_id are required", domain.ErrInvalidArgument)
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	defer tx.Rollback(ctx)

	wallet, err := getWalletForUpdate(ctx, tx, in.WalletID)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	replayed, ok, err := getLedgerByIdempotency(ctx, tx, in.WalletID, in.IdempotencyKey)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	if ok {
		if replayed.EntryType != domain.LedgerTypeHoldPlaced {
			return application.HoldMutationResult{}, fmt.Errorf("%w: idempotency key used for another operation", domain.ErrConflict)
		}
		hold, holdErr := getHoldByReference(ctx, tx, wallet.ID, in.ReferenceType, in.ReferenceID)
		if holdErr != nil {
			return application.HoldMutationResult{}, holdErr
		}
		if err := tx.Commit(ctx); err != nil {
			return application.HoldMutationResult{}, err
		}
		return application.HoldMutationResult{Wallet: wallet, Hold: hold, Entry: replayed}, nil
	}

	if wallet.AvailableMinor < in.AmountMinor {
		return application.HoldMutationResult{}, fmt.Errorf("%w: available balance too low", domain.ErrInsufficientFunds)
	}

	hold, err := insertHold(ctx, tx, holdInsertInput{
		WalletID:      wallet.ID,
		ReferenceType: strings.TrimSpace(in.ReferenceType),
		ReferenceID:   strings.TrimSpace(in.ReferenceID),
		AmountMinor:   in.AmountMinor,
		ExpiresAt:     in.ExpiresAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return application.HoldMutationResult{}, fmt.Errorf("%w: hold already exists for reference", domain.ErrAlreadyExists)
		}
		return application.HoldMutationResult{}, err
	}

	wallet, err = updateWalletBalances(ctx, tx, wallet.ID, wallet.AvailableMinor-in.AmountMinor, wallet.HeldMinor+in.AmountMinor)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	entry, err := insertLedgerEntry(ctx, tx, ledgerInsertInput{
		WalletID:            wallet.ID,
		EntryType:           domain.LedgerTypeHoldPlaced,
		AmountMinor:         in.AmountMinor,
		IdempotencyKey:      strings.TrimSpace(in.IdempotencyKey),
		ReferenceType:       strings.TrimSpace(in.ReferenceType),
		ReferenceID:         strings.TrimSpace(in.ReferenceID),
		Note:                strings.TrimSpace(in.Note),
		AvailableAfterMinor: wallet.AvailableMinor,
		HeldAfterMinor:      wallet.HeldMinor,
	})
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return application.HoldMutationResult{}, err
	}
	return application.HoldMutationResult{Wallet: wallet, Hold: hold, Entry: entry}, nil
}

func (r *WalletRepo) ReleaseHold(ctx context.Context, in application.ReleaseHoldInput) (application.HoldMutationResult, error) {
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return application.HoldMutationResult{}, fmt.Errorf("%w: idempotency_key is required", domain.ErrInvalidArgument)
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	defer tx.Rollback(ctx)

	hold, err := getHoldByIDForUpdate(ctx, tx, in.HoldID)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	wallet, err := getWalletForUpdate(ctx, tx, hold.WalletID)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	replayed, ok, err := getLedgerByIdempotency(ctx, tx, wallet.ID, in.IdempotencyKey)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	if ok {
		if replayed.EntryType != domain.LedgerTypeHoldReleased {
			return application.HoldMutationResult{}, fmt.Errorf("%w: idempotency key used for another operation", domain.ErrConflict)
		}
		if err := tx.Commit(ctx); err != nil {
			return application.HoldMutationResult{}, err
		}
		return application.HoldMutationResult{Wallet: wallet, Hold: hold, Entry: replayed}, nil
	}

	if hold.Status != domain.HoldStatusActive {
		return application.HoldMutationResult{}, fmt.Errorf("%w: hold is not active", domain.ErrConflict)
	}
	remaining := hold.RemainingMinor()
	if remaining <= 0 {
		return application.HoldMutationResult{}, fmt.Errorf("%w: hold has no releasable amount", domain.ErrConflict)
	}
	if wallet.HeldMinor < remaining {
		return application.HoldMutationResult{}, fmt.Errorf("%w: held balance is inconsistent", domain.ErrConflict)
	}

	hold, err = updateHoldRelease(ctx, tx, hold.ID)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	wallet, err = updateWalletBalances(ctx, tx, wallet.ID, wallet.AvailableMinor+remaining, wallet.HeldMinor-remaining)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	entry, err := insertLedgerEntry(ctx, tx, ledgerInsertInput{
		WalletID:            wallet.ID,
		EntryType:           domain.LedgerTypeHoldReleased,
		AmountMinor:         remaining,
		IdempotencyKey:      strings.TrimSpace(in.IdempotencyKey),
		ReferenceType:       hold.ReferenceType,
		ReferenceID:         hold.ReferenceID,
		Note:                strings.TrimSpace(in.Note),
		AvailableAfterMinor: wallet.AvailableMinor,
		HeldAfterMinor:      wallet.HeldMinor,
	})
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return application.HoldMutationResult{}, err
	}
	return application.HoldMutationResult{Wallet: wallet, Hold: hold, Entry: entry}, nil
}

func (r *WalletRepo) CaptureHold(ctx context.Context, in application.CaptureHoldInput) (application.HoldMutationResult, error) {
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return application.HoldMutationResult{}, fmt.Errorf("%w: idempotency_key is required", domain.ErrInvalidArgument)
	}
	if in.CaptureAmountMinor <= 0 {
		return application.HoldMutationResult{}, fmt.Errorf("%w: capture_amount_minor must be greater than zero", domain.ErrInvalidArgument)
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	defer tx.Rollback(ctx)

	hold, err := getHoldByIDForUpdate(ctx, tx, in.HoldID)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	wallet, err := getWalletForUpdate(ctx, tx, hold.WalletID)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	replayed, ok, err := getLedgerByIdempotency(ctx, tx, wallet.ID, in.IdempotencyKey)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	if ok {
		if replayed.EntryType != domain.LedgerTypeHoldCaptured {
			return application.HoldMutationResult{}, fmt.Errorf("%w: idempotency key used for another operation", domain.ErrConflict)
		}
		if err := tx.Commit(ctx); err != nil {
			return application.HoldMutationResult{}, err
		}
		return application.HoldMutationResult{Wallet: wallet, Hold: hold, Entry: replayed}, nil
	}

	if hold.Status != domain.HoldStatusActive {
		return application.HoldMutationResult{}, fmt.Errorf("%w: hold is not active", domain.ErrConflict)
	}
	remaining := hold.RemainingMinor()
	if in.CaptureAmountMinor > remaining {
		return application.HoldMutationResult{}, fmt.Errorf("%w: capture amount exceeds hold remainder", domain.ErrInvalidArgument)
	}
	if wallet.HeldMinor < in.CaptureAmountMinor {
		return application.HoldMutationResult{}, fmt.Errorf("%w: held balance is inconsistent", domain.ErrConflict)
	}

	hold, err = updateHoldCapture(ctx, tx, hold.ID, hold.CapturedMinor+in.CaptureAmountMinor, hold.AmountMinor)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	wallet, err = updateWalletBalances(ctx, tx, wallet.ID, wallet.AvailableMinor, wallet.HeldMinor-in.CaptureAmountMinor)
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	entry, err := insertLedgerEntry(ctx, tx, ledgerInsertInput{
		WalletID:            wallet.ID,
		EntryType:           domain.LedgerTypeHoldCaptured,
		AmountMinor:         in.CaptureAmountMinor,
		IdempotencyKey:      strings.TrimSpace(in.IdempotencyKey),
		ReferenceType:       strings.TrimSpace(in.ReferenceType),
		ReferenceID:         strings.TrimSpace(in.ReferenceID),
		Note:                strings.TrimSpace(in.Note),
		AvailableAfterMinor: wallet.AvailableMinor,
		HeldAfterMinor:      wallet.HeldMinor,
	})
	if err != nil {
		return application.HoldMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return application.HoldMutationResult{}, err
	}
	return application.HoldMutationResult{Wallet: wallet, Hold: hold, Entry: entry}, nil
}

func (r *WalletRepo) ListLedgerEntries(ctx context.Context, walletID int64, limit, offset int) ([]domain.LedgerEntry, error) {
	const q = `
		SELECT id, wallet_id, entry_type, amount_minor, idempotency_key, reference_type, reference_id, note,
		       available_after_minor, held_after_minor, created_at
		FROM wallet_ledger_entries
		WHERE wallet_id = $1
		ORDER BY id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, q, walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.LedgerEntry, 0, limit)
	for rows.Next() {
		var item domain.LedgerEntry
		if err := rows.Scan(
			&item.ID,
			&item.WalletID,
			&item.EntryType,
			&item.AmountMinor,
			&item.IdempotencyKey,
			&item.ReferenceType,
			&item.ReferenceID,
			&item.Note,
			&item.AvailableAfterMinor,
			&item.HeldAfterMinor,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return result, nil
}

func getWalletForUpdate(ctx context.Context, tx pgx.Tx, walletID int64) (domain.WalletAccount, error) {
	const q = `
		SELECT id, owner_id, currency, status, available_minor, held_minor, created_at, updated_at
		FROM wallet_accounts
		WHERE id = $1
		FOR UPDATE
	`
	var w domain.WalletAccount
	err := tx.QueryRow(ctx, q, walletID).
		Scan(&w.ID, &w.OwnerID, &w.Currency, &w.Status, &w.AvailableMinor, &w.HeldMinor, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WalletAccount{}, fmt.Errorf("%w: wallet not found", domain.ErrNotFound)
	}
	if err != nil {
		return domain.WalletAccount{}, err
	}
	return w, nil
}

func updateWalletBalances(ctx context.Context, tx pgx.Tx, walletID, availableMinor, heldMinor int64) (domain.WalletAccount, error) {
	const q = `
		UPDATE wallet_accounts
		SET available_minor = $2,
		    held_minor = $3,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, owner_id, currency, status, available_minor, held_minor, created_at, updated_at
	`
	var w domain.WalletAccount
	err := tx.QueryRow(ctx, q, walletID, availableMinor, heldMinor).
		Scan(&w.ID, &w.OwnerID, &w.Currency, &w.Status, &w.AvailableMinor, &w.HeldMinor, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return domain.WalletAccount{}, err
	}
	return w, nil
}

type ledgerInsertInput struct {
	WalletID            int64
	EntryType           string
	AmountMinor         int64
	IdempotencyKey      string
	ReferenceType       string
	ReferenceID         string
	Note                string
	AvailableAfterMinor int64
	HeldAfterMinor      int64
}

func insertLedgerEntry(ctx context.Context, tx pgx.Tx, in ledgerInsertInput) (domain.LedgerEntry, error) {
	const q = `
		INSERT INTO wallet_ledger_entries (
			wallet_id, entry_type, amount_minor, idempotency_key, reference_type, reference_id, note,
			available_after_minor, held_after_minor
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, wallet_id, entry_type, amount_minor, idempotency_key, reference_type, reference_id, note,
		          available_after_minor, held_after_minor, created_at
	`
	var item domain.LedgerEntry
	err := tx.QueryRow(ctx, q,
		in.WalletID,
		strings.TrimSpace(in.EntryType),
		in.AmountMinor,
		strings.TrimSpace(in.IdempotencyKey),
		strings.TrimSpace(in.ReferenceType),
		strings.TrimSpace(in.ReferenceID),
		strings.TrimSpace(in.Note),
		in.AvailableAfterMinor,
		in.HeldAfterMinor,
	).Scan(
		&item.ID,
		&item.WalletID,
		&item.EntryType,
		&item.AmountMinor,
		&item.IdempotencyKey,
		&item.ReferenceType,
		&item.ReferenceID,
		&item.Note,
		&item.AvailableAfterMinor,
		&item.HeldAfterMinor,
		&item.CreatedAt,
	)
	if err != nil {
		return domain.LedgerEntry{}, err
	}
	return item, nil
}

func getLedgerByIdempotency(ctx context.Context, tx pgx.Tx, walletID int64, key string) (domain.LedgerEntry, bool, error) {
	const q = `
		SELECT id, wallet_id, entry_type, amount_minor, idempotency_key, reference_type, reference_id, note,
		       available_after_minor, held_after_minor, created_at
		FROM wallet_ledger_entries
		WHERE wallet_id = $1 AND idempotency_key = $2
	`
	var item domain.LedgerEntry
	err := tx.QueryRow(ctx, q, walletID, strings.TrimSpace(key)).Scan(
		&item.ID,
		&item.WalletID,
		&item.EntryType,
		&item.AmountMinor,
		&item.IdempotencyKey,
		&item.ReferenceType,
		&item.ReferenceID,
		&item.Note,
		&item.AvailableAfterMinor,
		&item.HeldAfterMinor,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LedgerEntry{}, false, nil
	}
	if err != nil {
		return domain.LedgerEntry{}, false, err
	}
	return item, true, nil
}

type holdInsertInput struct {
	WalletID      int64
	ReferenceType string
	ReferenceID   string
	AmountMinor   int64
	ExpiresAt     *time.Time
}

func insertHold(ctx context.Context, tx pgx.Tx, in holdInsertInput) (domain.Hold, error) {
	const q = `
		INSERT INTO wallet_holds (wallet_id, reference_type, reference_id, amount_minor, captured_minor, status, expires_at)
		VALUES ($1, $2, $3, $4, 0, $5, $6)
		RETURNING id, wallet_id, reference_type, reference_id, amount_minor, captured_minor, status,
		          expires_at, created_at, updated_at
	`
	var hold domain.Hold
	err := tx.QueryRow(ctx, q,
		in.WalletID,
		strings.TrimSpace(in.ReferenceType),
		strings.TrimSpace(in.ReferenceID),
		in.AmountMinor,
		domain.HoldStatusActive,
		in.ExpiresAt,
	).Scan(
		&hold.ID,
		&hold.WalletID,
		&hold.ReferenceType,
		&hold.ReferenceID,
		&hold.AmountMinor,
		&hold.CapturedMinor,
		&hold.Status,
		&hold.ExpiresAt,
		&hold.CreatedAt,
		&hold.UpdatedAt,
	)
	if err != nil {
		return domain.Hold{}, err
	}
	return hold, nil
}

func getHoldByIDForUpdate(ctx context.Context, tx pgx.Tx, holdID int64) (domain.Hold, error) {
	const q = `
		SELECT id, wallet_id, reference_type, reference_id, amount_minor, captured_minor, status,
		       expires_at, created_at, updated_at
		FROM wallet_holds
		WHERE id = $1
		FOR UPDATE
	`
	var hold domain.Hold
	err := tx.QueryRow(ctx, q, holdID).Scan(
		&hold.ID,
		&hold.WalletID,
		&hold.ReferenceType,
		&hold.ReferenceID,
		&hold.AmountMinor,
		&hold.CapturedMinor,
		&hold.Status,
		&hold.ExpiresAt,
		&hold.CreatedAt,
		&hold.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Hold{}, fmt.Errorf("%w: hold not found", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Hold{}, err
	}
	return hold, nil
}

func getHoldByReference(ctx context.Context, tx pgx.Tx, walletID int64, referenceType, referenceID string) (domain.Hold, error) {
	const q = `
		SELECT id, wallet_id, reference_type, reference_id, amount_minor, captured_minor, status,
		       expires_at, created_at, updated_at
		FROM wallet_holds
		WHERE wallet_id = $1 AND reference_type = $2 AND reference_id = $3
	`
	var hold domain.Hold
	err := tx.QueryRow(ctx, q, walletID, strings.TrimSpace(referenceType), strings.TrimSpace(referenceID)).Scan(
		&hold.ID,
		&hold.WalletID,
		&hold.ReferenceType,
		&hold.ReferenceID,
		&hold.AmountMinor,
		&hold.CapturedMinor,
		&hold.Status,
		&hold.ExpiresAt,
		&hold.CreatedAt,
		&hold.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Hold{}, fmt.Errorf("%w: hold not found", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Hold{}, err
	}
	return hold, nil
}

func updateHoldRelease(ctx context.Context, tx pgx.Tx, holdID int64) (domain.Hold, error) {
	const q = `
		UPDATE wallet_holds
		SET status = $2,
		    updated_at = NOW(),
		    released_at = NOW()
		WHERE id = $1
		RETURNING id, wallet_id, reference_type, reference_id, amount_minor, captured_minor, status,
		          expires_at, created_at, updated_at
	`
	var hold domain.Hold
	err := tx.QueryRow(ctx, q, holdID, domain.HoldStatusReleased).Scan(
		&hold.ID,
		&hold.WalletID,
		&hold.ReferenceType,
		&hold.ReferenceID,
		&hold.AmountMinor,
		&hold.CapturedMinor,
		&hold.Status,
		&hold.ExpiresAt,
		&hold.CreatedAt,
		&hold.UpdatedAt,
	)
	if err != nil {
		return domain.Hold{}, err
	}
	return hold, nil
}

func updateHoldCapture(ctx context.Context, tx pgx.Tx, holdID, newCaptured, totalAmount int64) (domain.Hold, error) {
	status := domain.HoldStatusActive
	if newCaptured >= totalAmount {
		status = domain.HoldStatusCaptured
	}
	const q = `
		UPDATE wallet_holds
		SET captured_minor = $2,
		    status = $3,
		    updated_at = NOW(),
		    captured_at = CASE WHEN $3 = 'captured' THEN NOW() ELSE captured_at END
		WHERE id = $1
		RETURNING id, wallet_id, reference_type, reference_id, amount_minor, captured_minor, status,
		          expires_at, created_at, updated_at
	`
	var hold domain.Hold
	err := tx.QueryRow(ctx, q, holdID, newCaptured, status).Scan(
		&hold.ID,
		&hold.WalletID,
		&hold.ReferenceType,
		&hold.ReferenceID,
		&hold.AmountMinor,
		&hold.CapturedMinor,
		&hold.Status,
		&hold.ExpiresAt,
		&hold.CreatedAt,
		&hold.UpdatedAt,
	)
	if err != nil {
		return domain.Hold{}, err
	}
	return hold, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
