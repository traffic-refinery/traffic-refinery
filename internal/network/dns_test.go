package network

import (
	// "errors"
	// "github.com/stretchr/testify/assert"
	"encoding/json"
	"flag"
	"testing"

	"github.com/traffic-refinery/traffic-refinery/internal/config"
	"github.com/traffic-refinery/traffic-refinery/internal/servicemap"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

var ifname = flag.String("ifname", "", "Name of the interface to use")

func TestDNSParser(t *testing.T) {
	// TODO Run live on interface

}

func TestDNSCache(t *testing.T) {
	// Load test configuration
	c := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_video.json"
	c.ImportConfigFromFile(testConfig)
	b, _ := json.Marshal(c)
	t.Logf("Configuration for test:\n%sn", b)

	smapServices := []servicemap.Service{}
	for i, s := range c.Services {
		smapServices = append(smapServices, servicemap.Service{
			Name: s.Name,
			ServiceFilter: servicemap.Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: servicemap.ServiceID(i),
		})
	}
	b, _ = json.Marshal(smapServices)
	t.Logf("Services to use for test:\n%s\n", b)

	var smap *servicemap.ServiceMap
	var err error

	if smap, err = servicemap.NewServiceMap(c.DNSCache.EvictTime, c.DNSCache.CleanupTime); err != nil {
		t.Fatalf("Fatal error in creating service map %s", err)
	}
	smap.ConfigServiceMap(smapServices)

	// Create Query
	dnsTrace := GetDNSTrace(utils.GetRepoPath() + "/test/traffic_data/dns.pcap")
	for i := 0; int64(i) < dnsTrace.Count; i++ {
		smap.ParseDNSResponse(*dnsTrace.Trace[i].Data)
	}

	if ids, found := smap.LookupIP("52.17.164.26"); found {
		if name, found := smap.GetName(ids[0]); !found || name != "Netflix" {
			id, _ := smap.GetId("Netflix")
			t.Fatalf("IP 52.17.164.26 should be netflix and instead is %s [%d] [%d]\n", name, ids[0], id)
		}
	} else {
		t.Fatalf("IP 52.17.164.26 is not found in the dns map\n")
	}
}

func BenchmarkDNSCacheInsertsVideo(b *testing.B) {
	// Load test configuration
	c := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_video.json"
	c.ImportConfigFromFile(testConfig)

	smapServices := []servicemap.Service{}
	for i, s := range c.Services {
		smapServices = append(smapServices, servicemap.Service{
			Name: s.Name,
			ServiceFilter: servicemap.Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: servicemap.ServiceID(i),
		})
	}

	var smap *servicemap.ServiceMap
	var err error

	if smap, err = servicemap.NewServiceMap(c.DNSCache.EvictTime, c.DNSCache.CleanupTime); err != nil {
		b.Fatalf("Fatal error in creating service map %s", err)
	}
	smap.ConfigServiceMap(smapServices)

	// Create Query
	dnsTrace := GetDNSTrace(utils.GetRepoPath() + "/test/traffic_data/dns.pcap")
	for dnsTrace.Count < int64(b.N) {
		for i := 0; int64(i) < dnsTrace.Count; i++ {
			dnsTrace.Trace = append(dnsTrace.Trace, dnsTrace.Trace[i])
		}
		dnsTrace.Count += dnsTrace.Count
	}

	// Test Query
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		smap.ParseDNSResponse(*dnsTrace.Trace[i].Data)
	}
}

func BenchmarkDNSCacheInsertsAds(b *testing.B) {
	// Load test configuration
	c := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_ads.json"
	c.ImportConfigFromFile(testConfig)

	smapServices := []servicemap.Service{}
	for i, s := range c.Services {
		smapServices = append(smapServices, servicemap.Service{
			Name: s.Name,
			ServiceFilter: servicemap.Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: servicemap.ServiceID(i),
		})
	}

	var smap *servicemap.ServiceMap
	var err error

	if smap, err = servicemap.NewServiceMap(c.DNSCache.EvictTime, c.DNSCache.CleanupTime); err != nil {
		b.Fatalf("Fatal error in creating service map %s", err)
	}
	smap.ConfigServiceMap(smapServices)

	// Create Query
	dnsTrace := GetDNSTrace(utils.GetRepoPath() + "/test/traffic_data/dns.pcap")
	for dnsTrace.Count < int64(b.N) {
		for i := 0; int64(i) < dnsTrace.Count; i++ {
			dnsTrace.Trace = append(dnsTrace.Trace, dnsTrace.Trace[i])
		}
		dnsTrace.Count += dnsTrace.Count
	}

	// Test Query
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		smap.ParseDNSResponse(*dnsTrace.Trace[i].Data)
	}
}

func BenchmarkDNSCacheLookups(b *testing.B) {
	// Load test configuration
	c := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_default.json"
	c.ImportConfigFromFile(testConfig)

	smapServices := []servicemap.Service{}
	for i, s := range c.Services {
		smapServices = append(smapServices, servicemap.Service{
			Name: s.Name,
			ServiceFilter: servicemap.Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: servicemap.ServiceID(i),
		})
	}

	var smap *servicemap.ServiceMap
	var err error

	if smap, err = servicemap.NewServiceMap(c.DNSCache.EvictTime, c.DNSCache.CleanupTime); err != nil {
		b.Fatalf("Fatal error in creating service map %s", err)
	}
	smap.ConfigServiceMap(smapServices)

	// Create Query
	dnsTrace := GetDNSTrace(utils.GetRepoPath() + "/test/traffic_data/dns.pcap")
	for dnsTrace.Count < int64(b.N) {
		for i := 0; int64(i) < dnsTrace.Count; i++ {
			dnsTrace.Trace = append(dnsTrace.Trace, dnsTrace.Trace[i])
		}
		dnsTrace.Count += dnsTrace.Count
	}

	// Test Query
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		smap.ParseDNSResponse(*dnsTrace.Trace[i].Data)
	}

	ips := utils.GetStringLines(utils.GetRepoPath() + "/test/traffic_data/dns.txt")
	for len(ips) < b.N {
		l := len(ips)
		for i := 0; i < l; i++ {
			ips = append(ips, ips[i])
		}
	}

	// Test Query
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = smap.LookupIP(ips[i])
	}
}
