package counters

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
)

// VideoSegment is the data structure used to keep track of segments being downloaded
type VideoSegment struct {
	Len       int64
	Seq       int64
	TsStart   int64
	TsEnd     int64
	LastPkt   int64
	DownPkts  int64
	DonwBytes int64
	MaxDSeq   int64
}

type VideoCounters struct {
	UpstreamChunks  []VideoSegment
	RunningUpstream VideoSegment
}

const (
	// QUICHeaderLen represents the minimum length in bytes to determine when a
	// QUIC upstream packet contains payload
	QUICHeaderLen = 100
)

// AddPacket updates the flow states based on the packet pkt
func (vf *VideoCounters) AddPacket(pkt *network.Packet) error {
	log.Debugf("Updating counter of type %s with packet of dir %d and length %d", vf.Type(), pkt.Dir, pkt.DataLength)
	if pkt.Dir == network.TrafficOut {
		if (pkt.IsTCP && pkt.DataLength > 0) || (!pkt.IsTCP && pkt.DataLength > QUICHeaderLen) {
			if vf.RunningUpstream.TsStart != 0 && vf.RunningUpstream.DownPkts > 0 {
				vf.RunningUpstream.TsEnd = vf.RunningUpstream.LastPkt
				vf.UpstreamChunks = append(vf.UpstreamChunks, vf.RunningUpstream)
			}
			vf.RunningUpstream = VideoSegment{Len: int64(pkt.Length), TsStart: pkt.TStamp, Seq: int64(pkt.Tcp.Seq)}
		}
	} else if pkt.DataLength > 0 {
		vf.RunningUpstream.DownPkts++
		vf.RunningUpstream.DonwBytes += int64(pkt.DataLength)

		if int64(pkt.Tcp.Seq) > vf.RunningUpstream.MaxDSeq {
			vf.RunningUpstream.MaxDSeq = int64(pkt.Tcp.Seq)
		}
		if pkt.TStamp > vf.RunningUpstream.TsEnd {
			vf.RunningUpstream.LastPkt = pkt.TStamp
		}
	}

	return nil
}

// Reset resets the flow statistics
func (vf *VideoCounters) Reset() error {
	vf.RunningUpstream = VideoSegment{}
	vf.UpstreamChunks = make([]VideoSegment, 0)
	return nil
}

// Clear the flow statistics
func (vf *VideoCounters) Clear() error {
	vf.UpstreamChunks = make([]VideoSegment, 0)
	return nil
}

// Type returns a string with the type name of the counter.
func (c *VideoCounters) Type() string {
	return "VideoCounters"
}

type VideoCountersOut struct {
	VideoSegments []VideoSegment
}

// Collect returns a []byte representation of the counter
func (c *VideoCounters) Collect() []byte {
	out := VideoCountersOut{
		VideoSegments: c.UpstreamChunks,
	}
	if c.RunningUpstream.TsStart > 0 {
		out.VideoSegments = append(out.VideoSegments, c.RunningUpstream)
	}
	b, _ := json.Marshal(out)
	return b
}
