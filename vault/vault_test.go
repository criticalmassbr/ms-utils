package vault_test

import (
	"testing"

	"github.com/criticalmassbr/ms-utils/vault"
	"github.com/stretchr/testify/assert"
)

const (
	ENV_1     vault.VaultSecretKey = "ENV_1"
	ENV_2     vault.VaultSecretKey = "ENV_2"
	ENV_3     vault.VaultSecretKey = "ENV_3"
	ENV_4     vault.VaultSecretKey = "ENV_4"
	VAR       vault.VaultSecretKey = "VAR"
	OTHER_VAR vault.VaultSecretKey = "OTHER_VAR"
)

func TestVault(t *testing.T) {
	vaultMockData := vault.VaultMockData{
		"client1": {
			"ENV_1": "val 1",
			"ENV_2": "true",
			"ENV_3": "val 2",
			"ENV_4": "5",
		},
		"client2": {
			"VAR":       "value",
			"OTHER_VAR": "value",
		},
	}

	type SomeSecrets struct {
		Env1 string `koanf:"ENV_1" validate:"required"`
		Env2 bool   `koanf:"ENV_2"`
		Env3 string `koanf:"ENV_3"`
		Env4 int    `koanf:"ENV_4"`
	}

	vaultRepo := vault.NewMockVaultRepository(vaultMockData)
	vaultService := vault.NewVaultService(vaultRepo)

	t.Run("GetSecret should return correct value", func(t *testing.T) {
		val, err := vaultService.GetSecret("client1", ENV_2)
		assert.NoError(t, err)
		assert.Equal(t, "true", val)

		val, err = vaultService.GetSecret("client1", ENV_4)
		assert.NoError(t, err)
		assert.Equal(t, "5", val)

		val, err = vaultService.GetSecret("client2", VAR)
		assert.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("GetSecrets should return correct value", func(t *testing.T) {
		val, err := vaultService.GetSecrets("client1", []vault.VaultSecretKey{ENV_1, ENV_4})
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"ENV_1": "val 1",
			"ENV_4": "5",
		}, val)

		val, err = vaultService.GetSecrets("client2", []vault.VaultSecretKey{OTHER_VAR})
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"OTHER_VAR": "value",
		}, val)
	})

	t.Run("GetSecret should return error when repo returns error", func(t *testing.T) {
		_, err := vaultService.GetSecret("client3", ENV_1)
		assert.Error(t, err)
	})

	t.Run("GetSecrets should return error when repo returns error", func(t *testing.T) {
		_, err := vaultService.GetSecrets("client3", []vault.VaultSecretKey{ENV_1, ENV_2})
		assert.Error(t, err)
	})

	t.Run("secrets cache", func(t *testing.T) {
		repoMock := vault.NewMockVaultRepository(vaultMockData)
		vaultService := vault.NewVaultService(repoMock)

		vaultService.GetSecret("client1", ENV_1)
		assert.Equal(t, 1, repoMock.NumberOfCalls("client1"))

		vaultService.GetSecret("client1", ENV_2)
		assert.Equal(t, 1, repoMock.NumberOfCalls("client1"))

		vaultService.GetSecrets("client1", []vault.VaultSecretKey{})
		assert.Equal(t, 1, repoMock.NumberOfCalls("client1"))

		vaultService.GetSecrets("client2", []vault.VaultSecretKey{})
		assert.Equal(t, 1, repoMock.NumberOfCalls("client2"))
	})

	t.Run("ReadSecrets should parse values correctly", func(t *testing.T) {
		secrets := SomeSecrets{}

		err := vaultService.ReadSecrets("client1", &secrets)
		assert.NoError(t, err)

		assert.Equal(t, SomeSecrets{
			Env1: "val 1",
			Env2: true,
			Env3: "val 2",
			Env4: 5,
		}, secrets)
	})

	t.Run("ReadSecrets should ignore missing required values", func(t *testing.T) {
		secrets := SomeSecrets{}

		vaultService := vault.NewVaultService(vault.NewMockVaultRepository(vault.VaultMockData{
			"client1": {
				"ENV_1": "value",
				"ENV_2": "",
			},
		}))

		err := vaultService.ReadSecrets("client1", &secrets)
		assert.NoError(t, err)

		assert.Equal(t, SomeSecrets{
			Env1: "value",
			Env2: false,
			Env3: "",
			Env4: 0,
		}, secrets)
	})
}
