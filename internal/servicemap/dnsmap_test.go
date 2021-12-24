package servicemap

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/google/gopacket/layers"
	"github.com/traffic-refinery/traffic-refinery/internal/config"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestDNSCacheInsert(t *testing.T) {
	// Load test configuration
	c := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_video.json"
	c.ImportConfigFromFile(testConfig)

	smapServices := []Service{}
	for i, s := range c.Services {
		smapServices = append(smapServices, Service{
			Name: s.Name,
			ServiceFilter: Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: ServiceID(i),
		})
	}
	b, _ := json.Marshal(smapServices)
	t.Logf("Services to use for test:\n%s\n", b)

	var smap *ServiceMap
	var err error

	if smap, err = NewServiceMap(c.DNSCache.EvictTime, c.DNSCache.CleanupTime); err != nil {
		t.Fatalf("Fatal error in creating service map %s", err)
	}
	smap.ConfigServiceMap(smapServices)

	smap.ParseDNSResponse(layers.DNS{
		Answers: []layers.DNSResourceRecord{
			{
				IP:  net.ParseIP("1.1.1.1"),
				TTL: 1000000,
			},
		},
		Questions: []layers.DNSQuestion{
			{
				Name: []byte("api-global.netflix.com"),
			},
		},
	})

	if ids, found := smap.LookupIP("1.1.1.1"); found {
		if name, found := smap.GetName(ids[0]); !found || name != "Netflix" {
			id, _ := smap.GetId("Netflix")
			t.Fatalf("IP 1.1.1.1 should be netflix and instead is %s [%d] [%d]\n", name, ids[0], id)
		}
	} else {
		t.Fatalf("IP 1.1.1.1 is not found in the dns map\n")
	}
}

func TestDNSCacheNotFound(t *testing.T) {
	// Load test configuration
	c := config.TrafficRefineryConfig{}
	testConfig := utils.GetRepoPath() + "/test/config/trconfig_video.json"
	c.ImportConfigFromFile(testConfig)

	smapServices := []Service{}
	for i, s := range c.Services {
		smapServices = append(smapServices, Service{
			Name: s.Name,
			ServiceFilter: Filter{
				DomainsString: s.Filter.DomainsString,
				DomainsRegex:  s.Filter.DomainsRegex,
				Prefixes:      s.Filter.Prefixes,
			},
			Code: ServiceID(i),
		})
	}
	b, _ := json.Marshal(smapServices)
	t.Logf("Services to use for test:\n%s\n", b)

	var smap *ServiceMap
	var err error

	if smap, err = NewServiceMap(c.DNSCache.EvictTime, c.DNSCache.CleanupTime); err != nil {
		t.Fatalf("Fatal error in creating service map %s", err)
	}
	smap.ConfigServiceMap(smapServices)

	smap.ParseDNSResponse(layers.DNS{
		Answers: []layers.DNSResourceRecord{
			{
				IP:  net.ParseIP("1.1.1.1"),
				TTL: 1000000,
			},
		},
		Questions: []layers.DNSQuestion{
			{
				Name: []byte("api-global.netflix.com"),
			},
		},
	})

	if _, found := smap.LookupIP("1.1.1.2"); found {
		t.Fatalf("IP 1.1.1.2 should not be in the dns map\n")
	}
}
