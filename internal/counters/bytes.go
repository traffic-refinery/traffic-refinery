package counters

import (
	"encoding/json"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
)

const (
	HeadersOnly uint8 = 0
	AllLayers   uint8 = 1
	PayloadOnly uint8 = 2
)

// ByteCopy is a data structure to collect packet and byte counters
type ByteCopyCounters struct {
	CopiedBytes int32
	StoredBytes int32
	ToCopy      int32
	Layers      uint8
	Data        []byte
}

// AddPacket increment the counters based on information contained in pkt
func (c *ByteCopyCounters) AddPacket(pkt *network.Packet) {
	if c.CopiedBytes < c.ToCopy {
		// Initialize by accounting for the ethernet header
		headerSize := int64(14)
		if pkt.IsIPv4 {
			// Add IPv4 header size
			headerSize += int64(4 * uint16(pkt.Ip4.IHL))
		} else {
			headerSize += int64(pkt.Ip6.Length) - int64(pkt.DataLength)
		}
		// Add Transport header size
		headerSize += int64(pkt.Length) - int64(pkt.DataLength)
		if c.Layers < 2 {
			//Copy headers
			if int64(c.ToCopy-c.CopiedBytes) < headerSize {
				// Copy just the remaining data
				copied := int32(copy(c.Data[c.StoredBytes:], pkt.RawData[0:c.ToCopy-c.CopiedBytes]))
				c.CopiedBytes += copied
				c.StoredBytes += copied
			} else {
				// Copy the entire header size
				copied := int32(copy(c.Data[c.StoredBytes:], pkt.RawData[0:headerSize]))
				c.CopiedBytes += copied
				c.StoredBytes += copied
			}
		}
		if c.Layers > 0 && pkt.DataLength > 0 {
			// Copy Payload
			if int64(c.ToCopy-c.CopiedBytes) < pkt.DataLength {
				// Copy just the remaining data
				copied := int32(copy(c.Data[c.StoredBytes:], pkt.RawData[headerSize:(headerSize+int64(c.ToCopy-c.CopiedBytes))]))
				c.CopiedBytes += copied
				c.StoredBytes += copied
			} else {
				// Copy the entire header size
				copied := int32(copy(c.Data[c.StoredBytes:], pkt.RawData[headerSize:(headerSize+pkt.DataLength)]))
				c.CopiedBytes += copied
				c.StoredBytes += copied
			}
		}
	}
}

// Reset resets all counters
func (c *ByteCopyCounters) Reset() {
	c.CopiedBytes = 0
	c.StoredBytes = 0
	c.ToCopy = 400
	c.Layers = 0
	c.Data = make([]byte, c.ToCopy)
}

// Clear clears all counters
func (c *ByteCopyCounters) Clear() {
	c.StoredBytes = 0
	if c.CopiedBytes >= c.ToCopy {
		c.Data = nil
	}
}

// Type returns a string with the type name of the counter.
func (c *ByteCopyCounters) Type() string {
	return "ByteCopyCounters"
}

type ByteCopyCountersOut struct {
	CopiedBytes int32
	Data        []byte
}

// Collect returns a []byte representation of the counter
func (c *ByteCopyCounters) Collect() []byte {
	b, _ := json.Marshal(ByteCopyCountersOut{
		CopiedBytes: c.CopiedBytes,
		Data:        c.Data,
	})
	return b
}
