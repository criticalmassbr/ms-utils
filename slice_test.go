package utils_test

import (
	"fmt"
	"reflect"

	utils "github.com/criticalmassbr/ms-utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestRemoveDuplicates(t *testing.T) {
	type args struct {
		a []int
		b []string
	}

	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "empty slices are equal",
			args: args{
				a: []int{},
			},
			want: []int{},
		},
		{
			name: "should remove duplicates when slice is ordered",
			args: args{
				a: []int{1, 1, 1, 1, 1, 1},
			},
			want: []int{1},
		},
		{
			name: "should return same slice if no duplicates",
			args: args{
				a: []int{5, 1, 8, 2, 0, 3, 6, 4, 7},
			},
			want: []int{5, 1, 8, 2, 0, 3, 6, 4, 7},
		},
		{
			name: "should return same slice if there are duplicates but slice is not ordered",
			args: args{
				a: []int{5, 1, 2, 0, 8, 2, 5, 0, 1, 3, 6, 4, 7, 8},
			},
			want: []int{5, 1, 2, 0, 8, 2, 5, 0, 1, 3, 6, 4, 7, 8},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("int - %s", tt.name), func(t *testing.T) {
			got := utils.RemoveDuplicatesFromOrderedSlice(tt.args.a)
			assert.True(t, reflect.DeepEqual(tt.want, got), "got %v, want %v", got, tt.want)
		})
	}

	testsString := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty slices are equal",
			args: args{
				b: []string{},
			},
			want: []string{},
		},
		{
			name: "should remove duplicates when slice is ordered",
			args: args{
				b: []string{"1", "1", "1", "1", "1", "1"},
			},
			want: []string{"1"},
		},
		{
			name: "should return same slice if no duplicates",
			args: args{
				b: []string{"5", "1", "8", "2", "0", "3", "6", "4", "7"},
			},
			want: []string{"5", "1", "8", "2", "0", "3", "6", "4", "7"},
		},
		{
			name: "should return same slice if there are duplicates but slice is not ordered",
			args: args{
				b: []string{"5", "1", "2", "0", "8", "2", "5", "0", "1", "3", "6", "4", "7", "8"},
			},
			want: []string{"5", "1", "2", "0", "8", "2", "5", "0", "1", "3", "6", "4", "7", "8"},
		},
	}

	for _, tt := range testsString {
		t.Run(fmt.Sprintf("string - %s", tt.name), func(t *testing.T) {
			got := utils.RemoveDuplicatesFromOrderedSlice(tt.args.b)
			assert.True(t, reflect.DeepEqual(tt.want, got), "got %v, want %v", got, tt.want)
		})
	}
}

func TestSortAndRemoveDuplicates(t *testing.T) {
	type args struct {
		a []int
		b []string
	}

	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "empty slices are equal",
			args: args{
				a: []int{},
			},
			want: []int{},
		},
		{
			name: "should remove duplicates",
			args: args{
				a: []int{1, 1, 1, 1, 1, 1},
			},
			want: []int{1},
		},
		{
			name: "should remove duplicates and sort",
			args: args{
				a: []int{5, 1, 8, 2, 0, 3, 6, 4, 7},
			},
			want: []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("int - %s", tt.name), func(t *testing.T) {
			got := utils.SortAndRemoveDuplicates(tt.args.a)
			assert.True(t, reflect.DeepEqual(tt.want, got), "got %v, want %v", got, tt.want)
		})
	}

	testsString := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty slices are equal",
			args: args{
				b: []string{},
			},
			want: []string{},
		},
		{
			name: "should remove duplicates",
			args: args{
				b: []string{"1", "1", "1", "1", "1", "1"},
			},
			want: []string{"1"},
		},
		{
			name: "should remove duplicates and sort",
			args: args{
				b: []string{"5", "1", "8", "2", "0", "3", "6", "4", "7"},
			},
			want: []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"},
		},
	}

	for _, tt := range testsString {
		t.Run(fmt.Sprintf("string - %s", tt.name), func(t *testing.T) {
			got := utils.SortAndRemoveDuplicates(tt.args.b)
			assert.True(t, reflect.DeepEqual(tt.want, got), "got %v, want %v", got, tt.want)
		})
	}
}

