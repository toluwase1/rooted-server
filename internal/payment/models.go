package payment

import "time"

type Subscription struct {
	ID                       string     `json:"id"`
	UserID                   string     `json:"user_id"`
	Plan                     string     `json:"plan"` // plus, premium
	StarAmount               int        `json:"star_amount"`
	TelegramPaymentChargeID  string     `json:"telegram_payment_charge_id,omitempty"`
	Status                   string     `json:"status"` // active, cancelled, expired
	CurrentPeriodStart       *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd         *time.Time `json:"current_period_end,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
}

type Transaction struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Type         string     `json:"type"` // subscription, rose, boost, super_boost, reopen_chat, gift, event_ticket
	StarAmount   *int       `json:"star_amount,omitempty"`
	FiatAmount   *float64   `json:"fiat_amount,omitempty"`
	FiatCurrency *string    `json:"fiat_currency,omitempty"`
	Provider     string     `json:"provider"` // telegram_stars, paystack, flutterwave
	ProviderRef  string     `json:"provider_ref,omitempty"`
	Status       string     `json:"status"` // completed, pending, refunded
	CreatedAt    time.Time  `json:"created_at"`
}

// PurchaseItem represents something a user can buy.
type PurchaseItem struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Description string `json:"description"`
	StarPrice   int    `json:"star_price"`
}

// SubscriptionPlan describes a subscription tier.
type SubscriptionPlan struct {
	Plan        string   `json:"plan"`
	Label       string   `json:"label"`
	StarPrice   int      `json:"star_price"`
	Features    []string `json:"features"`
}
