package utils

import (
	"encoding/json"
	"time"
)

type Filter struct {
	Relation   Relation    `json:"relation"`
	Conditions []Condition `json:"conditions"`
	Exclude    Excludable  `json:"exclude"`
}

type Excludable struct {
	Users []int `json:"users"`
}

type FieldFilter struct {
	FieldName FieldCount `json:"fieldName"`
	Filter    Filter     `json:"filter"`
}

type Condition struct {
	FieldName FieldName   `json:"fieldName"`
	Operator  Operator    `json:"operator"`
	Value     interface{} `json:"value"` // RFCDate | [2]RFCDate | int | []int | string
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
		switch v[0].(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			c.Value = castToIntSlice(v)
		case map[string]interface{}:
			c.Value = castToRFCDateTuple(v)
		case string:
			c.Value = castToStringSlice(v)
		}
	case interface{}:
		switch v := v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			c.Value = castToint(v)
		case map[string]interface{}:
			c.Value = castToRFCDate(v)
		case string:
			c.Value = string(v)
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

func castToRFCDateTuple(v []interface{}) [2]RFCDate {
	var result [2]RFCDate
	for i, vv := range v {
		switch v := vv.(type) {
		case map[string]interface{}:
			result[i] = castToRFCDate(v)
		}
	}
	return result
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

func castToRFCDate(v map[string]interface{}) RFCDate {
	var r RFCDate
	if d, ok := v["date"]; ok {
		r.Date, _ = time.Parse(time.RFC3339, d.(string))
	}
	if f, ok := v["format"]; ok {
		for _, ff := range f.([]interface{}) {
			r.Format = append(r.Format, RFCDateFormat(ff.(string)))
		}
	}
	return r
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
	FieldNameHierarchy    FieldName = "hierarchy"

	FieldNameName  FieldName = "name"
	FieldNameEmail FieldName = "email"
	FieldNamePhone FieldName = "phone"
)

type FieldCount string

const (
	FieldCountDepartment  FieldCount = "department"
	FieldCountJob         FieldCount = "job"
	FieldCountCompanySite FieldCount = "companySite"
	FieldCountHierarchy   FieldCount = "hierarchy"
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
	ValidConditions = []ValidateCondition{
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
			Fields: []FieldName{FieldNameDepartmentId, FieldNameJobId, FieldNameCompanySite, FieldNameHierarchy},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorEq, OperatorIn, OperatorNotIn},
					ValueTypeValidators: []func(interface{}) bool{isInt, isIntSlice},
				},
			},
		},
		{
			Fields: []FieldName{FieldNameName, FieldNameEmail, FieldNamePhone},
			ValidOperators: []ValidateOperator{
				{
					Operators:           []Operator{OperatorEq},
					ValueTypeValidators: []func(interface{}) bool{isString},
				},
			},
		},
	}
)

func (filter *Filter) Validate() bool {
	return isEveryFilterConditionValid(filter, ValidConditions)
}

func isEveryFilterConditionValid(filter *Filter, validConditions []ValidateCondition) bool {
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

func isString(t interface{}) bool {
	_, ok := t.(string)
	return ok
}
