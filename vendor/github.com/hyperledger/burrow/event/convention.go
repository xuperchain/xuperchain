package event

import (
	"fmt"

	"github.com/hyperledger/burrow/event/query"
)

const (
	EventTypeKey   = "EventType"
	EventIDKey     = "EventID"
	MessageTypeKey = "MessageType"
	TxHashKey      = "TxHash"
	HeightKey      = "Height"
	IndexKey       = "Index"
	StackDepthKey  = "StackDepth"
	AddressKey     = "Address"
)

type EventID string

func (eid EventID) Matches(tags query.Tagged) bool {
	value, ok := tags.Get(EventIDKey)
	if !ok {
		return false
	}
	return string(eid) == value
}

func (eid EventID) String() string {
	return fmt.Sprintf("%s = %s", EventIDKey, string(eid))
}

func (eid EventID) MatchError() error {
	return nil
}

// Get a query that matches events with a specific eventID
func QueryForEventID(eventID string) query.Queryable {
	// Since we're accepting external output here there is a chance it won't parse...
	return query.AsQueryable(EventID(eventID))
}
