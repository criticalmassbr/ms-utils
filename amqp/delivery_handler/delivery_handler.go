package deliveryhandler

import (
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
	amqp "github.com/rabbitmq/amqp091-go"
)

type DeliveryHandler[I any, O any] interface {
	ClientSlug() DeliveryHandler[I, O]
	UnmarshalBody() DeliveryHandler[I, O]
	ValidateBody(validate *validator.Validate) DeliveryHandler[I, O]
	Handle(handler func(d DeliveryContext[I]) (O, error)) DeliveryHandler[I, O]
	HandleOnlyError(handler func(d DeliveryContext[I]) error) DeliveryHandler[I, O]
	MappedResponse(mapper func(output O) interface{}) ([]byte, error)
	Response() ([]byte, error)
	Context() (DeliveryContext[I], error)
}

type delivery[I any, O any] struct {
	delivery     *amqp.Delivery
	err          error
	handleResult O
	context      DeliveryContext[I]
	options      deliveryOptions
}

type deliveryOptions struct {
	noResponse bool
}

type DeliveryContext[I any] struct {
	ClientSlug string
	Body       I
}

type ReservedAmqpHeaderKeys string

const (
	RAH_CLIENT_SLUG ReservedAmqpHeaderKeys = "ClientSlug"
)

var (
	ErrClientSlugRequired = errors.New("client slug is required")
	ErrNoResponse         = errors.New("no response")
)

type NoResponse error

func New[I any, O any](d *amqp.Delivery) DeliveryHandler[I, O] {
	return &delivery[I, O]{
		delivery: d,
	}
}

func NewWithoutResult[I any](d *amqp.Delivery) DeliveryHandler[I, NoResponse] {
	return &delivery[I, NoResponse]{
		delivery:     d,
		handleResult: NoResponse(ErrNoResponse),
		options: deliveryOptions{
			noResponse: true,
		},
	}
}

func (d *delivery[I, O]) ClientSlug() DeliveryHandler[I, O] {
	if d.err != nil {
		return d
	}

	clientSlug, err := getClientSlug(d.delivery)
	if err != nil {
		d.err = err
		return d
	}

	d.context.ClientSlug = clientSlug
	return d
}

func (d *delivery[I, O]) UnmarshalBody() DeliveryHandler[I, O] {
	if d.err != nil {
		return d
	}

	err := json.Unmarshal(d.delivery.Body, &d.context.Body)
	if err != nil {
		d.err = err
		return d
	}

	return d
}

func (d *delivery[I, O]) Handle(handler func(d DeliveryContext[I]) (O, error)) DeliveryHandler[I, O] {
	if d.err != nil {
		return d
	}

	d.handleResult, d.err = handler(d.context)
	return d
}

func (d *delivery[I, O]) HandleOnlyError(handler func(d DeliveryContext[I]) error) DeliveryHandler[I, O] {
	if d.err != nil {
		return d
	}

	d.err = handler(d.context)
	return d
}

func (d *delivery[I, O]) MappedResponse(mapper func(output O) interface{}) ([]byte, error) {
	if d.err != nil {
		return nil, d.err
	}

	if !d.options.noResponse {
		return json.Marshal(mapper(d.handleResult))
	}

	return nil, nil
}

func (d *delivery[I, O]) Response() ([]byte, error) {
	if d.err != nil {
		return nil, d.err
	}

	if !d.options.noResponse {
		return json.Marshal(d.handleResult)
	}

	return nil, nil
}

func (d *delivery[I, O]) Context() (DeliveryContext[I], error) {
	if d.err != nil {
		return DeliveryContext[I]{}, d.err
	}

	return d.context, nil
}

func (d *delivery[I, O]) ValidateBody(validate *validator.Validate) DeliveryHandler[I, O] {
	if d.err != nil {
		return d
	}

	err := validate.Struct(d.context.Body)
	if err != nil {
		d.err = err
		return d
	}

	return d
}

func getClientSlug(d *amqp.Delivery) (string, error) {
	if d == nil || d.Headers == nil {
		return "", ErrClientSlugRequired
	}

	clientSlug, clientSlugIsString := d.Headers[string(RAH_CLIENT_SLUG)].(string)
	if !clientSlugIsString || clientSlug == "" {
		return "", ErrClientSlugRequired
	}

	return clientSlug, nil
}
