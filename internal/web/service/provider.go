package service

import (
	"encoding/json"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

type ProviderService struct{}

func (s *ProviderService) GetAll() ([]*model.ProviderRecord, error) {
	var providers []*model.ProviderRecord
	err := database.GetDB().Find(&providers).Error
	return providers, err
}

func (s *ProviderService) GetById(id int) (*model.ProviderRecord, error) {
	var provider model.ProviderRecord
	err := database.GetDB().First(&provider, id).Error
	return &provider, err
}

func (s *ProviderService) GetByName(name string) (*model.ProviderRecord, error) {
	var provider model.ProviderRecord
	err := database.GetDB().Where("name = ?", name).First(&provider).Error
	return &provider, err
}

func (s *ProviderService) Create(provider *model.ProviderRecord) error {
	return database.GetDB().Create(provider).Error
}

func (s *ProviderService) Update(provider *model.ProviderRecord) error {
	return database.GetDB().Save(provider).Error
}

func (s *ProviderService) Delete(id int) error {
	return database.GetDB().Delete(&model.ProviderRecord{}, id).Error
}

func (s *ProviderService) GetConfig(provider *model.ProviderRecord) (map[string]any, error) {
	if provider.Config == "" || provider.Config == "{}" {
		return map[string]any{}, nil
	}
	var config map[string]any
	err := json.Unmarshal([]byte(provider.Config), &config)
	return config, err
}
