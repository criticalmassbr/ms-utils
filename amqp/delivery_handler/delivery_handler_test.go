package deliveryhandler_test

import (
	"encoding/json"
	"errors"
	"testing"

	deliveryhandler "github.com/criticalmassbr/ms-utils/amqp/delivery_handler"
	"github.com/go-playground/validator/v10"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

type TestInput struct {
	Name string `json:"name" validate:"required"`
}

type TestOutput struct {
	Message string `json:"message"`
}

func TestDeliveryHandler(t *testing.T) {
	mockDelivery := &amqp.Delivery{
		Headers: amqp.Table{
			string(deliveryhandler.RAH_CLIENT_SLUG): "mock-client",
		},
		Body: []byte(`{"name": "John"}`),
	}

	mockHandler := func(d deliveryhandler.DeliveryContext[TestInput]) (TestOutput, error) {
		return TestOutput{
			Message: "Hello, " + d.Body.Name,
		}, nil
	}

	t.Run("Test ClientSlug()", func(t *testing.T) {
		testHandler := deliveryhandler.New[TestInput, TestOutput](mockDelivery)

		testHandler.ClientSlug()
		context, err := testHandler.Context()
		assert.NoError(t, err)
		assert.Equal(t, "mock-client", context.ClientSlug)
	})

	t.Run("Test UnmarshalBody()", func(t *testing.T) {
		testHandler := deliveryhandler.New[TestInput, TestOutput](mockDelivery)

		testHandler.UnmarshalBody()
		context, err := testHandler.Context()
		assert.NoError(t, err)
		assert.Equal(t, "John", context.Body.Name)
	})

	t.Run("Test ValidateBody()", func(t *testing.T) {
		testHandler := deliveryhandler.New[TestInput, TestOutput](mockDelivery)

		validate := validator.New()
		_, err := testHandler.UnmarshalBody().ValidateBody(validate).Context()
		assert.NoError(t, err)
	})

	t.Run("Test Handle()", func(t *testing.T) {
		testHandler := deliveryhandler.New[TestInput, TestOutput](mockDelivery)

		_, err := testHandler.UnmarshalBody().Handle(mockHandler).Context()
		assert.NoError(t, err)
	})

	t.Run("Test Response()", func(t *testing.T) {
		testHandler := deliveryhandler.New[TestInput, TestOutput](mockDelivery)

		response, err := testHandler.UnmarshalBody().Handle(mockHandler).Response()
		assert.NoError(t, err)
		expectedResponse, _ := json.Marshal(TestOutput{Message: "Hello, John"})
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("Test MappedResponse()", func(t *testing.T) {
		testHandler := deliveryhandler.New[TestInput, TestOutput](mockDelivery)

		mapper := func(output TestOutput) interface{} {
			return map[string]string{
				"message": output.Message,
			}
		}
		mappedResponse, err := testHandler.UnmarshalBody().Handle(mockHandler).MappedResponse(mapper)
		assert.NoError(t, err)
		expectedMappedResponse, _ := json.Marshal(map[string]string{
			"message": "Hello, John",
		})
		assert.Equal(t, expectedMappedResponse, mappedResponse)
	})

	t.Run("ErrClientSlugRequired", func(t *testing.T) {
		mockDeliveryWithouClientSlug := &amqp.Delivery{
			Headers: amqp.Table{},
			Body:    []byte(`{"name": "John"}`),
		}

		testHandler := deliveryhandler.New[TestInput, TestOutput](mockDeliveryWithouClientSlug)

		_, err := testHandler.ClientSlug().Context()
		assert.ErrorIs(t, err, deliveryhandler.ErrClientSlugRequired)
	})

	t.Run("ErrClientSlugRequired", func(t *testing.T) {
		mockDeliveryWithouClientSlug := &amqp.Delivery{
			Body: []byte(`{"name": "John"}`),
		}

		testHandler := deliveryhandler.New[TestInput, TestOutput](mockDeliveryWithouClientSlug)

		_, err := testHandler.ClientSlug().Context()
		assert.ErrorIs(t, err, deliveryhandler.ErrClientSlugRequired)
	})
}

func TestDeliveryHandlerWithoutResponse(t *testing.T) {
	mockDelivery := &amqp.Delivery{
		Headers: amqp.Table{
			string(deliveryhandler.RAH_CLIENT_SLUG): "mock-client",
		},
		Body: []byte(`{"name": "John"}`),
	}

	mockHandlerWithResult := func(d deliveryhandler.DeliveryContext[TestInput]) (deliveryhandler.Undefined, error) {
		return errors.New("improper usage of handler without response"), nil
	}

	mockHandler := func(d deliveryhandler.DeliveryContext[TestInput]) error {
		return nil
	}

	t.Run("Test ClientSlug()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutResult[TestInput](mockDelivery)

		testHandler.ClientSlug()
		context, err := testHandler.Context()
		assert.NoError(t, err)
		assert.Equal(t, "mock-client", context.ClientSlug)
	})

	t.Run("Test UnmarshalBody()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutResult[TestInput](mockDelivery)

		testHandler.UnmarshalBody()
		context, err := testHandler.Context()
		assert.NoError(t, err)
		assert.Equal(t, "John", context.Body.Name)
	})

	t.Run("Test ValidateBody()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutResult[TestInput](mockDelivery)

		validate := validator.New()
		_, err := testHandler.UnmarshalBody().ValidateBody(validate).Context()
		assert.NoError(t, err)
	})

	t.Run("Test Handle()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutResult[TestInput](mockDelivery)

		_, err := testHandler.UnmarshalBody().Handle(mockHandlerWithResult).Context()
		assert.NoError(t, err)
	})

	t.Run("Test HandleOnlyError()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutResult[TestInput](mockDelivery)

		_, err := testHandler.UnmarshalBody().HandleOnlyError(mockHandler).Context()
		assert.NoError(t, err)
	})

	t.Run("Test Response()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutResult[TestInput](mockDelivery)

		response, err := testHandler.UnmarshalBody().HandleOnlyError(mockHandler).Response()
		assert.NoError(t, err)
		assert.Nil(t, response)
	})

	t.Run("Test MappedResponse()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutResult[TestInput](mockDelivery)

		mapper := func(output deliveryhandler.Undefined) interface{} {
			return map[string]string{
				"message": output.Error(),
			}
		}
		mappedResponse, err := testHandler.UnmarshalBody().HandleOnlyError(mockHandler).MappedResponse(mapper)
		assert.NoError(t, err)
		assert.Nil(t, mappedResponse)
	})
}

