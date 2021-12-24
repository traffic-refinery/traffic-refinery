package counters

import (
	"testing"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestAddPacketVideoCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &VideoCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	if len(c.UpstreamChunks) != 0 {
		t.Fatalf("UpstreamChunks %d does not correspond to expected one %d", len(c.UpstreamChunks), 0)
	} else if c.RunningUpstream.DonwBytes != 882 {
		t.Fatalf("RunningUpstream.DonwBytes %d does not correspond to expected one %d", c.RunningUpstream.DonwBytes, 882)
	}
}

func TestClearVideoCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &VideoCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Clear()
	if len(c.UpstreamChunks) != 0 {
		t.Fatalf("UpstreamChunks %d does not correspond to expected one %d", len(c.UpstreamChunks), 0)
	}
}

func TestResetVideoCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &VideoCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Reset()
	if len(c.UpstreamChunks) != 0 {
		t.Fatalf("UpstreamChunks %d does not correspond to expected one %d", len(c.UpstreamChunks), 0)
	}
}

func TestCollectVideoCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &VideoCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	t.Logf("Counter representation: %s", c.Collect())
}
