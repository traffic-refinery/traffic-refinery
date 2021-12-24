package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/traffic-refinery/traffic-refinery/internal/config"
	"github.com/traffic-refinery/traffic-refinery/internal/flowstats"
	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/servicemap"
)

func BenchmarkMemory(traceFile, folder, conf string) {
	var err error

	runtime.MemProfileRate = 1

	// Prepare the service map and cache
	c := config.TrafficRefineryConfig{}
	c.ImportConfigFromFile(conf)
	smapServices := []servicemap.Service{}
	fcacheServices := []flowstats.Service{}
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
		fcacheServices = append(fcacheServices, flowstats.Service{
			Name:    s.Name,
			Collect: s.Collect,
		})
	}

	var smap *servicemap.ServiceMap
	if smap, err = servicemap.NewServiceMap(c.DNSCache.EvictTime, c.DNSCache.CleanupTime); err == nil {
		panic(err)
	}
	smap.ConfigServiceMap(smapServices)

	flowcache, err := flowstats.NewFlowCache(c.FlowCache.CacheType, smap, c.FlowCache.EvictTime, c.FlowCache.CleanupTime, uint32(c.FlowCache.ShardsCount), c.FlowCache.Anonymize)
	if err != nil {
		panic(err)
	}
	flowcache.AddServices(fcacheServices)

	if err != nil {
		log.Fatal("Flowcache error: ", err)
	}

	trace := network.GetTraceWithServices(traceFile, smap)

	N := int(trace.Count)
	lastTick := trace.Trace[0].Pkt.TStamp / int64(1000000000)
	j := 0
	for i := 0; i < N; i++ {
		nextPkt := trace.Trace[i]
		newTick := nextPkt.Pkt.TStamp / int64(1000000000)
		if newTick-lastTick > 10 {
			memf, err := os.Create(folder + string(os.PathSeparator) + fmt.Sprintf("memprofile_%d.out", j))
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			runtime.GC() // get up-to-date heap statistics
			if err := pprof.WriteHeapProfile(memf); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			memf.Close()
			j++
			lastTick = lastTick + 10
			flowcache.Dump()
		}
		flowcache.ProcessPacket(&nextPkt.Pkt)
	}
}

func main() {
	trace := flag.String("trace", "", "Pcap trace to use for profiling. If none provided it runs on a default one")
	conf := flag.String("conf", "", "Configuration file to use for parsing traffic. If none provided it runs on a default one")
	folder := flag.String("folder", "", "Temporary location where to place memprofile files. If none provided the current folder is used")
	flag.Parse()
	BenchmarkMemory(*trace, *folder, *conf)
}
