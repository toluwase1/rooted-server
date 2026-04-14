package payment

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateSubscription(ctx context.Context, sub Subscription) (*Subscription, error) {
	var s Subscription
	err := r.db.QueryRow(ctx, `
		INSERT INTO subscriptions (user_id, plan, star_amount, telegram_payment_charge_id,
			status, current_period_start, current_period_end)
		VALUES ($1, $2, $3, $4, 'active', NOW(), NOW() + INTERVAL '30 days')
		RETURNING id, user_id, plan, star_amount, telegram_payment_charge_id,
		          status, current_period_start, current_period_end, created_at
	`, sub.UserID, sub.Plan, sub.StarAmount, sub.TelegramPaymentChargeID).Scan(
		&s.ID, &s.UserID, &s.Plan, &s.StarAmount, &s.TelegramPaymentChargeID,
		&s.Status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating subscription: %w", err)
	}
	return &s, nil
}

func (r *PostgresRepo) GetActiveSubscription(ctx context.Context, userID string) (*Subscription, error) {
	var s Subscription
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, plan, star_amount, telegram_payment_charge_id,
		       status, current_period_start, current_period_end, created_at
		FROM subscriptions
		WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(
		&s.ID, &s.UserID, &s.Plan, &s.StarAmount, &s.TelegramPaymentChargeID,
		&s.Status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepo) UpdateSubscriptionStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE subscriptions SET status = $2 WHERE id = $1", id, status)
	return err
}

func (r *PostgresRepo) GetExpiringSubscriptions(ctx context.Context) ([]Subscription, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, plan, star_amount, telegram_payment_charge_id,
		       status, current_period_start, current_period_end, created_at
		FROM subscriptions
		WHERE status = 'active' AND current_period_end < NOW()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var s Subscription
		rows.Scan(&s.ID, &s.UserID, &s.Plan, &s.StarAmount, &s.TelegramPaymentChargeID,
			&s.Status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt)
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *PostgresRepo) RecordTransaction(ctx context.Context, tx Transaction) (*Transaction, error) {
	var t Transaction
	err := r.db.QueryRow(ctx, `
		INSERT INTO transactions (user_id, type, star_amount, fiat_amount, fiat_currency,
			provider, provider_ref, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, type, star_amount, fiat_amount, fiat_currency,
		          provider, provider_ref, status, created_at
	`, tx.UserID, tx.Type, tx.StarAmount, tx.FiatAmount, tx.FiatCurrency,
		tx.Provider, tx.ProviderRef, tx.Status).Scan(
		&t.ID, &t.UserID, &t.Type, &t.StarAmount, &t.FiatAmount, &t.FiatCurrency,
		&t.Provider, &t.ProviderRef, &t.Status, &t.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("recording transaction: %w", err)
	}
	return &t, nil
}

func (r *PostgresRepo) GetTransactions(ctx context.Context, userID string, limit int) ([]Transaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, type, star_amount, fiat_amount, fiat_currency,
		       provider, provider_ref, status, created_at
		FROM transactions WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []Transaction
	for rows.Next() {
		var t Transaction
		rows.Scan(&t.ID, &t.UserID, &t.Type, &t.StarAmount, &t.FiatAmount, &t.FiatCurrency,
			&t.Provider, &t.ProviderRef, &t.Status, &t.CreatedAt)
		txns = append(txns, t)
	}
	return txns, nil
}

func (r *PostgresRepo) GetTransactionByProviderRef(ctx context.Context, provider, ref string) (*Transaction, error) {
	var t Transaction
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, type, star_amount, fiat_amount, fiat_currency,
		       provider, provider_ref, status, created_at
		FROM transactions WHERE provider = $1 AND provider_ref = $2
	`, provider, ref).Scan(
		&t.ID, &t.UserID, &t.Type, &t.StarAmount, &t.FiatAmount, &t.FiatCurrency,
		&t.Provider, &t.ProviderRef, &t.Status, &t.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
