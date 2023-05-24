package utils_test

import (
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

	value, err := vault.GetSecret("TABS_DB_URL", "somosdialog")

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

	value, err := vault.GetSecrets("somosdialog", []string{"TABS_DB_URL", "OTHER_VAR"})

	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{
		"TABS_DB_URL": "value",
		"OTHER_VAR":   "othervalue",
	}, value)
}
