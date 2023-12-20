package counters

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
)

// PacketCounters is a data structure to collect packet and byte counters
type PacketCounters struct {
	InCounter  int64
	OutCounter int64
	InBytes    int64
	OutBytes   int64
}

// AddPacket increment the counters based on information contained in pkt
func (c *PacketCounters) AddPacket(pkt *network.Packet) error {
	if pkt.Dir == network.TrafficIn {
		c.InCounter++
		c.InBytes += pkt.Length
	} else if pkt.Dir == network.TrafficOut {
		c.OutCounter++
		c.OutBytes += pkt.Length
	}
	log.Debugf("Updated counter of type %s with packet of dir %d and length %d", c.Type(), pkt.Dir, pkt.DataLength)
	return nil
}

// Reset resets all counters
func (c *PacketCounters) Reset() error {
	c.InCounter = 0
	c.OutCounter = 0
	c.InBytes = 0
	c.OutBytes = 0
	return nil
}

// Clear clears all counters (same behavior as reset)
func (c *PacketCounters) Clear() error {
	c.InCounter = 0
	c.OutCounter = 0
	c.InBytes = 0
	c.OutBytes = 0
	return nil
}

// Type returns a string with the type name of the counter.
func (c *PacketCounters) Type() string {
	return "PacketCounters"
}

// Collect returns a []byte representation of the counter
func (c *PacketCounters) Collect() []byte {
	b, _ := json.Marshal(c)
	return b
}
