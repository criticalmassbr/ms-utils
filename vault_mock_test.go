package utils_test

import (
	"sort"
	"testing"

	utils "github.com/criticalmassbr/ms-utils"
	"github.com/stretchr/testify/assert"
)

func TestVaultMockGetSecret(t *testing.T) {
	vaultMockData := utils.VaultMockData{
		"somosdialog": map[string]interface{}{
			"TABS_DB_URL": "value",
			"OTHER_VAR":   "othervalue",
		},
	}

	vault := utils.NewMockVaultService(vaultMockData)

	value, err := vault.GetSecret("somosdialog", "TABS_DB_URL")

	assert.Nil(t, err)
	assert.Equal(t, "value", value)
}

func TestVaultMockGetSecrets(t *testing.T) {
	vaultMockData := utils.VaultMockData{
		"somosdialog": map[string]interface{}{
			"TABS_DB_URL": "value",
			"OTHER_VAR":   "othervalue",
		},
	}

	vault := utils.NewMockVaultService(vaultMockData)

	value, err := vault.GetSecrets("somosdialog", []utils.VaultSecretKey{"TABS_DB_URL", "OTHER_VAR"})

	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{
		"TABS_DB_URL": "value",
		"OTHER_VAR":   "othervalue",
	}, value)
}

func TestVaultMockList(t *testing.T) {
	vaultMockData := utils.VaultMockData{
		"somosdialog": map[string]interface{}{
			"TABS_DB_URL": "value",
			"OTHER_VAR":   "othervalue",
		},
		"client_1": map[string]interface{}{
			"TABS_DB_URL": "value",
			"OTHER_VAR":   "othervalue",
		},
		"client_2": map[string]interface{}{
			"TABS_DB_URL": "value",
			"OTHER_VAR":   "othervalue",
		},
	}

	vault := utils.NewMockVaultService(vaultMockData)

	keys, err := vault.List()

	assert.Nil(t, err)

	sort.Strings(keys)
	assert.Equal(t, []string{
		"client_1",
		"client_2",
		"somosdialog",
	}, keys)
}
