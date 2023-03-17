package utils

import (
	"fmt"
)

type VaultMockConfig struct {
	RoleId    string
	SecretId  string
	Url       string
	MountPath string
	Cert      string
}

type VaultMockService struct {
	config *VaultMockConfig
}

type VaultMockSecretKey string

var VaultMock = VaultMockService{}

var VaultMockData = map[string]string{
	"DATABASE_HOST": "localhost",
	"DATABASE_NAME": "dial_somosdialog_dev",
	"DATABASE_USER": "root",
	"DATABASE_PASS": "",
}

func (v *VaultMockService) NewVaultService(cfg *VaultMockConfig) VaultMockService {
	service := VaultMockService{
		config: cfg,
	}
	return service
}

func (s *VaultMockService) GetSecret(key VaultMockSecretKey, clientSlug string) (string, error) {

	value, ok := VaultMockData[string(key)]
	if !ok {
		return "", fmt.Errorf("secret value type assertion failed")
	}

	return value, nil
}

func (s *VaultMockService) GetSecrets(clientSlug string, keys []string) (map[string]interface{}, error) {
	filteredSecrets := make(map[string]interface{})
	for _, key := range keys {
		if val, ok := VaultMockData[key]; ok {
			filteredSecrets[key] = val
		}
	}

	return filteredSecrets, nil
}

func (s *VaultMockService) UpdateMockSecrets(data map[string]string) {
	VaultMockData = data
}
