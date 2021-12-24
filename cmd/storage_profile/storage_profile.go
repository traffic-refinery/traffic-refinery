package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/traffic-refinery/traffic-refinery/internal/config"
	"github.com/traffic-refinery/traffic-refinery/internal/flowstats"
	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/servicemap"
	"github.com/traffic-refinery/traffic-refinery/internal/stats"
)

func BenchmarkStorage(traceFile, folder, conf string) {
	var err error

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
			outJson := stats.OutJson{
				Version: "3.0",
				Conf:    "--",
				Type:    "Storage",
				TsStart: 0,
				TsEnd:   0,
				Data:    nil,
			}

			trafficData := flowcache.DumpToString()

			data, _ := json.Marshal(trafficData)
			outJson.Data = data
			data, _ = json.Marshal(outJson)
			j++
			lastTick = lastTick + 10
			ioutil.WriteFile(folder+string(os.PathSeparator)+fmt.Sprintf("storprofile_%d.json", j), data, 0644)
		}
		flowcache.ProcessPacket(&nextPkt.Pkt)
	}
}

func main() {
	trace := flag.String("trace", "", "Pcap trace to use for profiling. If none provided it runs on a default one")
	conf := flag.String("conf", "", "Configuration file to use for parsing traffic. If none provided it runs on a default one")
	folder := flag.String("folder", "", "Temporary location where to place memprofile files. If none provided the current folder is used")
	flag.Parse()
	BenchmarkStorage(*trace, *folder, *conf)
}
