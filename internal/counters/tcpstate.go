package counters

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/welford"
)

// TCPState is a data structure to collect packet and byte counters
type TCPState struct {
	AckUpCounter     int64
	AckDownCounter   int64
	SynUpCounter     int64
	SynDownCounter   int64
	RstUpCounter     int64
	RstDownCounter   int64
	PushUpCounter    int64
	PushDownCounter  int64
	UrgUpCounter     int64
	UrgDownCounter   int64
	BytesUpCounter   int64
	BytesDownCounter int64
	RetrUpCounter    int64
	RetrDownCounter  int64
	OOOUpCounter     int64
	OOODownCounter   int64

	UpRecWindow       welford.Welford
	DownRecWindow     welford.Welford
	UpBytesPerPkt     welford.Welford
	DownBytesPerPkt   welford.Welford
	UpBytesInFlight   welford.Welford
	DownBytesInFlight welford.Welford
	RTT               welford.Welford

	unAckedUp   map[uint32]int64
	unAckedDown map[uint32]int64

	lastSeqUp, lastSeqDown, lastAckUp, lastAckDown uint32
}

func (c *TCPState) AddPacket(pkt *network.Packet) error {
	if !pkt.IsTCP {
		log.Debugln("TCPState can not process a non TCP packet")
		return errors.New("can not process a non TCP packet")
	}
	if pkt.Dir == network.TrafficIn {
		// log.Debugf("Incoming traffic %t %d", pkt.Tcp.ACK, pkt.Tcp.Ack)
		if pkt.Tcp.ACK {
			c.AckDownCounter += 1
			if ts, ok := c.unAckedUp[pkt.Tcp.Ack]; ok {
				c.RTT.AddValue(float64(pkt.TStamp - ts))
				// log.Debugf("The RTT computed is %d", pkt.TStamp-ts)
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
		if pkt.Tcp.SYN {
			c.SynDownCounter += 1
		}
		if pkt.Tcp.RST {
			c.RstDownCounter += 1
		}
		if pkt.Tcp.PSH {
			c.PushDownCounter += 1
		}
		if pkt.Tcp.URG {
			c.UrgDownCounter += 1
		}
		bytes := pkt.Length - 4*int64(pkt.Tcp.DataOffset)
		c.BytesDownCounter += bytes
		c.DownRecWindow.AddValue(float64(pkt.Tcp.Window))
		c.DownBytesPerPkt.AddValue(float64(bytes))
		c.DownBytesInFlight.AddValue(float64(pkt.SeqNumber + uint32(bytes) - c.lastAckUp))
		// Only process retransmissions for packets with data
		if bytes > 0 {
			if pkt.Tcp.Seq < c.lastSeqDown {
				c.OOODownCounter += 1
			} else {
				c.lastSeqDown = pkt.Tcp.Seq
			}
			if _, ok := c.unAckedDown[pkt.Tcp.Seq+uint32(bytes)]; !ok {
				c.unAckedDown[pkt.Tcp.Seq+uint32(bytes)] = pkt.TStamp
			} else {
				c.RetrDownCounter += 1
			}
		}
	} else if pkt.Dir == network.TrafficOut {
		if pkt.Tcp.ACK {
			c.AckUpCounter += 1
			if pkt.Tcp.Ack > c.lastAckUp {
				c.lastAckUp = pkt.Tcp.Ack
				for key := range c.unAckedDown {
					if key < c.lastAckUp {
						delete(c.unAckedDown, key)
					}
				}
			}
		}
		if pkt.Tcp.SYN {
			c.SynUpCounter += 1
		}
		if pkt.Tcp.RST {
			c.RstUpCounter += 1
		}
		if pkt.Tcp.PSH {
			c.PushUpCounter += 1
		}
		if pkt.Tcp.URG {
			c.UrgUpCounter += 1
		}
		bytes := pkt.Length - 4*int64(pkt.Tcp.DataOffset)
		c.BytesUpCounter += bytes
		c.UpRecWindow.AddValue(float64(pkt.Tcp.Window))
		c.UpBytesPerPkt.AddValue(float64(bytes))
		c.UpBytesInFlight.AddValue(float64(pkt.SeqNumber + uint32(bytes) - c.lastAckDown))
		// Only process retransmissions for packets with data
		if bytes > 0 {
			if pkt.Tcp.Seq < c.lastSeqUp {
				c.OOOUpCounter += 1
			} else {
				c.lastSeqUp = pkt.Tcp.Seq
			}
			if _, ok := c.unAckedUp[pkt.Tcp.Seq+uint32(bytes)]; !ok {
				c.unAckedUp[pkt.Tcp.Seq+uint32(bytes)] = pkt.TStamp
			} else {
				c.RetrUpCounter += 1
			}
		}
	}
	return nil
}

func (c *TCPState) Reset() error {
	c.AckUpCounter = 0
	c.AckDownCounter = 0
	c.SynUpCounter = 0
	c.SynDownCounter = 0
	c.RstUpCounter = 0
	c.RstDownCounter = 0
	c.PushUpCounter = 0
	c.PushDownCounter = 0
	c.UrgUpCounter = 0
	c.UrgDownCounter = 0
	c.BytesUpCounter = 0
	c.BytesDownCounter = 0
	c.RetrUpCounter = 0
	c.RetrDownCounter = 0
	c.OOOUpCounter = 0
	c.OOODownCounter = 0

	c.UpRecWindow.Reset()
	c.DownRecWindow.Reset()
	c.UpBytesPerPkt.Reset()
	c.DownBytesPerPkt.Reset()
	c.UpBytesInFlight.Reset()
	c.DownBytesInFlight.Reset()

	c.RTT.Reset()

	// fmt.Println("Resetting counters")
	c.unAckedUp = make(map[uint32]int64)
	c.unAckedDown = make(map[uint32]int64)

	c.lastSeqUp = 0
	c.lastSeqDown = 0
	c.lastAckUp = 0
	c.lastAckDown = 0

	return nil
}

func (c *TCPState) Clear() error {
	c.AckUpCounter = 0
	c.AckDownCounter = 0
	c.SynUpCounter = 0
	c.SynDownCounter = 0
	c.RstUpCounter = 0
	c.RstDownCounter = 0
	c.PushUpCounter = 0
	c.PushDownCounter = 0
	c.UrgUpCounter = 0
	c.UrgDownCounter = 0
	c.BytesUpCounter = 0
	c.BytesDownCounter = 0
	c.RetrUpCounter = 0
	c.RetrDownCounter = 0
	c.OOOUpCounter = 0
	c.OOODownCounter = 0

	c.UpRecWindow.Reset()
	c.DownRecWindow.Reset()
	c.UpBytesPerPkt.Reset()
	c.DownBytesPerPkt.Reset()
	c.UpBytesInFlight.Reset()
	c.DownBytesInFlight.Reset()

	c.RTT.Reset()

	c.unAckedUp = make(map[uint32]int64)
	c.unAckedDown = make(map[uint32]int64)

	c.lastSeqUp = 0
	c.lastSeqDown = 0
	c.lastAckUp = 0
	c.lastAckDown = 0

	return nil
}

// Type returns a string with the type name of the counter.
func (c *TCPState) Type() string {
	return "TCPState"
}

type TCPStateOut struct {
	AckUpCounter     int64
	AckDownCounter   int64
	SynUpCounter     int64
	SynDownCounter   int64
	RstUpCounter     int64
	RstDownCounter   int64
	PushUpCounter    int64
	PushDownCounter  int64
	UrgUpCounter     int64
	UrgDownCounter   int64
	BytesUpCounter   int64
	BytesDownCounter int64
	RetrUpCounter    int64
	RetrDownCounter  int64
	OOOUpCounter     int64
	OOODownCounter   int64

	UpRecWindowAvg       float64
	DownRecWindowAvg     float64
	UpBytesPerPktAvg     float64
	DownBytesPerPktAvg   float64
	UpBytesInFlightAvg   float64
	DownBytesInFlightAvg float64
	RTTAvg               float64
	UpRecWindowVar       float64
	DownRecWindowVar     float64
	UpBytesPerPktVar     float64
	DownBytesPerPktVar   float64
	UpBytesInFlightVar   float64
	DownBytesInFlightVar float64
	RTTVar               float64
}

// Collect returns a []byte representation of the counter
func (c *TCPState) Collect() []byte {
	b, _ := json.Marshal(TCPStateOut{
		AckUpCounter:         c.AckUpCounter,
		AckDownCounter:       c.AckDownCounter,
		SynUpCounter:         c.SynUpCounter,
		SynDownCounter:       c.SynDownCounter,
		RstUpCounter:         c.RstUpCounter,
		RstDownCounter:       c.RstDownCounter,
		PushUpCounter:        c.PushUpCounter,
		PushDownCounter:      c.PushDownCounter,
		UrgUpCounter:         c.UrgUpCounter,
		UrgDownCounter:       c.UrgDownCounter,
		BytesUpCounter:       c.BytesUpCounter,
		BytesDownCounter:     c.BytesDownCounter,
		RetrUpCounter:        c.RetrUpCounter,
		RetrDownCounter:      c.RetrDownCounter,
		OOOUpCounter:         c.OOOUpCounter,
		OOODownCounter:       c.OOODownCounter,
		UpRecWindowAvg:       c.UpRecWindow.Avg,
		UpRecWindowVar:       c.UpRecWindow.Var,
		DownRecWindowAvg:     c.DownRecWindow.Avg,
		DownRecWindowVar:     c.DownRecWindow.Var,
		UpBytesPerPktAvg:     c.UpBytesPerPkt.Avg,
		UpBytesPerPktVar:     c.UpBytesPerPkt.Var,
		DownBytesPerPktAvg:   c.DownBytesPerPkt.Avg,
		DownBytesPerPktVar:   c.DownBytesPerPkt.Var,
		UpBytesInFlightAvg:   c.UpBytesInFlight.Avg,
		UpBytesInFlightVar:   c.UpBytesInFlight.Var,
		DownBytesInFlightAvg: c.DownBytesInFlight.Avg,
		DownBytesInFlightVar: c.DownBytesInFlight.Var,
		RTTAvg:               c.RTT.Avg,
		RTTVar:               c.RTT.Var,
	})
	return b
}
