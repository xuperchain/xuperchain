package matrics

import (
	"testing"
)

func TestNewServerMetrics(t *testing.T) {
	serverMetrics := NewServerMetrics()
	if serverMetrics == nil {
		t.Error("expected:", "not null", "actual:", "null")
	}
}
