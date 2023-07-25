package utils

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Filter[T Excludable] struct {
	Relation   Relation    `json:"relation"`
	Conditions []Condition `json:"conditions"`
	Exclude    T           `json:"exclude"`
}

type Excludable interface{ ExcludableV1 | ExcludableBurst }

type ExcludableV1 struct {
	Users []int `json:"users"`
}

type ExcludableBurst struct {
	Users []uuid.UUID `json:"users"`
}

type FieldFilter[T Excludable] struct {
	FieldName  FieldCount `json:"fieldName"`
	EmployeeID int        `json:"employeeId"`
	Filter     Filter[T]  `json:"filter"`
}

type Condition struct {
	FieldName FieldName   `json:"fieldName"`
	Operator  Operator    `json:"operator"`
	Value     interface{} `json:"value"` // RFCDate | [2]RFCDate | int | []int | string | []string | uuid.UUID | []uuid.UUID
}

func (c *Condition) UnmarshalJSON(data []byte) error {
	type Alias Condition
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	switch v := c.Value.(type) {
	case []interface{}:
		if len(v) == 0 {
			return errors.New("value is required")
		}
		switch v[0].(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			c.Value = castToIntSlice(v)
		case map[string]interface{}:
			d, err := castToRFCDateTuple(v)
			if err != nil {
				return err
			}
			c.Value = d
		case string:
			if ids, err := castToUUIDSlice(c.Value); err == nil {
				c.Value = ids
			} else {
				c.Value = castToStringSlice(v)
			}
		default:
			return errors.New("value is invalid")
		}
	case interface{}:
		switch v := v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			c.Value = castToint(v)
		case map[string]interface{}:
			d, err := castToRFCDate(v)
			if err != nil {
				return err
			}
			c.Value = d
		case string:
			if id, err := castToUUID(c.Value); err == nil {
				c.Value = id
			} else {
				c.Value = string(v)
			}
		default:
			return errors.New("value is invalid")
		}
	}
	return nil
}

func castToIntSlice(v []interface{}) []int {
	var result []int
	for _, vv := range v {
		switch v := vv.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			result = append(result, castToint(v))
		}
	}
	return result
}

func castToRFCDateTuple(v []interface{}) ([2]RFCDate, error) {
	var result [2]RFCDate
	for i, vv := range v {
		switch v := vv.(type) {
		case map[string]interface{}:
			d, err := castToRFCDate(v)
			if err != nil {
				return result, err
			}
			result[i] = d
		}
	}
	return result, nil
}

func castToStringSlice(v []interface{}) []string {
	var result []string
	for _, vv := range v {
		switch v := vv.(type) {
		case string:
			result = append(result, string(v))
		}
	}
	return result
}

func castToUUID(v interface{}) (uuid.UUID, error) {
	switch v := (v).(type) {
	case string:
		u, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil, err
		}
		return u, nil
	}
	return uuid.Nil, nil
}

func castToUUIDSlice(v interface{}) ([]uuid.UUID, error) {
	var result []uuid.UUID
	switch v := v.(type) {
	case []interface{}:
		for _, vv := range v {
			switch v := vv.(type) {
			case string:
				u, err := uuid.Parse(v)
				if err != nil {
					return nil, err
				}
				result = append(result, u)
			}
		}
	}
	return result, nil
}

func castToint(v interface{}) int {
	switch v := v.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

func castToRFCDate(v map[string]interface{}) (RFCDate, error) {
	var r RFCDate

	date, ok := v["date"].(string)
	if !ok {
		return RFCDate{}, errors.New("date is required")
	}
	format, ok := v["format"].([]interface{})
	if !ok {
		return RFCDate{}, errors.New("format is required")
	}

	d, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return RFCDate{}, err
	}
	r.Date = d
	for _, f := range format {
		f, ok := f.(string)
		if !ok {
			return RFCDate{}, errors.New("format must be string")
		}
		r.Format = append(r.Format, RFCDateFormat(f))
	}
	return r, nil
}

type FieldName string

const (
	FieldNameCreatedAt FieldName = "createdAt"
	FieldNameUpdatedAt FieldName = "updatedAt"
	FieldNameBirthday  FieldName = "birthday"
	FieldNameHireDate  FieldName = "hireDate"

	FieldNameDepartmentId FieldName = "department"
	FieldNameJobId        FieldName = "job"
	FieldNameCompanySite  FieldName = "companySite"
	FieldNameUnit         FieldName = "unit"
	FieldNameCity         FieldName = "city"
	FieldNameState        FieldName = "state"
	FieldNameHierarchy    FieldName = "hierarchy"
	FieldNameGroup        FieldName = "group"

	FieldNameName  FieldName = "name"
	FieldNameEmail FieldName = "email"
	FieldNamePhone FieldName = "phone"

	FieldNameRelationalCustom1 FieldName = "custom1"
	FieldNameRelationalCustom2 FieldName = "custom2"
	FieldNameRelationalCustom3 FieldName = "custom3"
	FieldNameRelationalCustom4 FieldName = "custom4"
	FieldNameRelationalCustom5 FieldName = "custom5"
	FieldNameRelationalCustom6 FieldName = "custom6"
	FieldNameRelationalCustom7 FieldName = "custom7"
	FieldNameRelationalCustom8 FieldName = "custom8"
	FieldNameRelationalCustom9 FieldName = "custom9"
)

