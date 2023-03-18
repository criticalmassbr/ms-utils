package utils

import (
	"fmt"
)

type VaultMockService struct {
	config   *VaultConfig
	mockData *map[string]string
}

var VaultMock = VaultMockService{
	mockData: &map[string]string{
		"DATABASE_HOST": "localhost",
		"DATABASE_NAME": "dial_somosdialog_dev",
		"DATABASE_USER": "root",
		"DATABASE_PASS": "",
	},
}

func (v *VaultMockService) NewVaultService(cfg *VaultConfig) IVaultService {
	service := &VaultMockService{
		config: cfg,
	}
	return service
}

func (s *VaultMockService) GetSecret(key VaultSecretKey, clientSlug string) (string, error) {

	value, ok := (*s.mockData)[string(key)]
	if !ok {
		return "", fmt.Errorf("secret value type assertion failed")
	}

	return value, nil
}

func (s *VaultMockService) GetSecrets(clientSlug string, keys []string) (map[string]interface{}, error) {
	filteredSecrets := make(map[string]interface{})
	for _, key := range keys {
		if s.mockData == nil {
			if val, ok := (*s.mockData)[key]; ok {
				filteredSecrets[key] = val
			}
		}
	}

	return filteredSecrets, nil
}

func SetMockData(vaultMock VaultMockService, data *map[string]string) {
	vaultMock.mockData = data
}