func TestSortAndRemoveDuplicateUUIDs(t *testing.T) {
	type args struct {
		uuids []uuid.UUID
	}

	tests := []struct {
		name string
		args args
		want []uuid.UUID
	}{
		{
			name: "empty slices are equal",
			args: args{
				uuids: []uuid.UUID{},
			},
			want: []uuid.UUID{},
		},
		{
			name: "should remove duplicates",
			args: args{
				uuids: []uuid.UUID{
					uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					uuid.MustParse("00000000-0000-0000-0000-000000000000"),
				},
			},
			want: []uuid.UUID{
				uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			},
		},
		{
			name: "should remove duplicates and sort",
			args: args{
				uuids: []uuid.UUID{
					uuid.MustParse("00000000-0000-0000-0000-000000000005"),
					uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					uuid.MustParse("00000000-0000-0000-0000-000000000008"),
					uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					uuid.MustParse("00000000-0000-0000-0000-000000000006"),
					uuid.MustParse("00000000-0000-0000-0000-000000000004"),
					uuid.MustParse("00000000-0000-0000-0000-000000000007"),
				},
			},
			want: []uuid.UUID{
				uuid.MustParse("00000000-0000-0000-0000-000000000000"),
				uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				uuid.MustParse("00000000-0000-0000-0000-000000000003"),
				uuid.MustParse("00000000-0000-0000-0000-000000000004"),
				uuid.MustParse("00000000-0000-0000-0000-000000000005"),
				uuid.MustParse("00000000-0000-0000-0000-000000000006"),
				uuid.MustParse("00000000-0000-0000-0000-000000000007"),
				uuid.MustParse("00000000-0000-0000-0000-000000000008"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.SortAndRemoveDuplicateUUIDs(tt.args.uuids)
			assert.True(t, reflect.DeepEqual(tt.want, got), "got %v, want %v", got, tt.want)
		})
	}
}

func TestSortAndRemoveDuplicateConditions(t *testing.T) {
	type args struct {
		conditions []utils.Condition
	}

	tests := []struct {
		name string
		args args
		want []utils.Condition
	}{
		{
			name: "empty slices are equal",
			args: args{
				conditions: []utils.Condition{},
			},
			want: []utils.Condition{},
		},
		{
			name: "should order by field name followed by operator name",
			args: args{
				conditions: []utils.Condition{
					{
						FieldName: utils.FieldNameEmail,
						Operator:  utils.OperatorEq,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameCreatedAt,
						Operator:  utils.OperatorNotEq,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameUpdatedAt,
						Operator:  utils.OperatorIn,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameBirthday,
						Operator:  utils.OperatorNotIn,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameBirthday,
						Operator:  utils.OperatorBetween,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameHireDate,
						Operator:  utils.OperatorBetween,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameDepartmentId,
						Operator:  utils.OperatorGt,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameEmail,
						Operator:  utils.OperatorNotEq,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameEmail,
						Operator:  utils.OperatorIn,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameJobId,
						Operator:  utils.OperatorLt,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameCompanySite,
						Operator:  utils.OperatorEq,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameHierarchy,
						Operator:  utils.OperatorEq,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameName,
						Operator:  utils.OperatorEq,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameEmail,
						Operator:  utils.OperatorNotIn,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNamePhone,
						Operator:  utils.OperatorEq,
						Value:     nil,
					},
				},
			},
			want: []utils.Condition{
				{
					FieldName: utils.FieldNameBirthday,
					Operator:  utils.OperatorBetween,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameBirthday,
					Operator:  utils.OperatorNotIn,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameCompanySite,
					Operator:  utils.OperatorEq,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameCreatedAt,
					Operator:  utils.OperatorNotEq,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameDepartmentId,
					Operator:  utils.OperatorGt,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameEmail,
					Operator:  utils.OperatorEq,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameEmail,
					Operator:  utils.OperatorIn,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameEmail,
					Operator:  utils.OperatorNotEq,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameEmail,
					Operator:  utils.OperatorNotIn,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameHierarchy,
					Operator:  utils.OperatorEq,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameHireDate,
					Operator:  utils.OperatorBetween,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameJobId,
					Operator:  utils.OperatorLt,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameName,
					Operator:  utils.OperatorEq,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNamePhone,
					Operator:  utils.OperatorEq,
					Value:     nil,
				},
				{
					FieldName: utils.FieldNameUpdatedAt,
					Operator:  utils.OperatorIn,
					Value:     nil,
				},
			},
		},
		{
			name: "should remove duplicate Conditions that have same field name and operator name",
			args: args{
				conditions: []utils.Condition{
					{
						FieldName: utils.FieldNameUpdatedAt,
						Operator:  utils.OperatorIn,
						Value:     nil,
					},
					{
						FieldName: utils.FieldNameUpdatedAt,
						Operator:  utils.OperatorIn,
						Value:     nil,
					},
				},
			},
			want: []utils.Condition{
				{
					FieldName: utils.FieldNameUpdatedAt,
					Operator:  utils.OperatorIn,
					Value:     nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.SortAndRemoveDuplicateConditions(tt.args.conditions)
			assert.True(t, reflect.DeepEqual(tt.want, got), "got %v, want %v", got, tt.want)
		})
	}
}

func TestRemove(t *testing.T) {
	type args struct {
		slice   []int
		element int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "remove element from ordered slice",
			args: args{
				slice:   []int{1, 2, 3, 4, 5},
				element: 3,
			},
			want: []int{1, 2, 4, 5},
		},
		{
			name: "remove element from unordered slice",
			args: args{
				slice:   []int{5, 1, 3, 2, 4},
				element: 1,
			},
			want: []int{5, 3, 2, 4},
		},
		{
			name: "remove element from slice with only one element",
			args: args{
				slice:   []int{1},
				element: 1,
			},
			want: []int{},
		},
		{
			name: "remove element from empty slice",
			args: args{
				slice:   []int{},
				element: 1,
			},
			want: []int{},
		},
		{
			name: "remove element from slice with duplicated elements",
			args: args{
				slice:   []int{1, 1, 1, 1, 1},
				element: 1,
			},
			want: []int{1, 1, 1, 1},
		},
		{
			name: "remove element from slice with duplicated elements and other elements",
			args: args{
				slice:   []int{2, 1, 3, 1, 5, 4, 1, 10, 1, 9, 1},
				element: 1,
			},
			want: []int{2, 3, 1, 5, 4, 1, 10, 1, 9, 1},
		},
		{
			name: "remove nothing from slice with element not in slice",
			args: args{
				slice:   []int{1, 2, 3, 4, 5},
				element: 6,
			},
			want: []int{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.RemoveFirstOcurrence(tt.args.slice, tt.args.element)
			assert.True(t, reflect.DeepEqual(tt.want, got), "got %v, want %v", got, tt.want)
		})
	}
}
