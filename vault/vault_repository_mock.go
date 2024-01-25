package vault

import (
	"encoding/json"
	"fmt"
	"os"
)

type VaultMockData map[string]map[string]interface{}

type vaultMockRepository struct {
	mockData      VaultMockData
	numberOfCalls map[string]int
}

func NewMockVaultRepository(mockData VaultMockData) *vaultMockRepository {
	service := &vaultMockRepository{
		mockData:      mockData,
		numberOfCalls: map[string]int{},
	}
	return service
}

func NewMockVaultRepositoryFromJsonFile(fileName string) (*vaultMockRepository, error) {
	if fileName == "" {
		return nil, fmt.Errorf("mocked json file name not provided")
	}

	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	mockData := map[string]map[string]interface{}{}
	err = json.Unmarshal(content, &mockData)
	if err != nil {
		return nil, err
	}

	return &vaultMockRepository{
		mockData:      mockData,
		numberOfCalls: map[string]int{},
	}, nil
}

func (s *vaultMockRepository) GetSecrets(clientSlug string) (map[string]interface{}, error) {
	n, _ := s.numberOfCalls[clientSlug]
	s.numberOfCalls[clientSlug] = n + 1

	clientSecrets, ok := s.mockData[clientSlug]
	if !ok {
		return nil, fmt.Errorf("No secrets for client %s", clientSlug)
	}

	return clientSecrets, nil
}

func (s *vaultMockRepository) NumberOfCalls(clientSlug string) int {
	n, _ := s.numberOfCalls[clientSlug]
	return n
}

func (s *vaultMockRepository) List() ([]string, error) {
	keys := make([]string, 0)

	for key := range s.mockData {
		keys = append(keys, key)
	}

	return keys, nil
}
