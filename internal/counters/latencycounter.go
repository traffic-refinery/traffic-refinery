package counters

import (
	"encoding/json"
	"errors"
	"math"

	log "github.com/sirupsen/logrus"
	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/welford"
)

// LatencyJitterCounter is a data structure to collect packet and byte counters
type LatencyJitterCounter struct {
	RTT    welford.Welford
	Jitter welford.Welford

	unAckedUp map[uint32]int64

	lastAckDown uint32
	lastLatency float64
}

func (c *LatencyJitterCounter) AddPacket(pkt *network.Packet) error {
	if !pkt.IsTCP {
		log.Debugln("TCPState can not process a non TCP packet")
		return errors.New("can not process a non TCP packet")
	}
	if pkt.Dir == network.TrafficIn {
		if pkt.Tcp.ACK {
			if ts, ok := c.unAckedUp[pkt.Tcp.Ack]; ok {
				newLatency := float64(pkt.TStamp - ts)
				c.RTT.AddValue(newLatency)
				c.Jitter.AddValue(math.Abs(c.lastLatency - newLatency))
				c.lastLatency = newLatency
			}
			if pkt.Tcp.Ack > c.lastAckDown {
				// log.Debugf("Removing acks")
				c.lastAckDown = pkt.Tcp.Ack
				for key := range c.unAckedUp {
					if key < c.lastAckDown {
						delete(c.unAckedUp, key)
					}
				}
			}
		}
	} else if pkt.Dir == network.TrafficOut {
		bytes := pkt.Length - 4*int64(pkt.Tcp.DataOffset)
		// Only process retransmissions for packets with data
		if bytes > 0 {
			if _, ok := c.unAckedUp[pkt.Tcp.Seq+uint32(bytes)]; !ok {
				c.unAckedUp[pkt.Tcp.Seq+uint32(bytes)] = pkt.TStamp
			}
		}
	}
	return nil
}

func (c *LatencyJitterCounter) Reset() error {
	c.RTT.Reset()
	c.Jitter.Reset()

	c.unAckedUp = make(map[uint32]int64)
	c.lastAckDown = 0
	return nil
}

func (c *LatencyJitterCounter) Clear() error {
	c.RTT.Reset()
	c.Jitter.Reset()
	return nil
}

// Type returns a string with the type name of the counter.
func (c *LatencyJitterCounter) Type() string {
	return "LatencyJitterCounter"
}

type LatencyJitterCounterOut struct {
	RTTAvg    float64
	RTTVar    float64
	JitterAvg float64
	JitterVar float64
}

// Collect returns a []byte representation of the counter
func (c *LatencyJitterCounter) Collect() []byte {
	b, _ := json.Marshal(LatencyJitterCounterOut{
		RTTAvg:    c.RTT.Avg,
		RTTVar:    c.RTT.Var,
		JitterAvg: c.Jitter.Avg,
		JitterVar: c.Jitter.Var,
	})
	return b
}
