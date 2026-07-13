package provider

import (
	"context"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"

	"gorm.io/gorm"
)

type XrayProvider struct {
	name       string
	settingSvc interface {
		BuildSubURIBase(host string) string
		GetSubPath() (string, error)
	}
}

func NewXrayProvider(name string, settingSvc interface {
	BuildSubURIBase(host string) string
	GetSubPath() (string, error)
}) *XrayProvider {
	return &XrayProvider{name: name, settingSvc: settingSvc}
}

func (p *XrayProvider) Name() string        { return p.name }
func (p *XrayProvider) Type() ProviderType  { return ProviderTypeNative }

func (p *XrayProvider) CreateSubscription(_ context.Context, req CreateSubscriptionRequest) (*Subscription, error) {
	db := database.GetDB()
	rec := &model.ClientRecord{
		Email:      req.Email,
		Enable:     true,
		TotalGB:    req.TrafficLimitBytes,
		ExpiryTime: time.Now().Add(time.Duration(req.DurationDays) * 24 * time.Hour).UnixMilli(),
		LimitIP:    req.DeviceLimit,
		Comment:    req.Notes,
	}
	if err := db.Create(rec).Error; err != nil {
		return nil, err
	}
	return p.recordToSubscription(rec, nil), nil
}

func (p *XrayProvider) RenewSubscription(_ context.Context, id string, req RenewRequest) error {
	db := database.GetDB()
	rec, err := p.findRecord(db, id)
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	base := rec.ExpiryTime
	if base < now {
		base = now
	}
	newExpiry := base + int64(req.DurationDays)*24*60*60*1000
	updates := map[string]any{"expiry_time": newExpiry}
	if req.TrafficLimitBytes > 0 {
		updates["total_gb"] = req.TrafficLimitBytes
	}
	return db.Model(&model.ClientRecord{}).Where("email = ?", id).Updates(updates).Error
}

func (p *XrayProvider) SuspendSubscription(_ context.Context, id string) error {
	return database.GetDB().Model(&model.ClientRecord{}).Where("email = ?", id).Update("enable", false).Error
}

func (p *XrayProvider) ResumeSubscription(_ context.Context, id string) error {
	return database.GetDB().Model(&model.ClientRecord{}).Where("email = ?", id).Update("enable", true).Error
}

func (p *XrayProvider) DeleteSubscription(_ context.Context, id string) error {
	db := database.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		var rec model.ClientRecord
		if err := tx.Where("email = ?", id).First(&rec).Error; err != nil {
			return err
		}
		if err := tx.Where("client_id = ?", rec.Id).Delete(&model.ClientInbound{}).Error; err != nil {
			return err
		}
		if err := tx.Where("email = ?", id).Delete(&xray.ClientTraffic{}).Error; err != nil {
			return err
		}
		return tx.Delete(&rec).Error
	})
}

func (p *XrayProvider) GetSubscription(_ context.Context, id string) (*Subscription, error) {
	db := database.GetDB()
	rec, err := p.findRecord(db, id)
	if err != nil {
		return nil, err
	}
	var traffic xray.ClientTraffic
	_ = db.Where("email = ?", id).First(&traffic).Error
	return p.recordToSubscription(rec, &traffic), nil
}

func (p *XrayProvider) GetSubscriptionURL(_ context.Context, id string) (string, error) {
	db := database.GetDB()
	rec, err := p.findRecord(db, id)
	if err != nil {
		return "", err
	}
	if rec.SubID == "" {
		return "", ErrSubscriptionNotFound
	}
	base := p.settingSvc.BuildSubURIBase("")
	subPath, _ := p.settingSvc.GetSubPath()
	return base + subPath + rec.SubID, nil
}

func (p *XrayProvider) GetStatus(_ context.Context, id string) (*SubStatus, error) {
	db := database.GetDB()
	rec, err := p.findRecord(db, id)
	if err != nil {
		return nil, err
	}
	var traffic xray.ClientTraffic
	_ = db.Where("email = ?", id).First(&traffic).Error

	status := StatusActive
	if !rec.Enable {
		status = StatusSuspended
	} else if rec.ExpiryTime > 0 && rec.ExpiryTime < time.Now().UnixMilli() {
		status = StatusExpired
	}

	return &SubStatus{
		ID:                id,
		Status:            status,
		TrafficLimitBytes: rec.TotalGB,
		TrafficUsedBytes:  traffic.Up + traffic.Down,
		UploadBytes:       traffic.Up,
		DownloadBytes:     traffic.Down,
		DeviceLimit:       rec.LimitIP,
		ExpiresAt:         time.UnixMilli(rec.ExpiryTime),
		Online:            traffic.LastOnline > time.Now().Add(-2*time.Minute).UnixMilli(),
	}, nil
}

func (p *XrayProvider) SyncUsage(_ context.Context, id string) (*UsageReport, error) {
	db := database.GetDB()
	var traffic xray.ClientTraffic
	if err := db.Where("email = ?", id).First(&traffic).Error; err != nil {
		return nil, err
	}
	return &UsageReport{
		ID:            id,
		UploadBytes:   traffic.Up,
		DownloadBytes: traffic.Down,
		TotalBytes:    traffic.Up + traffic.Down,
		LastSyncAt:    time.Now(),
	}, nil
}

func (p *XrayProvider) IsHealthy(_ context.Context) bool {
	return true
}

func (p *XrayProvider) findRecord(db *gorm.DB, email string) (*model.ClientRecord, error) {
	var rec model.ClientRecord
	if err := db.Where("email = ?", email).First(&rec).Error; err != nil {
		if database.IsNotFound(err) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func (p *XrayProvider) recordToSubscription(rec *model.ClientRecord, traffic *xray.ClientTraffic) *Subscription {
	status := StatusActive
	if !rec.Enable {
		status = StatusSuspended
	} else if rec.ExpiryTime > 0 && rec.ExpiryTime < time.Now().UnixMilli() {
		status = StatusExpired
	}

	sub := &Subscription{
		ID:                rec.Email,
		ProviderName:      p.name,
		Email:             rec.Email,
		Status:            status,
		TrafficLimitBytes: rec.TotalGB,
		DeviceLimit:       rec.LimitIP,
		ExpiresAt:         time.UnixMilli(rec.ExpiryTime),
		CreatedAt:         time.UnixMilli(rec.CreatedAt),
	}
	if traffic != nil {
		sub.TrafficUsedBytes = traffic.Up + traffic.Down
	}
	return sub
}
