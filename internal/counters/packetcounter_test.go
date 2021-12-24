package counters

import (
	"testing"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestAddPacketPacketCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &PacketCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	if c.InCounter != 3 {
		t.Fatalf("InCounter %d does not correspond to expected one %d", c.InCounter, 3)
	} else if c.OutCounter != 4 {
		t.Fatalf("OutCounter %d does not correspond to expected one %d", c.OutCounter, 4)
	}
}

func TestClearPacketCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &PacketCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Clear()
	if c.InCounter != 0 {
		t.Fatalf("InCounter %d does not correspond to expected one %d", c.InCounter, 0)
	} else if c.OutCounter != 0 {
		t.Fatalf("OutCounter %d does not correspond to expected one %d", c.OutCounter, 0)
	}
}

func TestResetPacketCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &PacketCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Reset()
	if c.InCounter != 0 {
		t.Fatalf("InCounter %d does not correspond to expected one %d", c.InCounter, 0)
	} else if c.OutCounter != 0 {
		t.Fatalf("OutCounter %d does not correspond to expected one %d", c.OutCounter, 0)
	}
}

func TestCollectPacketCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &PacketCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	t.Logf("Counter representation: %s", c.Collect())
}
