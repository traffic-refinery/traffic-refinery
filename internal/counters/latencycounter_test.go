package counters

import (
	"testing"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestAddPacketLatencyJitterCounter(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &LatencyJitterCounter{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	if c.RTT.N != 2 {
		t.Fatalf("Number of RTTs %f does not correspond to expected one %d", c.RTT.N, 2)
	}
}

func TestClearLatencyJitterCounter(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &LatencyJitterCounter{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Clear()
	if c.RTT.N != 0 {
		t.Fatalf("Number of RTTs %f does not correspond to expected one %d", c.RTT.N, 0)
	}
}

func TestResetLatencyJitterCounter(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &LatencyJitterCounter{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Reset()
	if c.RTT.N != 0 {
		t.Fatalf("Number of RTTs %f does not correspond to expected one %d", c.RTT.N, 0)
	}
}

func TestCollectLatencyJitterCounter(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &LatencyJitterCounter{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	t.Logf("Counter representation: %s", c.Collect())
}
