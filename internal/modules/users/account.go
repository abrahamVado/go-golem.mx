package users

import (
	"strings"
	"time"
)

const (
	AccountTypeOwner         = "owner"
	AccountTypeFounder       = "founder"
	AccountTypePremiumClient = "premium_client"
	AccountTypeFreeClient    = "free_client"
	AccountTypeInvalidClient = "invalid_client"
)

type AccountSnapshot struct {
	Type                 string
	IsPremium            bool
	IsBlocked            bool
	PremiumDaysRemaining int
	FreeDaysRemaining    int
	BlockedAt            *time.Time
}

func NormalizeAccountType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case AccountTypeOwner:
		return AccountTypeOwner
	case AccountTypeFounder:
		return AccountTypeFounder
	case AccountTypePremiumClient:
		return AccountTypePremiumClient
	case AccountTypeInvalidClient:
		return AccountTypeInvalidClient
	default:
		return AccountTypeFreeClient
	}
}

func ResolveAccountSnapshot(user User, now time.Time) AccountSnapshot {
	accountType := NormalizeAccountType(user.AccountType)
	now = now.UTC()

	switch accountType {
	case AccountTypeOwner, AccountTypeFounder:
		return AccountSnapshot{
			Type:      accountType,
			IsPremium: true,
			IsBlocked: false,
		}
	case AccountTypePremiumClient:
		if user.PremiumExpiresAt != nil && user.PremiumExpiresAt.After(now) {
			return AccountSnapshot{
				Type:                 AccountTypePremiumClient,
				IsPremium:            true,
				IsBlocked:            false,
				PremiumDaysRemaining: daysRemaining(now, user.PremiumExpiresAt),
				FreeDaysRemaining:    daysRemaining(now, user.FreeExpiresAt),
			}
		}
		if user.FreeExpiresAt != nil && user.FreeExpiresAt.After(now) {
			return AccountSnapshot{
				Type:                 AccountTypeFreeClient,
				IsPremium:            false,
				IsBlocked:            false,
				PremiumDaysRemaining: 0,
				FreeDaysRemaining:    daysRemaining(now, user.FreeExpiresAt),
			}
		}
		blockedAt := user.FreeExpiresAt
		if blockedAt == nil {
			blockedAt = &now
		}
		return AccountSnapshot{
			Type:                 AccountTypeInvalidClient,
			IsPremium:            false,
			IsBlocked:            true,
			PremiumDaysRemaining: 0,
			FreeDaysRemaining:    0,
			BlockedAt:            blockedAt,
		}
	case AccountTypeFreeClient:
		if user.FreeExpiresAt != nil && user.FreeExpiresAt.After(now) {
			return AccountSnapshot{
				Type:                 AccountTypeFreeClient,
				IsPremium:            false,
				IsBlocked:            false,
				PremiumDaysRemaining: 0,
				FreeDaysRemaining:    daysRemaining(now, user.FreeExpiresAt),
			}
		}
		blockedAt := user.FreeExpiresAt
		if blockedAt == nil {
			blockedAt = &now
		}
		return AccountSnapshot{
			Type:                 AccountTypeInvalidClient,
			IsPremium:            false,
			IsBlocked:            true,
			PremiumDaysRemaining: 0,
			FreeDaysRemaining:    0,
			BlockedAt:            blockedAt,
		}
	default:
		return AccountSnapshot{
			Type:      AccountTypeInvalidClient,
			IsPremium: false,
			IsBlocked: true,
			BlockedAt: &now,
		}
	}
}

func daysRemaining(now time.Time, deadline *time.Time) int {
	if deadline == nil || !deadline.After(now) {
		return 0
	}
	duration := deadline.Sub(now)
	days := int(duration.Hours() / 24)
	if duration%(24*time.Hour) != 0 {
		days++
	}
	if days < 0 {
		return 0
	}
	return days
}
