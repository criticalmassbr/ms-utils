package vault

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/criticalmassbr/ms-utils/typed_sync_map"
	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
)

type VaultMockConfig struct {
	Enabled  bool   `koanf:"enabled"`
	JsonFile string `koang:"json_file"`
}

type VaultConfig struct {
	RoleId    string `koanf:"role_id" validate:"required"`
	SecretId  string `koanf:"secret_id" validate:"required"`
	Url       string `koanf:"url" validate:"required"`
	MountPath string `koanf:"mount_path" validate:"required"`
	Cert      string `koanf:"cert" validate:"required"`
	Mock      VaultMockConfig
}

type IVaultService interface {
	GetSecret(clientSlug string, key VaultSecretKey) (interface{}, error)
	GetSecretAsString(clientSlug string, key VaultSecretKey) (string, error)
	GetSecrets(clientSlug string, keys []VaultSecretKey) (map[string]interface{}, error)
	ReadSecrets(clientSlug string, dest interface{}) error
	List() ([]string, error)
}

type VaultRepository interface {
	GetSecrets(clientSlug string) (map[string]interface{}, error)
	List() ([]string, error)
}

type VaultService struct {
	repo     VaultRepository
	cache    typed_sync_map.TypedSyncMap[string, map[string]interface{}]
	validate *validator.Validate
}

type VaultSecretKey string
type ClientSlug string

func NewVaultService(vaultRepo VaultRepository) IVaultService {
	return &VaultService{
		repo:     vaultRepo,
		validate: validator.New(),
	}
}

func NewVaultServiceFromConfig(cfg VaultConfig) (IVaultService, error) {
	if cfg.Mock.Enabled {
		if cfg.Mock.JsonFile == "" {
			return nil, fmt.Errorf("mocked json file name not provided")
		}

		content, err := os.ReadFile(cfg.Mock.JsonFile)
		if err != nil {
			return nil, err
		}

		mockData := map[string]map[string]interface{}{}
		err = json.Unmarshal(content, &mockData)
		if err != nil {
			return nil, err
		}

		fmt.Println(mockData)

		return NewVaultService(NewMockVaultRepository(mockData)), nil
	}

	vaultRepo, err := NewVaultRepository(&cfg)
	if err != nil {
		return nil, err
	}

	return NewVaultService(vaultRepo), nil
}

func (s *VaultService) getClientSecrets(clientSlug string) (map[string]interface{}, error) {
	cachedSecrets, ok := s.cache.Load(clientSlug)
	if ok {
		return cachedSecrets, nil
	}

	secrets, err := s.repo.GetSecrets(clientSlug)
	if err != nil {
		return secrets, err
	}

	s.cache.Store(clientSlug, secrets)
	return secrets, nil
}

func (s *VaultService) GetSecret(clientSlug string, key VaultSecretKey) (interface{}, error) {
	secrets, err := s.getClientSecrets(clientSlug)
	if err != nil {
		return "", err
	}

	value, _ := secrets[string(key)]
	return value, nil
}

func (s *VaultService) GetSecretAsString(clientSlug string, key VaultSecretKey) (string, error) {
	secrets, err := s.getClientSecrets(clientSlug)
	if err != nil {
		return "", err
	}

	value, ok := secrets[string(key)]
	if !ok {
		return "", nil
	}

	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value is not string")
	}

	return stringValue, nil
}

func (s *VaultService) GetSecrets(clientSlug string, keys []VaultSecretKey) (map[string]interface{}, error) {
	secrets, err := s.getClientSecrets(clientSlug)
	if err != nil {
		return nil, err
	}

	filteredSecrets := make(map[string]interface{})
	for _, key := range keys {
		if val, ok := secrets[string(key)]; ok {
			filteredSecrets[string(key)] = val
		}
	}

	return filteredSecrets, nil
}

func (s *VaultService) ReadSecrets(clientSlug string, dest interface{}) error {
	secrets, err := s.getClientSecrets(clientSlug)
	if err != nil {
		return err
	}

	k := koanf.New(".")
	err = k.Load(confmap.Provider(secrets, ""), nil)
	if err != nil {
		return err
	}
	err = k.Unmarshal("", dest)
	if err != nil {
		return err
	}

	err = s.validate.Struct(dest)
	if err != nil {
		return err
	}

	return nil
}

func (s *VaultService) List() ([]string, error) {
	return s.repo.List()
}
