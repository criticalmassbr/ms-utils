package configloader_test

import (
	"fmt"
	"reflect"
	"testing"

	configLoader "github.com/criticalmassbr/ms-utils/config_loader"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	App    testConfigApp `koanf:"app" validate:"required"`
	Source string        `koanf:"source" validate:"required"`
}

type testConfigApp struct {
	ServiceName string `koanf:"service_name" validate:"required"`
	Environment string `koanf:"environment" validate:"required"`
}

func TestConfigLoader(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mockValidator := validator.New()

	expectedQuery := `select e."key", e.value from environment_variables e
	inner join (select "key", max(e.created_at) as created_at from environment_variables e group by e."key") e2
	on e."key" = e2."key" and e.created_at = e2.created_at`

	mockQueryResult := sqlmock.NewRows([]string{"key", "value"}).AddRow("APP.SERVICE_NAME", "config_loader").AddRow("APP.ENVIRONMENT", "testing").AddRow("SOURCE", "db")

	t.Run("Test DefaultInit()", func(t *testing.T) {
		loader := configLoader.New[testConfig](mockValidator)

		context, err := loader.DefaultInit("", "").Context()
		assert.ErrorIs(t, err, configLoader.ErrConfigNotSet)
		assert.Nil(t, context)
	})

	t.Run("Test YAML()", func(t *testing.T) {
		loader := configLoader.New[testConfig](mockValidator)

		expectedConfig := &testConfig{
			App: testConfigApp{
				ServiceName: "config_loader",
				Environment: "test",
			},
			Source: "yaml",
		}

		context, err := loader.DefaultInit("", "").YAML("./test/config_loader_test.yaml").Context()
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(context, expectedConfig), fmt.Sprintf("Expected: %+v\nGot: %+v", expectedConfig, context))
	})

	t.Run("Test Validate()", func(t *testing.T) {
		loader := configLoader.New[testConfig](mockValidator)

		context, err := loader.DefaultInit("", "").YAML("./test/config_loader_test.yaml").Validate().Context()
		assert.NoError(t, err)
		assert.NotNil(t, context)
	})

	t.Run("Test DB()", func(t *testing.T) {
		loader := configLoader.New[testConfig](mockValidator)

		mock.ExpectQuery(expectedQuery).WillReturnRows(mockQueryResult)

		context, err := loader.DefaultInit("", "./test/config_loader_test.yaml").DB(db).Context()
		assert.NoError(t, err)
		assert.NotNil(t, context)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
