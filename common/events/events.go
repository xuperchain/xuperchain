package events

// EventType event type definition
type EventType uint

// Definition for internal system events
const (
	/* Events for xchain core, start from 1 */
	// no events now

	/* Events for consensus, start from 1000 */
	// ProposerReady current consensus proposers ready for use
	ProposerReady = 1000
	// ProposerChanged next round consensus proposers ready for use
	ProposerChanged = 1010
)

// EventMessage is the event message body
type EventMessage struct {
	BcName   string
	Type     EventType
	Priority uint
	Sender   interface{}
	Message  interface{}
}

// EventHandler define the message handler
type EventHandler func(e *EventMessage)
