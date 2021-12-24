package flowstats

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/traffic-refinery/traffic-refinery/internal/config"
	"github.com/traffic-refinery/traffic-refinery/internal/counters"
	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/servicemap"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestFlowcacheAddFlow(t *testing.T) {
	var err error
	// Load test configuration
	conf := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_simple.json"
	conf.ImportConfigFromFile(testConfig)

	smapServices := []servicemap.Service{}
	fcacheServices := []Service{}
	for i, s := range conf.Services {
		smapServices = append(smapServices, servicemap.Service{
			Name: s.Name,
			ServiceFilter: servicemap.Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: servicemap.ServiceID(i),
		})
		fcacheServices = append(fcacheServices, Service{
			Name:    s.Name,
			Collect: s.Collect,
		})
	}

	var smap *servicemap.ServiceMap
	if smap, err = servicemap.NewServiceMap(conf.DNSCache.EvictTime, conf.DNSCache.CleanupTime); err != nil {
		panic(err)
	}
	smap.ConfigServiceMap(smapServices)

	flowcache, err := NewFlowCache(conf.FlowCache.CacheType, smap, conf.FlowCache.EvictTime, conf.FlowCache.CleanupTime, uint32(conf.FlowCache.ShardsCount), conf.FlowCache.Anonymize)
	if err != nil {
		panic(err)
	}

	err = flowcache.AddServices(fcacheServices)
	if err != nil {
		t.Fatalf("Can not initialize flowcache services: %s", err)
	}

	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")

	for _, pkt := range trace.Trace {
		flowcache.ProcessPacket(&pkt.Pkt)
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d-%d", "198.38.120.133", "192.168.43.72", 443, 51751))))
	if value, ok := flowcache.cache.GetAndLock(hash); ok {
		flow, ok := value.(*Flow)
		if !ok {
			t.Fatalf("It's not a flow pointer")
		}
		c, ok := flow.Cntrs[0].(*counters.PacketCounters)
		if !ok {
			t.Fatalf("It's not a PacketCounters")
		}
		if c.InCounter != 3 {
			t.Fatalf("InCounter %d does not correspond to expected one %d", c.InCounter, 3)
		} else if c.OutCounter != 4 {
			t.Fatalf("OutCounter %d does not correspond to expected one %d", c.OutCounter, 4)
		}
	} else {
		b, _ := json.Marshal(flowcache.Dump())
		t.Fatalf("Flow not found in cache:\n%s\n", b)
	}
}

func TestFlowcacheFlowExpire(t *testing.T) {
	var err error
	// Load test configuration
	conf := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_simple.json"
	conf.ImportConfigFromFile(testConfig)

	smapServices := []servicemap.Service{}
	fcacheServices := []Service{}
	for i, s := range conf.Services {
		smapServices = append(smapServices, servicemap.Service{
			Name: s.Name,
			ServiceFilter: servicemap.Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: servicemap.ServiceID(i),
		})
		fcacheServices = append(fcacheServices, Service{
			Name:    s.Name,
			Collect: s.Collect,
		})
	}

	var smap *servicemap.ServiceMap
	if smap, err = servicemap.NewServiceMap(conf.DNSCache.EvictTime, conf.DNSCache.CleanupTime); err != nil {
		panic(err)
	}
	smap.ConfigServiceMap(smapServices)

	flowcache, err := NewFlowCache(conf.FlowCache.CacheType, smap, 2*time.Second, 3*time.Second, uint32(conf.FlowCache.ShardsCount), conf.FlowCache.Anonymize)
	if err != nil {
		panic(err)
	}

	err = flowcache.AddServices(fcacheServices)
	if err != nil {
		t.Fatalf("Can not initialize flowcache services: %s", err)
	}

	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/short_test.pcap")

	for _, pkt := range trace.Trace {
		flowcache.ProcessPacket(&pkt.Pkt)
	}
	time.Sleep(4 * time.Second)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d-%d", "198.38.120.133", "192.168.43.72", 443, 51751))))
	if _, ok := flowcache.cache.GetAndLock(hash); ok {
		t.Fatalf("Flow %s should not be in  thecache:\n", hash)
	}
}

func TestFlowcacheAddFlows(t *testing.T) {
	var err error
	// Load test configuration
	conf := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_simple.json"
	conf.ImportConfigFromFile(testConfig)

	smapServices := []servicemap.Service{}
	fcacheServices := []Service{}
	for i, s := range conf.Services {
		smapServices = append(smapServices, servicemap.Service{
			Name: s.Name,
			ServiceFilter: servicemap.Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: servicemap.ServiceID(i),
		})
		fcacheServices = append(fcacheServices, Service{
			Name:    s.Name,
			Collect: s.Collect,
		})
	}

	var smap *servicemap.ServiceMap
	if smap, err = servicemap.NewServiceMap(conf.DNSCache.EvictTime, conf.DNSCache.CleanupTime); err != nil {
		panic(err)
	}
	smap.ConfigServiceMap(smapServices)

	flowcache, err := NewFlowCache(conf.FlowCache.CacheType, smap, conf.FlowCache.EvictTime, conf.FlowCache.CleanupTime, uint32(conf.FlowCache.ShardsCount), conf.FlowCache.Anonymize)
	if err != nil {
		panic(err)
	}

	err = flowcache.AddServices(fcacheServices)
	if err != nil {
		t.Fatalf("Can not initialize flowcache services: %s", err)
	}

	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/video_trace.pcap")

	for _, pkt := range trace.Trace {
		flowcache.ProcessPacket(&pkt.Pkt)
	}
	d := flowcache.DumpToString()
	if len(d) != 62 {
		t.Fatalf("The number of collected flows is incorrect: %d", len(d))
	}
}

func TestFlowcacheDump(t *testing.T) {
	var err error
	// Load test configuration
	conf := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_simple.json"
	conf.ImportConfigFromFile(testConfig)

	smapServices := []servicemap.Service{}
	fcacheServices := []Service{}
	for i, s := range conf.Services {
		smapServices = append(smapServices, servicemap.Service{
			Name: s.Name,
			ServiceFilter: servicemap.Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: servicemap.ServiceID(i),
		})
		fcacheServices = append(fcacheServices, Service{
			Name:    s.Name,
			Collect: s.Collect,
		})
	}

	var smap *servicemap.ServiceMap
	if smap, err = servicemap.NewServiceMap(conf.DNSCache.EvictTime, conf.DNSCache.CleanupTime); err != nil {
		panic(err)
	}
	smap.ConfigServiceMap(smapServices)

	flowcache, err := NewFlowCache(conf.FlowCache.CacheType, smap, conf.FlowCache.EvictTime, conf.FlowCache.CleanupTime, uint32(conf.FlowCache.ShardsCount), conf.FlowCache.Anonymize)
	if err != nil {
		panic(err)
	}

	err = flowcache.AddServices(fcacheServices)
	if err != nil {
		t.Fatalf("Can not initialize flowcache services: %s", err)
	}

	trace := network.GetTrace(utils.GetRepoPath() + "/test/traffic_data/video_trace.pcap")

	for _, pkt := range trace.Trace {
		flowcache.ProcessPacket(&pkt.Pkt)
	}
	b, _ := json.Marshal(flowcache.DumpToString())
	t.Logf("Cache dump:\n%s\n", b)
}

func BenchmarkFlowcacheHash(b *testing.B) {
	trace := network.GetRandomTrace(b.N, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pkt := &trace.Trace[i].Pkt
		_ = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d-%d", pkt.ServiceIP, pkt.MyIP, pkt.ServicePort, pkt.MyPort))))
	}
}
