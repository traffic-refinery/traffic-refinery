package counters

import (
	"github.com/traffic-refinery/traffic-refinery/internal/network"
)

// Counter is a general counter interface.
// Functions that all flow type structures have to implement.
type Counter interface {
	// Updates the flow states based on the packet
	AddPacket(pkt *network.Packet) error
	// Reset resets the counter statistics for periodic counting.
	// This function triggers at the "emit" time.
	Reset() error
	// Clear re-initializes the counter to its zero state
	Clear() error
	// Type returns a string with the type name of the counter.
	Type() string
	// Collect returns a []byte representation of the counter
	Collect() []byte
}
