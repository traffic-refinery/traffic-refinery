package counters

import (
	"testing"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestAddPacketPNGCopyCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &PNGCopyCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	if c.CopiedBytes != 400 {
		t.Fatalf("Number of copied bytes %d does not correspond to expected one %d", c.CopiedBytes, 400)
	} else if c.StoredBytes != 400 {
		t.Fatalf("Number of copied bytes %d does not correspond to expected one %d", c.StoredBytes, 400)
	}
}

func TestClearPNGCopyCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &PNGCopyCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Clear()
	if c.CopiedBytes != 400 {
		t.Fatalf("Number of copied bytes %d does not correspond to expected one %d", c.CopiedBytes, 400)
	} else if c.StoredBytes != 0 {
		t.Fatalf("Number of copied bytes %d does not correspond to expected one %d", c.StoredBytes, 0)
	}
}

func TestResetPNGCopyCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &PNGCopyCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Reset()
	if c.CopiedBytes != 0 {
		t.Fatalf("Number of copied bytes %d does not correspond to expected one %d", c.CopiedBytes, 0)
	} else if c.StoredBytes != 0 {
		t.Fatalf("Number of copied bytes %d does not correspond to expected one %d", c.StoredBytes, 0)
	}
}

func TestCollectPNGCopyCounters(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &PNGCopyCounters{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	t.Logf("Counter representation: %s", c.Collect())
}
