package flowstats

import (
	"testing"

	"github.com/traffic-refinery/traffic-refinery/internal/counters"
	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestAddPacketFlow(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	f := CreateFlow()
	f.Cntrs = append(f.Cntrs, &counters.PacketCounters{})
	f.Reset()
	for _, pkt := range trace.Trace {
		f.AddPacket(&pkt.Pkt)
	}
	if f.Cntrs[0].(*counters.PacketCounters).InCounter != 3 {
		t.Fatalf("InCounter %d does not correspond to expected one %d", f.Cntrs[0].(*counters.PacketCounters).InCounter, 3)
	} else if f.Cntrs[0].(*counters.PacketCounters).OutCounter != 4 {
		t.Fatalf("OutCounter %d does not correspond to expected one %d", f.Cntrs[0].(*counters.PacketCounters).OutCounter, 4)
	}
}

func TestClearFlow(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	f := CreateFlow()
	f.Cntrs = append(f.Cntrs, &counters.PacketCounters{})
	f.Reset()
	for _, pkt := range trace.Trace {
		f.AddPacket(&pkt.Pkt)
	}
	f.Clear()
	if f.Cntrs[0].(*counters.PacketCounters).InCounter != 0 {
		t.Fatalf("InCounter %d does not correspond to expected one %d", f.Cntrs[0].(*counters.PacketCounters).InCounter, 0)
	} else if f.Cntrs[0].(*counters.PacketCounters).OutCounter != 0 {
		t.Fatalf("OutCounter %d does not correspond to expected one %d", f.Cntrs[0].(*counters.PacketCounters).OutCounter, 0)
	}
}

func TestResetFlow(t *testing.T) {
	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	f := CreateFlow()
	f.Cntrs = append(f.Cntrs, &counters.PacketCounters{})
	f.Reset()
	for _, pkt := range trace.Trace {
		f.AddPacket(&pkt.Pkt)
	}
	f.Reset()
	if f.Cntrs[0].(*counters.PacketCounters).InCounter != 0 {
		t.Fatalf("InCounter %d does not correspond to expected one %d", f.Cntrs[0].(*counters.PacketCounters).InCounter, 0)
	} else if f.Cntrs[0].(*counters.PacketCounters).OutCounter != 0 {
		t.Fatalf("OutCounter %d does not correspond to expected one %d", f.Cntrs[0].(*counters.PacketCounters).OutCounter, 0)
	}
}

func TestCollectFlow(t *testing.T) {

	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")
	f := CreateFlow()
	f.Cntrs = append(f.Cntrs, &counters.PacketCounters{})
	f.Reset()
	for _, pkt := range trace.Trace {
		f.AddPacket(&pkt.Pkt)
	}
	t.Logf("Flow representation: %s", f.Collect())
}
