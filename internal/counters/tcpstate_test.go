package counters

import (
	"testing"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestAddPacketTCPState(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &TCPState{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	if c.AckUpCounter != 4 {
		t.Fatalf("AckUpCounter %d does not correspond to expected one %d", c.AckUpCounter, 4)
	}
}

func TestClearTCPState(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &TCPState{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Clear()
	if c.AckUpCounter != 0 {
		t.Fatalf("AckUpCounter %d does not correspond to expected one %d", c.AckUpCounter, 0)
	}
}

func TestResetTCPState(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &TCPState{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	c.Reset()
	if c.AckUpCounter != 0 {
		t.Fatalf("AckUpCounter %d does not correspond to expected one %d", c.AckUpCounter, 0)
	}
}

func TestCollectTCPState(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	c := &TCPState{}
	c.Reset()
	for _, pkt := range trace.Trace {
		c.AddPacket(&pkt.Pkt)
	}
	t.Logf("Counter representation: %s", c.Collect())
}
