package exec

import (
	"fmt"
)

func (h *Header) String() string {
	if h == nil {
		return fmt.Sprintf("Header{<Empty>}")
	}
	return fmt.Sprintf("Header{Tx{%v}: %v; Event{%v}: %v; Height: %v; Index: %v; Exception: %v}",
		h.TxType, h.TxHash, h.EventType, h.EventID, h.Height, h.Index, h.Exception)
}
