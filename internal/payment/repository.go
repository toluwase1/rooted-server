package payment

import "context"

type Repository interface {
	// Subscriptions
	CreateSubscription(ctx context.Context, sub Subscription) (*Subscription, error)
	GetActiveSubscription(ctx context.Context, userID string) (*Subscription, error)
	UpdateSubscriptionStatus(ctx context.Context, id, status string) error
	GetExpiringSubscriptions(ctx context.Context) ([]Subscription, error)

	// Transactions
	RecordTransaction(ctx context.Context, tx Transaction) (*Transaction, error)
	GetTransactions(ctx context.Context, userID string, limit int) ([]Transaction, error)
	GetTransactionByProviderRef(ctx context.Context, provider, ref string) (*Transaction, error)
}
