package query

// Empty query matches any set of tags.
type Empty struct {
}

var _ Query = Empty{}
var _ Queryable = Empty{}

// Matches always returns true.
func (Empty) Matches(tags Tagged) bool {
	return true
}

func (Empty) String() string {
	return "empty"
}

func (Empty) Query() (Query, error) {
	return Empty{}, nil
}

func (empty Empty) MatchError() error {
	return nil
}