func TestDeliveryHandlerWithoutInput(t *testing.T) {
	mockDelivery := &amqp.Delivery{
		Headers: amqp.Table{
			string(deliveryhandler.RAH_CLIENT_SLUG): "mock-client",
		},
	}

	mockHandler := func(d deliveryhandler.DeliveryContext[deliveryhandler.Undefined]) (string, error) {
		return d.ClientSlug, nil
	}

	mockHandlerOnlyError := func(d deliveryhandler.DeliveryContext[deliveryhandler.Undefined]) error {
		return nil
	}

	t.Run("Test ClientSlug()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutInput[deliveryhandler.Undefined](mockDelivery)

		testHandler.ClientSlug()
		context, err := testHandler.Context()
		assert.NoError(t, err)
		assert.Equal(t, "mock-client", context.ClientSlug)
	})

	t.Run("Test UnmarshalBody()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutInput[deliveryhandler.Undefined](mockDelivery)

		testHandler.UnmarshalBody()
		context, err := testHandler.Context()
		assert.NoError(t, err)
		assert.Equal(t, deliveryhandler.ErrUndefinedField, context.Body)
	})

	t.Run("Test ValidateBody()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutInput[deliveryhandler.Undefined](mockDelivery)

		validate := validator.New()
		_, err := testHandler.UnmarshalBody().ValidateBody(validate).Context()
		assert.NoError(t, err)
	})

	t.Run("Test Handle()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutInput[string](mockDelivery)

		_, err := testHandler.UnmarshalBody().Handle(mockHandler).Context()
		assert.NoError(t, err)
	})

	t.Run("Test HandleOnlyError()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutInput[deliveryhandler.Undefined](mockDelivery)

		_, err := testHandler.UnmarshalBody().HandleOnlyError(mockHandlerOnlyError).Context()
		assert.NoError(t, err)
	})

	t.Run("Test Response()", func(t *testing.T) {
		testHandler := deliveryhandler.NewWithoutInput[string](mockDelivery)

		response, err := testHandler.ClientSlug().Handle(mockHandler).Response()
		assert.NoError(t, err)
		assert.Equal(t, []byte(`"mock-client"`), response)
	})

}
