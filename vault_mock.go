package utils

import (
	"fmt"
)

type VaultMockData map[string]map[string]interface{}

type VaultMockService struct {
	mockData VaultMockData
}

var VaultMock = VaultMockService{
	mockData: make(VaultMockData),
}

func NewMockVaultService(mockData VaultMockData) IVaultService {
	service := &VaultMockService{
		mockData: mockData,
	}
	return service
}

func (s *VaultMockService) GetSecret(key VaultSecretKey, clientSlug string) (interface{}, error) {
	clientSecrets, ok := s.mockData[string(clientSlug)]
	if !ok {
		return "", fmt.Errorf("client slug does was not set")
	}

	value, ok := clientSecrets[string(key)]
	if !ok {
		return "", fmt.Errorf("secret is not set on client")
	}

	return value, nil
}

func (s *VaultMockService) GetSecrets(keys []string, clientSlug string) (map[string]interface{}, error) {
	filteredSecrets := make(map[string]interface{})
	clientSecrets := s.mockData[string(clientSlug)]
	for _, key := range keys {
		if s.mockData != nil {
			if val, ok := clientSecrets[string(key)]; ok {
				filteredSecrets[key] = val
			}
		}
	}

	return filteredSecrets, nil
}
