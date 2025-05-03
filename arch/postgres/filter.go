package postgres

import (
	"fmt"
	"strings"
)

// Filter provides SQL filter building capabilities
type Filter struct {
	conditions []string
	args       []interface{}
	paramCount int
}

// NewFilter creates a new SQL filter builder
func NewFilter() *Filter {
	return &Filter{
		conditions: []string{},
		args:       []interface{}{},
		paramCount: 0,
	}
}

// Add adds a condition to the filter
func (f *Filter) Add(condition string, args ...interface{}) *Filter {
	f.conditions = append(f.conditions, condition)
	for _, arg := range args {
		f.paramCount++
		modifiedCondition := strings.Replace(condition, "?", fmt.Sprintf("$%d", f.paramCount), -1)
		f.conditions[len(f.conditions)-1] = modifiedCondition
		f.args = append(f.args, arg)
	}
	return f
}

// AddIf adds a condition to the filter if the condition is true
func (f *Filter) AddIf(shouldAdd bool, condition string, args ...interface{}) *Filter {
	if shouldAdd {
		f.Add(condition, args...)
	}
	return f
}

// AddIfNotEmpty adds a condition to the filter if the value is not empty
func (f *Filter) AddIfNotEmpty(value string, condition string, args ...interface{}) *Filter {
	if value != "" {
		values := make([]interface{}, len(args))
		copy(values, args)
		values = append(values, value)
		f.Add(condition, values...)
	}
	return f
}

// AddIfNotZero adds a condition to the filter if the value is not zero
func (f *Filter) AddIfNotZero(value int64, condition string, args ...interface{}) *Filter {
	if value != 0 {
		values := make([]interface{}, len(args))
		copy(values, args)
		values = append(values, value)
		f.Add(condition, values...)
	}
	return f
}

// AddIn adds an IN condition to the filter
func (f *Filter) AddIn(field string, values []interface{}) *Filter {
	if len(values) == 0 {
		return f
	}

	placeholders := make([]string, len(values))
	for i := range values {
		f.paramCount++
		placeholders[i] = fmt.Sprintf("$%d", f.paramCount)
		f.args = append(f.args, values[i])
	}

	condition := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
	f.conditions = append(f.conditions, condition)
	return f
}

// Build builds the WHERE clause
func (f *Filter) Build() (string, []interface{}) {
	if len(f.conditions) == 0 {
		return "", f.args
	}

	where := "WHERE " + strings.Join(f.conditions, " AND ")
	return where, f.args
}

// BuildWithoutWhere builds the filter conditions without the WHERE keyword
func (f *Filter) BuildWithoutWhere() (string, []interface{}) {
	if len(f.conditions) == 0 {
		return "", f.args
	}

	conditions := strings.Join(f.conditions, " AND ")
	return conditions, f.args
}

// BuildWithJoin builds the filter conditions with a custom join operator
func (f *Filter) BuildWithJoin(joinOperator string) (string, []interface{}) {
	if len(f.conditions) == 0 {
		return "", f.args
	}

	where := "WHERE " + strings.Join(f.conditions, fmt.Sprintf(" %s ", joinOperator))
	return where, f.args
}

// Reset resets the filter
func (f *Filter) Reset() {
	f.conditions = []string{}
	f.args = []interface{}{}
	f.paramCount = 0
}

// Count returns the number of conditions in the filter
func (f *Filter) Count() int {
	return len(f.conditions)
}

// IsEmpty returns true if the filter is empty
func (f *Filter) IsEmpty() bool {
	return len(f.conditions) == 0
}
