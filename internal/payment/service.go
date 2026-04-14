package payment

import (
	"context"
	"fmt"

	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/telegram"
)

type Service struct {
	repo   Repository
	bot    *telegram.Bot
	config *config.DynamicConfig
}

func NewService(repo Repository, bot *telegram.Bot, cfg *config.DynamicConfig) *Service {
	return &Service{repo: repo, bot: bot, config: cfg}
}

// GetPlans returns available subscription plans with current pricing from admin config.
func (s *Service) GetPlans(ctx context.Context) []SubscriptionPlan {
	plusPrice := s.config.GetInt(ctx, "plus_price_stars", 250)
	premiumPrice := s.config.GetInt(ctx, "premium_price_stars", 500)

	return []SubscriptionPlan{
		{
			Plan:      "plus",
			Label:     "Rooted Plus",
			StarPrice: plusPrice,
			Features: []string{
				"15 daily matches",
				"Unlimited swipes",
				"See who liked you",
				"Advanced filters",
				"Read receipts",
				"Ad-free",
				"1 Spotlight Boost/month",
			},
		},
		{
			Plan:      "premium",
			Label:     "Rooted Premium",
			StarPrice: premiumPrice,
			Features: []string{
				"Everything in Plus",
				"Crossed Continents",
				"Incognito mode",
				"Priority matching",
				"3 Roses/week",
				"3 Spotlight Boosts/month",
				"Message translation",
				"Extended voice notes",
			},
		},
	}
}

// GetPurchaseItems returns available a la carte items with current pricing.
func (s *Service) GetPurchaseItems(ctx context.Context) []PurchaseItem {
	return []PurchaseItem{
		{
			Type:        "rose_1",
			Label:       "Rose",
			Description: "Send a strong interest signal",
			StarPrice:   s.config.GetInt(ctx, "rose_price_stars_1", 15),
		},
		{
			Type:        "rose_5",
			Label:       "5 Roses",
			Description: "Bulk pack — save 20%",
			StarPrice:   s.config.GetInt(ctx, "rose_price_stars_5", 60),
		},
		{
			Type:        "boost",
			Label:       "Spotlight Boost",
			Description: "30 min visibility boost",
			StarPrice:   s.config.GetInt(ctx, "boost_price_stars", 75),
		},
		{
			Type:        "super_boost",
			Label:       "Super Boost",
			Description: "3 hour visibility boost",
			StarPrice:   s.config.GetInt(ctx, "super_boost_price_stars", 200),
		},
		{
			Type:        "reopen_chat",
			Label:       "Reopen Chat",
			Description: "Restart an expired conversation",
			StarPrice:   s.config.GetInt(ctx, "reopen_chat_price_stars", 30),
		},
	}
}

// SendSubscriptionInvoice sends a Star payment invoice for a subscription.
func (s *Service) SendSubscriptionInvoice(ctx context.Context, chatID int64, plan string) error {
	plans := s.GetPlans(ctx)
	var selected *SubscriptionPlan
	for _, p := range plans {
		if p.Plan == plan {
			selected = &p
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("unknown plan: %s", plan)
	}

	payload := fmt.Sprintf("subscription:%s", plan)
	return s.bot.SendInvoice(ctx, chatID,
		selected.Label,
		fmt.Sprintf("Monthly subscription — %d Stars", selected.StarPrice),
		payload,
		selected.StarPrice,
	)
}

// SendItemInvoice sends a Star payment invoice for an a la carte purchase.
func (s *Service) SendItemInvoice(ctx context.Context, chatID int64, itemType string) error {
	items := s.GetPurchaseItems(ctx)
	var selected *PurchaseItem
	for _, item := range items {
		if item.Type == itemType {
			selected = &item
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("unknown item: %s", itemType)
	}

	payload := fmt.Sprintf("purchase:%s", itemType)
	return s.bot.SendInvoice(ctx, chatID,
		selected.Label,
		selected.Description,
		payload,
		selected.StarPrice,
	)
}

// HandleSuccessfulPayment processes a completed Star payment.
func (s *Service) HandleSuccessfulPayment(ctx context.Context, userID string, payload, chargeID string, amount int) error {
	// Parse payload to determine what was purchased
	var txType, plan string

	if len(payload) > 13 && payload[:13] == "subscription:" {
		plan = payload[13:]
		txType = "subscription"
	} else if len(payload) > 9 && payload[:9] == "purchase:" {
		txType = payload[9:]
	} else {
		return fmt.Errorf("unknown payment payload: %s", payload)
	}

	// Record transaction
	starAmount := amount
	_, err := s.repo.RecordTransaction(ctx, Transaction{
		UserID:     userID,
		Type:       txType,
		StarAmount: &starAmount,
		Provider:   "telegram_stars",
		ProviderRef: chargeID,
		Status:     "completed",
	})
	if err != nil {
		return fmt.Errorf("recording transaction: %w", err)
	}

	// Handle subscription activation
	if txType == "subscription" {
		_, err := s.repo.CreateSubscription(ctx, Subscription{
			UserID:                  userID,
			Plan:                    plan,
			StarAmount:              amount,
			TelegramPaymentChargeID: chargeID,
		})
		if err != nil {
			return fmt.Errorf("creating subscription: %w", err)
		}
	}

	return nil
}

// GetActiveSubscription returns the user's current subscription.
func (s *Service) GetActiveSubscription(ctx context.Context, userID string) (*Subscription, error) {
	return s.repo.GetActiveSubscription(ctx, userID)
}

// GetTransactionHistory returns recent transactions.
func (s *Service) GetTransactionHistory(ctx context.Context, userID string) ([]Transaction, error) {
	return s.repo.GetTransactions(ctx, userID, 50)
}

// ExpireSubscriptions finds and expires lapsed subscriptions.
func (s *Service) ExpireSubscriptions(ctx context.Context) (int, error) {
	subs, err := s.repo.GetExpiringSubscriptions(ctx)
	if err != nil {
		return 0, err
	}
	for _, sub := range subs {
		s.repo.UpdateSubscriptionStatus(ctx, sub.ID, "expired")
	}
	return len(subs), nil
}