type FieldCount string

const (
	FieldCountDepartment  FieldCount = "department"
	FieldCountJob         FieldCount = "job"
	FieldCountCompanySite FieldCount = "companySite"
	FieldCountHierarchy   FieldCount = "hierarchy"
	FieldCountCity        FieldCount = "city"
	FieldCountState       FieldCount = "state"
	FieldCountGroup       FieldCount = "group"
	FieldCountLocation    FieldCount = "location"

	FieldCountRelationalCustom1 FieldCount = "custom1"
	FieldCountRelationalCustom2 FieldCount = "custom2"
	FieldCountRelationalCustom3 FieldCount = "custom3"
	FieldCountRelationalCustom4 FieldCount = "custom4"
	FieldCountRelationalCustom5 FieldCount = "custom5"
	FieldCountRelationalCustom6 FieldCount = "custom6"
	FieldCountRelationalCustom7 FieldCount = "custom7"
	FieldCountRelationalCustom8 FieldCount = "custom8"
	FieldCountRelationalCustom9 FieldCount = "custom9"
)

type RFCDate struct {
	Date   time.Time       `json:"date"`
	Format []RFCDateFormat `json:"format"`
}

type RFCDateFormat string

const (
	RFCDateFormatDay   RFCDateFormat = "day"
	RFCDateFormatMonth RFCDateFormat = "month"
	RFCDateFormatYear  RFCDateFormat = "year"
	RFCDateFormatTime  RFCDateFormat = "time"
)

type Relation string

const (
	RelationAnd Relation = "and"
	RelationOr  Relation = "or"
)

type Operator string

const (
	OperatorEq      Operator = "eq"
	OperatorNotEq   Operator = "ne"
	OperatorIn      Operator = "in"
	OperatorNotIn   Operator = "notIn"
	OperatorBetween Operator = "between"
	OperatorGt      Operator = "gt"
	OperatorLt      Operator = "lt"
)

type ValidateCondition struct {
	Fields         []FieldName
	ValidOperators []ValidateOperator
}

type ValidateOperator struct {
	Operators           []Operator
	ValueTypeValidators []func(interface{}) bool
}

var (
	ValidConditionsV1 = []ValidateCondition{
		{
			Fields: []FieldName{FieldNameCreatedAt, FieldNameUpdatedAt, FieldNameBirthday, FieldNameHireDate},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorBetween},
					ValueTypeValidators: []func(interface{}) bool{isRFCDateTuple},
				},
				{
					Operators:           []Operator{OperatorEq, OperatorGt, OperatorLt},
					ValueTypeValidators: []func(interface{}) bool{isRFCDate},
				},
			},
		},
		{
			Fields: []FieldName{FieldNameDepartmentId, FieldNameJobId, FieldNameCompanySite, FieldNameCity, FieldNameState, FieldNameUnit, FieldNameHierarchy, FieldNameGroup, FieldNameRelationalCustom1, FieldNameRelationalCustom2, FieldNameRelationalCustom3, FieldNameRelationalCustom4, FieldNameRelationalCustom5, FieldNameRelationalCustom6, FieldNameRelationalCustom7, FieldNameRelationalCustom8, FieldNameRelationalCustom9},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorEq, OperatorNotEq, OperatorIn, OperatorNotIn},
					ValueTypeValidators: []func(interface{}) bool{isInt, isIntSlice},
				},
			},
		},
		{
			Fields: []FieldName{FieldNameName, FieldNamePhone},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorEq, OperatorNotEq},
					ValueTypeValidators: []func(interface{}) bool{isString},
				},
			},
		},
		{
			Fields: []FieldName{FieldNameEmail},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorEq, OperatorNotEq, OperatorIn, OperatorNotIn},
					ValueTypeValidators: []func(interface{}) bool{isString, isStringSlice},
				},
			},
		},
	}
	ValidConditionsBurst = []ValidateCondition{
		{
			Fields: []FieldName{FieldNameCreatedAt, FieldNameUpdatedAt, FieldNameBirthday, FieldNameHireDate},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorBetween},
					ValueTypeValidators: []func(interface{}) bool{isRFCDateTuple},
				},
				{
					Operators:           []Operator{OperatorEq, OperatorGt, OperatorLt},
					ValueTypeValidators: []func(interface{}) bool{isRFCDate},
				},
			},
		},
		{
			Fields: []FieldName{FieldNameDepartmentId, FieldNameJobId, FieldNameCompanySite, FieldNameHierarchy, FieldNameGroup},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorEq, OperatorNotEq, OperatorIn, OperatorNotIn},
					ValueTypeValidators: []func(interface{}) bool{isUUID, isUUIDSlice},
				},
			},
		},
		{
			Fields: []FieldName{FieldNameName, FieldNamePhone},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorEq, OperatorNotEq},
					ValueTypeValidators: []func(interface{}) bool{isString},
				},
			},
		},
		{
			Fields: []FieldName{FieldNameEmail},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorEq, OperatorNotEq, OperatorIn, OperatorNotIn},
					ValueTypeValidators: []func(interface{}) bool{isString, isStringSlice},
				},
			},
		},
	}

	ValidCountFields = []FieldCount{
		FieldCountDepartment,
		FieldCountJob,
		FieldCountCompanySite,
		FieldCountHierarchy,
		FieldCountGroup,
		FieldCountLocation,
		FieldCountRelationalCustom1,
		FieldCountRelationalCustom2,
		FieldCountRelationalCustom3,
		FieldCountRelationalCustom4,
		FieldCountRelationalCustom5,
		FieldCountRelationalCustom6,
		FieldCountRelationalCustom7,
		FieldCountRelationalCustom8,
		FieldCountRelationalCustom9,
	}
)

