package counters

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"log"
	"math"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
)

// ByteCopy is a data structure to collect packet and byte counters
type PNGCopyCounters struct {
	CopiedBytes int32
	StoredBytes int32
	ToCopy      int32
	Layers      uint8
	Image       *image.Gray
	Buffer      *bytes.Buffer
	Created     bool
}

// AddPacket increment the counters based on information contained in pkt
func (c *PNGCopyCounters) AddPacket(pkt *network.Packet) {
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
				copied := int32(copy(c.Image.Pix[c.StoredBytes:], pkt.RawData[0:(c.ToCopy-c.CopiedBytes)]))
				c.CopiedBytes += copied
				c.StoredBytes += copied
			} else {
				// Copy the entire header size
				copied := int32(copy(c.Image.Pix[c.StoredBytes:], pkt.RawData[0:headerSize]))
				c.CopiedBytes += copied
				c.StoredBytes += copied
			}
		}
		if c.Layers > 0 {
			// Copy Payload
			if int64(c.ToCopy-c.CopiedBytes) < pkt.DataLength {
				// Copy just the remaining data
				copied := int32(copy(c.Image.Pix[c.StoredBytes:], pkt.RawData[headerSize:headerSize+int64(c.ToCopy-c.CopiedBytes)]))
				c.CopiedBytes += copied
				c.StoredBytes += copied
			} else {
				// Copy the entire header size
				copied := int32(copy(c.Image.Pix[c.StoredBytes:], pkt.RawData[headerSize:headerSize+pkt.DataLength]))
				c.CopiedBytes += copied
				c.StoredBytes += copied
			}
		}
	}
	if !c.Created && c.CopiedBytes >= c.ToCopy {
		if err := png.Encode(c.Buffer, c.Image); err != nil {
			log.Fatal(err)
		} else {
			c.Created = true
		}
	}
}

// Reset resets all counters
func (c *PNGCopyCounters) Reset() {
	c.ToCopy = 400
	c.Layers = 0
	c.CopiedBytes = 0
	c.StoredBytes = 0
	side := int(math.Sqrt(float64(c.ToCopy)))
	c.Image = image.NewGray(image.Rect(0, 0, side, side))
	c.Image.Pix = make([]byte, c.ToCopy)
	c.Buffer = &bytes.Buffer{}
	c.Buffer.Reset()
}

// Clear clears all counters
func (c *PNGCopyCounters) Clear() {
	c.StoredBytes = 0
	if c.CopiedBytes >= c.ToCopy && c.Created {
		c.Image = nil
		c.Buffer = nil
	}
}

// Type returns a string with the type name of the counter.
func (c *PNGCopyCounters) Type() string {
	return "PNGCopyCounters"
}

type PNGCopyCountersOut struct {
	CopiedBytes int32
	Data        []byte
}

// Collect returns a []byte representation of the counter
func (c *PNGCopyCounters) Collect() []byte {
	if c.Created {
		b, _ := json.Marshal(ByteCopyCountersOut{
			CopiedBytes: c.CopiedBytes,
			Data:        c.Buffer.Bytes(),
		})
		return b
	} else {
		b, _ := json.Marshal(ByteCopyCountersOut{
			CopiedBytes: c.CopiedBytes,
			Data:        nil,
		})
		return b
	}

}
