package provider

import (
	"context"
	"errors"
	"time"
)

type ProviderType string

const (
	ProviderTypeNative   ProviderType = "native"
	ProviderTypeExternal ProviderType = "external"
	ProviderTypeStatic   ProviderType = "static"
)

var (
	ErrProviderUnavailable  = errors.New("provider unavailable")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrQuotaExceeded        = errors.New("quota exceeded")
)

type Provider interface {
	Name() string
	Type() ProviderType
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error)
	RenewSubscription(ctx context.Context, id string, req RenewRequest) error
	SuspendSubscription(ctx context.Context, id string) error
	ResumeSubscription(ctx context.Context, id string) error
	DeleteSubscription(ctx context.Context, id string) error
	GetSubscription(ctx context.Context, id string) (*Subscription, error)
	GetSubscriptionURL(ctx context.Context, id string) (string, error)
	GetStatus(ctx context.Context, id string) (*SubStatus, error)
	SyncUsage(ctx context.Context, id string) (*UsageReport, error)
	IsHealthy(ctx context.Context) bool
}

type CreateSubscriptionRequest struct {
	Email             string `json:"email"`
	PlanName          string `json:"planName"`
	TrafficLimitBytes int64  `json:"trafficLimitBytes"`
	DeviceLimit       int    `json:"deviceLimit"`
	DurationDays      int    `json:"durationDays"`
	Notes             string `json:"notes,omitempty"`
}

type RenewRequest struct {
	TrafficLimitBytes int64 `json:"trafficLimitBytes,omitempty"`
	DurationDays      int   `json:"durationDays"`
}

type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusSuspended SubscriptionStatus = "suspended"
	StatusExpired   SubscriptionStatus = "expired"
	StatusDeleted   SubscriptionStatus = "deleted"
)

type Subscription struct {
	ID                string             `json:"id"`
	ProviderName      string             `json:"providerName"`
	Email             string             `json:"email"`
	PlanName          string             `json:"planName"`
	Status            SubscriptionStatus `json:"status"`
	SubscriptionURL   string             `json:"subscriptionUrl,omitempty"`
	TrafficLimitBytes int64              `json:"trafficLimitBytes"`
	TrafficUsedBytes  int64              `json:"trafficUsedBytes"`
	DeviceLimit       int                `json:"deviceLimit"`
	ExpiresAt         time.Time          `json:"expiresAt"`
	CreatedAt         time.Time          `json:"createdAt"`
	Metadata          map[string]any     `json:"metadata,omitempty"`
}

type SubStatus struct {
	ID                string             `json:"id"`
	Status            SubscriptionStatus `json:"status"`
	TrafficLimitBytes int64              `json:"trafficLimitBytes"`
	TrafficUsedBytes  int64              `json:"trafficUsedBytes"`
	UploadBytes       int64              `json:"uploadBytes"`
	DownloadBytes     int64              `json:"downloadBytes"`
	DeviceLimit       int                `json:"deviceLimit"`
	ExpiresAt         time.Time          `json:"expiresAt"`
	Online            bool               `json:"online"`
}

type UsageReport struct {
	ID            string    `json:"id"`
	UploadBytes   int64     `json:"uploadBytes"`
	DownloadBytes int64     `json:"downloadBytes"`
	TotalBytes    int64     `json:"totalBytes"`
	LastSyncAt    time.Time `json:"lastSyncAt"`
}
