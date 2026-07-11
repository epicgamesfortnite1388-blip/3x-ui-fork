package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/provider"
)

type SubscriptionService struct {
	providerService ProviderService
	registry        *provider.Registry
}

func NewSubscriptionService(registry *provider.Registry) *SubscriptionService {
	return &SubscriptionService{registry: registry}
}

func (s *SubscriptionService) Create(ctx context.Context, providerName string, req provider.CreateSubscriptionRequest) (*provider.Subscription, error) {
	p, ok := s.registry.Get(providerName)
	if !ok {
		return nil, provider.ErrProviderUnavailable
	}
	sub, err := p.CreateSubscription(ctx, req)
	if err != nil {
		return nil, err
	}
	providerRec, err := s.providerService.GetByName(providerName)
	if err != nil {
		return sub, nil
	}
	s.logEvent(sub.ID, providerRec.Id, "created", nil)
	return sub, nil
}

func (s *SubscriptionService) Get(ctx context.Context, providerName, id string) (*provider.Subscription, error) {
	p, ok := s.registry.Get(providerName)
	if !ok {
		return nil, provider.ErrProviderUnavailable
	}
	return p.GetSubscription(ctx, id)
}

func (s *SubscriptionService) Renew(ctx context.Context, providerName, id string, req provider.RenewRequest) error {
	p, ok := s.registry.Get(providerName)
	if !ok {
		return provider.ErrProviderUnavailable
	}
	if err := p.RenewSubscription(ctx, id, req); err != nil {
		return err
	}
	if rec, err := s.providerService.GetByName(providerName); err == nil {
		s.logEvent(id, rec.Id, "renewed", map[string]any{"durationDays": req.DurationDays})
	}
	return nil
}

func (s *SubscriptionService) Suspend(ctx context.Context, providerName, id string) error {
	p, ok := s.registry.Get(providerName)
	if !ok {
		return provider.ErrProviderUnavailable
	}
	if err := p.SuspendSubscription(ctx, id); err != nil {
		return err
	}
	if rec, err := s.providerService.GetByName(providerName); err == nil {
		s.logEvent(id, rec.Id, "suspended", nil)
	}
	return nil
}

func (s *SubscriptionService) Resume(ctx context.Context, providerName, id string) error {
	p, ok := s.registry.Get(providerName)
	if !ok {
		return provider.ErrProviderUnavailable
	}
	if err := p.ResumeSubscription(ctx, id); err != nil {
		return err
	}
	if rec, err := s.providerService.GetByName(providerName); err == nil {
		s.logEvent(id, rec.Id, "resumed", nil)
	}
	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, providerName, id string) error {
	p, ok := s.registry.Get(providerName)
	if !ok {
		return provider.ErrProviderUnavailable
	}
	if err := p.DeleteSubscription(ctx, id); err != nil {
		return err
	}
	if rec, err := s.providerService.GetByName(providerName); err == nil {
		s.logEvent(id, rec.Id, "deleted", nil)
	}
	return nil
}

func (s *SubscriptionService) GetURL(ctx context.Context, providerName, id string) (string, error) {
	p, ok := s.registry.Get(providerName)
	if !ok {
		return "", provider.ErrProviderUnavailable
	}
	return p.GetSubscriptionURL(ctx, id)
}

func (s *SubscriptionService) GetStatus(ctx context.Context, providerName, id string) (*provider.SubStatus, error) {
	p, ok := s.registry.Get(providerName)
	if !ok {
		return nil, provider.ErrProviderUnavailable
	}
	return p.GetStatus(ctx, id)
}

func (s *SubscriptionService) SyncUsage(ctx context.Context, providerName, id string) (*provider.UsageReport, error) {
	p, ok := s.registry.Get(providerName)
	if !ok {
		return nil, provider.ErrProviderUnavailable
	}
	report, err := p.SyncUsage(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec, err := s.providerService.GetByName(providerName); err == nil {
		s.logEvent(id, rec.Id, "usage_synced", map[string]any{"totalBytes": report.TotalBytes})
	}
	return report, nil
}

func (s *SubscriptionService) ListExternal(providerIdFilter int, statusFilter string) ([]*model.ExternalSubscription, error) {
	db := database.GetDB()
	query := db.Model(&model.ExternalSubscription{})
	if providerIdFilter > 0 {
		query = query.Where("provider_id = ?", providerIdFilter)
	}
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	var subs []*model.ExternalSubscription
	err := query.Order("created_at DESC").Find(&subs).Error
	return subs, err
}

func (s *SubscriptionService) CreateExternal(providerRec *model.ProviderRecord, req provider.CreateSubscriptionRequest) (*model.ExternalSubscription, error) {
	sub := &model.ExternalSubscription{
		Id:                uuid.NewString(),
		ProviderId:        providerRec.Id,
		ClientEmail:       req.Email,
		PlanName:          req.PlanName,
		Status:            string(provider.StatusActive),
		TrafficLimitBytes: req.TrafficLimitBytes,
		DeviceLimit:       req.DeviceLimit,
		ExpiresAt:         time.Now().Add(time.Duration(req.DurationDays) * 24 * time.Hour).UnixMilli(),
		Notes:             req.Notes,
	}
	err := database.GetDB().Create(sub).Error
	if err != nil {
		return nil, err
	}
	s.logEvent(sub.Id, providerRec.Id, "created", nil)
	return sub, nil
}

func (s *SubscriptionService) logEvent(subscriptionId string, providerId int, eventType string, details map[string]any) {
	var detailsJSON string
	if details != nil {
		b, _ := json.Marshal(details)
		detailsJSON = string(b)
	}
	event := &model.SubscriptionEvent{
		SubscriptionId: subscriptionId,
		ProviderId:     providerId,
		EventType:      eventType,
		Details:        detailsJSON,
	}
	_ = database.GetDB().Create(event).Error
}
