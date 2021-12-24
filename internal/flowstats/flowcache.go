// Package flowstats implements the main functions to process and store flow
// statistics.
package flowstats

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/traffic-refinery/traffic-refinery/internal/cache"
	"github.com/traffic-refinery/traffic-refinery/internal/counters"
	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/servicemap"
)

// FlowCache is a cache used to store flows' statistics
type FlowCache struct {
	// Internal cache. Any Cache type can be used
	cache cache.Cache
	// DNS Cache for service type detection
	serviceMap *servicemap.ServiceMap
	// Whether to anonymize IP addresses or not
	anonymize bool
	// serviceIdToCountersId
	serviceIdToCountersId map[servicemap.ServiceID][]int
	// availableCounters
	availableCounters counters.AvailableCounters
}

// NewFlowCache initiates a new FlowCache.
// t specifies the cache type. Possible cache types:
//
// - "ConcurrentCacheMap": Concurrent Map with periodic eviction of expired
// entries. Currently the only supported cache type
//
// - "BigCache": (NOT IMPLEMENTED) modified version of https://github.com/allegro/bigcache
//
// - "CacheMap": (NOT IMPLEMENTED) a simple map with no concurrency support
//
// - "Map": (NOT IMPLEMENTED) a simple map with no concurrency support
func NewFlowCache(t string, serviceMap *servicemap.ServiceMap, evictTime, cleanupTime time.Duration, shardsCount uint32, anonymize bool) (*FlowCache, error) {
	ret := &FlowCache{}

	log.Debugf("Cache type selected %s", t)

	if strings.ToLower(t) == "concurrentcachemap" {
		ret.cache = cache.NewConcurrentCacheMap(shardsCount, evictTime, nil, cleanupTime)
	} else {
		return nil, errors.New("incorrect type for cache")
	}

	ret.serviceMap = serviceMap
	ret.anonymize = anonymize

	ret.serviceIdToCountersId = make(map[servicemap.ServiceID][]int)

	log.Debugln("Flowcache initialized correctly")
	return ret, nil
}

func (fc *FlowCache) AddServices(services []Service) error {
	slist := []string{}
	for _, service := range services {
		slist = append(slist, service.Collect...)
	}
	nameToID, err := fc.availableCounters.Build(slist)
	if err != nil {
		return err
	}

	for _, service := range services {
		if id, ok := fc.serviceMap.GetId(service.Name); ok {
			fc.serviceIdToCountersId[id] = []int{}
			for _, c := range service.Collect {
				fc.serviceIdToCountersId[id] = append(fc.serviceIdToCountersId[id], nameToID[c])
			}
		} else {
			return errors.New("can't find service " + service.Name)
		}
	}

	return nil
}

func (fc *FlowCache) addPacket(pkt *network.Packet, hash *string) error {
	if value, ok := fc.cache.GetAndLock(*hash); ok {
		log.Debugln("Packet already in the cache, processing service ip ", pkt.ServiceIP)
		flow, _ := value.(*Flow)
		flow.AddPacket(pkt)
		fc.cache.SetAndUnlock(*hash, flow)
	} else {
		//Query dns cache for the flow type
		if s, ok := fc.serviceMap.LookupIP(pkt.ServiceIP); ok {
			// TODO Assumes only first service match per IP is used
			sid := s[0]
			log.Debugln("Create new flow of service type ", sid, " for service ip ", pkt.ServiceIP)
			if service, found := fc.serviceMap.GetService(sid); found {
				flow := CreateFlow()
				flow.Id = *hash
				flow.Service = service.Name
				flow.DomainName = ""
				flow.ServiceIP = pkt.ServiceIP
				flow.LocalIP = pkt.MyIP
				if pkt.IsTCP {
					flow.Protocol = "tcp"
				} else {
					flow.Protocol = "udp"
				}
				flow.LocalPort = strconv.Itoa(int(pkt.MyPort))
				flow.ServicePort = strconv.Itoa(int(pkt.ServicePort))
				for _, counter := range fc.serviceIdToCountersId[sid] {
					instance, _ := fc.availableCounters.InstantiateById(counter)
					flow.Cntrs = append(flow.Cntrs, instance)
				}
				flow.Reset()
				flow.AddPacket(pkt)
				fc.cache.Set(*hash, flow)
			}
		} else {
			log.Debugln("IP ", pkt.ServiceIP, " does not belong to a known service")
		}
	}
	return nil
}

// ProcessPacket processes incoming packets. If the flow is already in the cache, it updates
// its counters. If not, it creates it based on the DNS type and inserts it into
// the cache.
func (fc *FlowCache) ProcessPacket(pkt *network.Packet) error {
	if fc.anonymize {
		var testKey = []byte{45, 148, 31, 183, 121, 99, 98, 199, 103, 48, 199, 151, 176, 128, 82, 175, 33, 228, 17, 204, 122, 199, 124, 65, 130, 80, 120, 210, 81, 207, 169, 48}
		cpan, _ := network.NewCryptoPAn(testKey)

		var obfsaddr = cpan.Anonymize(net.ParseIP(pkt.MyIP))
		pkt.MyIP = obfsaddr.String()
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d-%d", pkt.ServiceIP, pkt.MyIP, pkt.ServicePort, pkt.MyPort))))
	log.Debugf("Received packet for flow %s", hash)

	return fc.addPacket(pkt, &hash)

}

// Dump copies the entire cache int a map.
func (fc *FlowCache) Dump() map[string]Flow {
	log.Debugln("Dumping the flow cache into a map")
	if i, err := fc.cache.IterativeDump(); err != nil {
		log.Errorln(err)
		return nil
	} else {
		ret := make(map[string]Flow)
		for {
			v, err := fc.cache.NextElement(i)
			if err != nil {
				log.Errorln(err)
				break
			} else if v == nil {
				break
			} else {
				f := v.(*Flow)
				ret[f.Id] = *f
				f.Clear()
			}
		}
		return ret
	}
}

// DumpToChannel copies the entire cache int a channel, entry by entry.
func (fc *FlowCache) DumpToChannel(c chan Flow) {
	if i, err := fc.cache.IterativeDump(); err != nil {
		log.Errorln(err)
		return
	} else {
		for {
			v, err := fc.cache.NextElement(i)
			if err != nil {
				break
			} else {
				if v == nil {
					close(c)
				}
				f := v.(*Flow)
				c <- *f
				f.Clear()
			}
		}
	}
}

// Dump copies the entire cache int a map.
func (fc *FlowCache) DumpToString() []json.RawMessage {
	log.Debugln("Dumping the flow cache into a map")
	flows := []json.RawMessage{}
	if i, err := fc.cache.IterativeDump(); err != nil {
		log.Errorln(err)
		return nil
	} else {
		for {
			v, err := fc.cache.NextElement(i)
			if err != nil {
				log.Errorln(err)
				break
			} else if v == nil {
				break
			} else {
				f := v.(*Flow)
				flows = append(flows, f.Collect())
				f.Clear()
			}
		}
	}
	return flows
}
