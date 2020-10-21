package query

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/hyperledger/burrow/logging/errors"
)

const (
	// DateLayout defines a layout for all dates (`DATE date`)
	DateLayout = "2006-01-02"
	// TimeLayout defines a layout for all times (`TIME time`)
	TimeLayout = time.RFC3339
)

// Operator is an operator that defines some kind of relation between tag and
// operand (equality, etc.).
type Operator uint8

const (
	OpTerminal Operator = iota
	OpAnd
	OpOr
	OpLessEqual
	OpGreaterEqual
	OpLess
	OpGreater
	OpEqual
	OpContains
)

var opNames = map[Operator]string{
	OpAnd:          "AND",
	OpOr:           "OR",
	OpLessEqual:    "<=",
	OpGreaterEqual: ">=",
	OpLess:         "<",
	OpGreater:      ">",
	OpEqual:        "=",
	OpContains:     "CONTAINS",
}

func (op Operator) String() string {
	return opNames[op]
}

// Instruction is a container suitable for the code tape and the stack to hold values an operations
type instruction struct {
	op     Operator
	tag    *string
	string *string
	time   *time.Time
	number *big.Float
	match  bool
}

func (in *instruction) String() string {
	switch {
	case in.op != OpTerminal:
		return in.op.String()
	case in.tag != nil:
		return *in.tag
	case in.string != nil:
		return "'" + *in.string + "'"
	case in.time != nil:
		return in.time.String()
	case in.number != nil:
		return in.number.String()
	default:
		if in.match {
			return "true"
		}
		return "false"
	}
}

// A Boolean expression for the query grammar
type Expression struct {
	// This is our 'bytecode'
	code      []*instruction
	errors    errors.MultipleErrors
	explainer func(format string, args ...interface{})
}

// Evaluate expects an Execute() to have filled the code of the Expression so it can be run in the little stack machine
// below
func (e *Expression) Evaluate(getTagValue func(tag string) (interface{}, bool)) (bool, error) {
	if len(e.errors) > 0 {
		return false, e.errors
	}
	var left, right *instruction
	stack := make([]*instruction, 0, len(e.code))
	for _, in := range e.code {
		if in.op == OpTerminal {
			// just push terminals on to the stack
			stack = append(stack, in)
			continue
		}

		if len(stack) < 2 {
			return false, fmt.Errorf("cannot pop from stack for query expression [%v] because stack has "+
				"fewer than 2 elements", e)
		}
		stack, left, right = pop(stack)
		ins := &instruction{}
		switch in.op {
		case OpAnd:
			ins.match = left.match && right.match
		case OpOr:
			ins.match = left.match || right.match
		default:
			// We have a a non-terminal, non-connective operation
			tagValue, ok := getTagValue(*left.tag)
			// No match if we can't get tag value
			if ok {
				switch {
				case right.string != nil:
					ins.match = compareString(in.op, tagValue, *right.string)
				case right.number != nil:
					ins.match = compareNumber(in.op, tagValue, right.number)
				case right.time != nil:
					ins.match = compareTime(in.op, tagValue, *right.time)
				}
			}
			// Uncomment this for a little bit of debug:
			//e.explainf("%v := %v\n", left, tagValue)
		}
		// Uncomment this for a little bit of debug:
		//e.explainf("%v %v %v => %v\n", left, in.op, right, ins.match)

		// Push whether this was a match back on to stack
		stack = append(stack, ins)
	}
	if len(stack) != 1 {
		return false, fmt.Errorf("stack for query expression [%v] should have exactly one element after "+
			"evaulation but has %d", e, len(stack))
	}
	return stack[0].match, nil
}

func (e *Expression) explainf(fmt string, args ...interface{}) {
	if e.explainer != nil {
		e.explainer(fmt, args...)
	}
}

func pop(stack []*instruction) ([]*instruction, *instruction, *instruction) {
	return stack[:len(stack)-2], stack[len(stack)-2], stack[len(stack)-1]
}

func compareString(op Operator, tagValue interface{}, value string) bool {
	tagString := StringFromValue(tagValue)
	switch op {
	case OpContains:
		return strings.Contains(tagString, value)
	case OpEqual:
		return tagString == value
	}
	return false
}

func compareNumber(op Operator, tagValue interface{}, value *big.Float) bool {
	tagNumber := new(big.Float)
	switch n := tagValue.(type) {
	case string:
		f, _, err := big.ParseFloat(n, 10, 64, big.ToNearestEven)
		if err != nil {
			return false
		}
		tagNumber.Set(f)
	case *big.Float:
		tagNumber.Set(n)
	case *big.Int:
		tagNumber.SetInt(n)
	case float32:
		tagNumber.SetFloat64(float64(n))
	case float64:
		tagNumber.SetFloat64(n)
	case int:
		tagNumber.SetInt64(int64(n))
	case int32:
		tagNumber.SetInt64(int64(n))
	case int64:
		tagNumber.SetInt64(n)
	case uint:
		tagNumber.SetUint64(uint64(n))
	case uint32:
		tagNumber.SetUint64(uint64(n))
	case uint64:
		tagNumber.SetUint64(n)
	default:
		return false
	}
	cmp := tagNumber.Cmp(value)
	switch op {
	case OpLessEqual:
		return cmp < 1
	case OpGreaterEqual:
		return cmp > -1
	case OpLess:
		return cmp == -1
	case OpGreater:
		return cmp == 1
	case OpEqual:
		return cmp == 0
	}
	return false
}

func compareTime(op Operator, tagValue interface{}, value time.Time) bool {
	var tagTime time.Time
	var err error
	switch t := tagValue.(type) {
	case time.Time:
		tagTime = t
	case int64:
		// Hmmm, should we?
		tagTime = time.Unix(t, 0)
	case string:
		tagTime, err = time.Parse(TimeLayout, t)
		if err != nil {
			tagTime, err = time.Parse(DateLayout, t)
			if err != nil {
				return false
			}
		}
	default:
		return false
	}
	switch op {
	case OpLessEqual:
		return tagTime.Before(value) || tagTime.Equal(value)
	case OpGreaterEqual:
		return tagTime.Equal(value) || tagTime.After(value)
	case OpLess:
		return tagTime.Before(value)
	case OpGreater:
		return tagTime.After(value)
	case OpEqual:
		return tagTime.Equal(value)
	}
	return false
}

// These methods implement the various visitors that are called in the PEG grammar with statements like
// { p.Operator(OpEqual) }

func (e *Expression) String() string {
	strs := make([]string, len(e.code))
	for i, in := range e.code {
		strs[i] = in.String()
	}
	return strings.Join(strs, ", ")
}

func (e *Expression) Operator(operator Operator) {
	e.code = append(e.code, &instruction{
		op: operator,
	})
}

// Terminals...

func (e *Expression) Tag(value string) {
	e.code = append(e.code, &instruction{
		tag: &value,
	})
}

func (e *Expression) Time(value string) {
	t, err := time.Parse(TimeLayout, value)
	e.pushErr(err)
	e.code = append(e.code, &instruction{
		time: &t,
	})

}
func (e *Expression) Date(value string) {
	date, err := time.Parse(DateLayout, value)
	e.pushErr(err)
	e.code = append(e.code, &instruction{
		time: &date,
	})
}

func (e *Expression) Number(value string) {
	number, _, err := big.ParseFloat(value, 10, 64, big.ToNearestEven)
	e.pushErr(err)
	e.code = append(e.code, &instruction{
		number: number,
	})
}

func (e *Expression) Value(value string) {
	e.code = append(e.code, &instruction{
		string: &value,
	})
}

func (e *Expression) pushErr(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}