func (filter *FieldFilter[T]) Validate(validCountFields []FieldCount, validConditions []ValidateCondition) bool {
	return Contains(validCountFields, filter.FieldName) && isEveryFilterConditionValid(&filter.Filter, validConditions)
}

func (filter *Filter[T]) Validate(validConditions []ValidateCondition) bool {
	return isEveryFilterConditionValid(filter, validConditions)
}

func isEveryFilterConditionValid[T Excludable](filter *Filter[T], validConditions []ValidateCondition) bool {
	for _, condition := range filter.Conditions {
		if !isConditionValid(condition, validConditions) {
			return false
		}
	}
	return true
}

func isConditionValid(condition Condition, validConditions []ValidateCondition) bool {
	for _, validCondition := range validConditions {
		if isConditionValidForField(condition, validCondition) {
			return true
		}
	}
	return false
}

func isConditionValidForField(condition Condition, validCondition ValidateCondition) bool {
	for _, field := range validCondition.Fields {
		if condition.FieldName == field {
			return isConditionValidForOperator(condition, validCondition)
		}
	}
	return false
}

func isConditionValidForOperator(condition Condition, validCondition ValidateCondition) bool {
	for _, validOperator := range validCondition.ValidOperators {
		if isConditionValidForOperatorType(condition, validOperator) {
			return true
		}
	}
	return false
}

func isConditionValidForOperatorType(condition Condition, validOperator ValidateOperator) bool {
	for _, operator := range validOperator.Operators {
		if condition.Operator == operator {
			return isConditionValidForValueType(condition, validOperator)
		}
	}
	return false
}

func isConditionValidForValueType(condition Condition, validOperator ValidateOperator) bool {
	for _, valueTypeValidator := range validOperator.ValueTypeValidators {
		if valueTypeValidator(condition.Value) {
			return true
		}
	}
	return false
}

func isRFCDateTuple(t interface{}) bool {
	tuple, ok := t.([2]RFCDate)
	if !ok {
		return false
	}
	for _, date := range tuple {
		if !date.HasValidFormat() {
			return false
		}
	}
	return true
}

func isRFCDate(t interface{}) bool {
	date, ok := t.(RFCDate)
	return ok && date.HasValidFormat()
}

var (
	ValidFormats = [][]RFCDateFormat{
		{RFCDateFormatDay, RFCDateFormatMonth, RFCDateFormatYear},
		{RFCDateFormatDay, RFCDateFormatMonth, RFCDateFormatYear, RFCDateFormatTime},
		{RFCDateFormatMonth, RFCDateFormatYear},
		{RFCDateFormatDay, RFCDateFormatMonth},
		{RFCDateFormatYear},
	}
)

func (rfcDate *RFCDate) HasValidFormat() bool {
	for _, validFormat := range ValidFormats {
		if HaveSameElements(rfcDate.Format, validFormat) {
			return true
		}
	}
	return false
}

func isInt(t interface{}) bool {
	_, ok := t.(int)
	return ok
}

func isIntSlice(t interface{}) bool {
	_, ok := t.([]int)
	return ok
}

func isStringSlice(t interface{}) bool {
	_, ok := t.([]string)
	return ok
}

func isString(t interface{}) bool {
	_, ok := t.(string)
	return ok
}

func isUUID(t interface{}) bool {
	_, ok := t.(uuid.UUID)
	return ok
}

func isUUIDSlice(t interface{}) bool {
	_, ok := t.([]uuid.UUID)
	return ok
}
