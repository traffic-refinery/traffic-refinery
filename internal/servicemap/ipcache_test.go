package servicemap

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/traffic-refinery/traffic-refinery/internal/config"
	"github.com/traffic-refinery/traffic-refinery/internal/utils"
)

func TestIPCacheExpire(t *testing.T) {
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
				TTL: 1,
			},
		},
		Questions: []layers.DNSQuestion{
			{
				Name: []byte("api-global.netflix.com"),
			},
		},
	})

	time.Sleep(2 * time.Second)

	if _, found := smap.LookupIP("1.1.1.1"); found {
		t.Fatalf("IP 1.1.1.1 should not be in the ip cache\n")
	}
}
