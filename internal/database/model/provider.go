package model

import "time"

type ProviderRecord struct {
	Id        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string `json:"name" gorm:"uniqueIndex;not null"`
	Type      string `json:"type" gorm:"not null;default:'native'"`
	Config    string `json:"config" gorm:"type:text"`
	IsEnabled bool   `json:"isEnabled" gorm:"default:true"`
	CreatedAt int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

func (ProviderRecord) TableName() string { return "providers" }

type ExternalSubscription struct {
	Id                string `json:"id" gorm:"primaryKey;size:36"`
	ProviderId        int    `json:"providerId" gorm:"index;not null"`
	ClientEmail       string `json:"clientEmail" gorm:"index;not null"`
	SubscriptionURL   string `json:"subscriptionUrl" gorm:"type:text"`
	PlanName          string `json:"planName"`
	Status            string `json:"status" gorm:"index;default:'active'"`
	TrafficLimitBytes int64  `json:"trafficLimitBytes" gorm:"default:0"`
	TrafficUsedBytes  int64  `json:"trafficUsedBytes" gorm:"default:0"`
	DeviceLimit       int    `json:"deviceLimit" gorm:"default:0"`
	ExpiresAt         int64  `json:"expiresAt" gorm:"index"`
	Metadata          string `json:"metadata" gorm:"type:text"`
	Notes             string `json:"notes" gorm:"type:text"`
	CreatedAt         int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt         int64  `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

func (ExternalSubscription) TableName() string { return "external_subscriptions" }

func (e *ExternalSubscription) IsExpired() bool {
	if e.ExpiresAt == 0 {
		return false
	}
	return e.ExpiresAt < time.Now().UnixMilli()
}

type SubscriptionEvent struct {
	Id             int    `json:"id" gorm:"primaryKey;autoIncrement"`
	SubscriptionId string `json:"subscriptionId" gorm:"index;not null"`
	ProviderId     int    `json:"providerId" gorm:"index;not null"`
	EventType      string `json:"eventType" gorm:"not null"`
	Details        string `json:"details" gorm:"type:text"`
	CreatedAt      int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
}

func (SubscriptionEvent) TableName() string { return "subscription_events" }
