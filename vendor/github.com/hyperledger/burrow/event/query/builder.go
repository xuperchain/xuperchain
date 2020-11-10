package query

import (
	"bytes"
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	MultipleValueTagSeparator = ";"

	// Operators
	equalString          = "="
	greaterThanString    = ">"
	lessThanString       = "<"
	greaterOrEqualString = ">="
	lessOrEqualString    = "<="
	containsString       = "CONTAINS"
	andString            = "AND"

	// Values
	trueString  = "true"
	falseString = "false"
	emptyString = "empty"
	timeString  = "TIME"
	dateString  = "DATE"
)

type Queryable interface {
	Query() (Query, error)
}

type parsedQuery struct {
	query Query
}

func AsQueryable(query Query) parsedQuery {
	return parsedQuery{query: query}
}

func (pq parsedQuery) Query() (Query, error) {
	return pq.query, nil
}

func Must(qry Query, err error) Query {
	if err != nil {
		panic(fmt.Errorf("could not compile: %v", qry))
	}
	return qry
}

// A yet-to-be-parsed query
type String string

func (qs String) Query() (Query, error) {
	if isEmpty(string(qs)) {
		return Empty{}, nil
	}
	return New(string(qs))
}

func MatchAllQueryable() Queryable {
	return Empty{}
}

// A fluent query builder
type Builder struct {
	queryString string
	condition
	// reusable buffer for building queryString
	bytes.Buffer
	error
}

// Templates
type condition struct {
	Tag     string
	Op      string
	Operand string
}

var conditionTemplate = template.Must(template.New("condition").Parse("{{.Tag}} {{.Op}} {{.Operand}}"))

// Creates a new query builder with a base query that is the conjunction of all queries passed
func NewBuilder(queries ...string) *Builder {
	qb := new(Builder)
	qb.queryString = qb.and(stringIterator(queries...))
	return qb
}

func (qb *Builder) String() string {
	return qb.queryString
}

func (qb *Builder) Query() (Query, error) {
	if qb.error != nil {
		return nil, qb.error
	}
	return NewOrEmpty(qb.queryString)
}

func NewOrEmpty(queryString string) (Query, error) {
	if isEmpty(queryString) {
		return Empty{}, nil
	}
	return New(queryString)
}

// Creates the conjunction of Builder and rightQuery
func (qb *Builder) And(queryBuilders ...*Builder) *Builder {
	return NewBuilder(qb.and(queryBuilderIterator(queryBuilders...)))
}

// Creates the conjunction of Builder and tag = operand
func (qb *Builder) AndEquals(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = equalString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndGreaterThanOrEqual(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = greaterOrEqualString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndLessThanOrEqual(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = lessOrEqualString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndStrictlyGreaterThan(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = greaterThanString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndStrictlyLessThan(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = lessThanString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndContains(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = containsString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) and(queryIterator func(func(string))) string {
	defer qb.Buffer.Reset()
	qb.Buffer.WriteString(qb.queryString)
	queryIterator(func(q string) {
		if !isEmpty(q) {
			if qb.Buffer.Len() > 0 {
				qb.Buffer.WriteByte(' ')
				qb.Buffer.WriteString(andString)
				qb.Buffer.WriteByte(' ')
			}
			qb.Buffer.WriteString(q)
		}
	})
	return qb.Buffer.String()
}

func operandString(value interface{}) string {
	buf := new(bytes.Buffer)
	switch v := value.(type) {
	case string:
		buf.WriteByte('\'')
		buf.WriteString(v)
		buf.WriteByte('\'')
		return buf.String()
	case fmt.Stringer:
		return operandString(v.String())
	default:
		return StringFromValue(v)
	}
}

func StringFromValue(value interface{}) string {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return "nil"
	}
	switch v := value.(type) {
	case string:
		return v
	case time.Time:
		return timeString + " " + v.Format(time.RFC3339)
	case encoding.TextMarshaler:
		bs, _ := v.MarshalText()
		return string(bs)
	case fmt.Stringer:
		return v.String()
	case bool:
		if v {
			return trueString
		}
		return falseString
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	default:
		if rv.Kind() == reflect.Slice {
			values := make([]string, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				values[i] = StringFromValue(rv.Index(i).Interface())
			}
			return strings.Join(values, MultipleValueTagSeparator)
		}
		return fmt.Sprintf("%v", v)
	}
}

func (qb *Builder) conditionString() string {
	defer qb.Buffer.Reset()
	err := conditionTemplate.Execute(&qb.Buffer, qb.condition)
	if err != nil && qb.error == nil {
		qb.error = err
	}
	return qb.Buffer.String()
}

func isEmpty(queryString string) bool {
	return queryString == "" || queryString == emptyString
}

// Iterators over some strings
func stringIterator(strs ...string) func(func(string)) {
	return func(callback func(string)) {
		for _, s := range strs {
			callback(s)
		}
	}
}

func queryBuilderIterator(qbs ...*Builder) func(func(string)) {
	return func(callback func(string)) {
		for _, qb := range qbs {
			callback(qb.String())
		}
	}
}
