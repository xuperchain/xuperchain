// Package query provides a parser for a custom query format:
//
//		abci.invoice.number=22 AND abci.invoice.owner=Ivan
//
// See query.peg for the grammar, which is a https://en.wikipedia.org/wiki/Parsing_expression_grammar.
// More: https://github.com/PhilippeSigaud/Pegged/wiki/PEG-Basics
//
// It has a support for numbers (integer and floating point), dates and times.
package query

import (
	"fmt"
)

type Query interface {
	Matches(tags Tagged) bool
	String() string
	MatchError() error
}

var _ Query = &PegQuery{}
var _ Queryable = &PegQuery{}

// Query holds the query string and the query parser.
type PegQuery struct {
	str    string
	parser *QueryParser
	error  *MatchError
}

type MatchError struct {
	Tagged Tagged
	Cause  error
}

func (m *MatchError) Error() string {
	return fmt.Sprintf("matching error occurred with tagged: %v", m.Cause)
}

// Condition represents a single condition within a query and consists of tag
// (e.g. "tx.gas"), operator (e.g. "=") and operand (e.g. "7").
type Condition struct {
	Tag     string
	Op      Operator
	Operand interface{}
}

// New parses the given string and returns a query or error if the string is
// invalid.
func New(s string) (*PegQuery, error) {
	p := &QueryParser{
		Buffer: s,
	}
	p.Expression.explainer = func(format string, args ...interface{}) {
		fmt.Printf(format, args...)
	}
	err := p.Init()
	if err != nil {
		return nil, err
	}
	err = p.Parse()
	if err != nil {
		return nil, err
	}
	p.Execute()
	return &PegQuery{str: s, parser: p}, nil
}

// MustParse turns the given string into a query or panics; for tests or others
// cases where you know the string is valid.
func MustParse(s string) *PegQuery {
	q, err := New(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s: %v", s, err))
	}
	return q
}

// String returns the original string.
func (q *PegQuery) String() string {
	return q.str
}

func (q *PegQuery) Query() (Query, error) {
	return q, nil
}

// Matches returns true if the query matches the given set of tags, false otherwise.
//
// For example, query "name=John" matches tags = {"name": "John"}. More
// examples could be found in parser_test.go and query_test.go.
func (q *PegQuery) Matches(tags Tagged) bool {
	match, err := q.parser.Evaluate(tags.Get)
	if err != nil {
		q.error = &MatchError{Cause: err, Tagged: tags}
		return false
	}
	return match
}

// Returns whether a matching error occurred (which would result in a false from Matches)
func (q *PegQuery) MatchError() error {
	if q.error == nil {
		return nil
	}
	return q.error
}

func (q *PegQuery) ExplainTo(explainer func(fmt string, args ...interface{})) {
	q.parser.explainer = explainer
}
