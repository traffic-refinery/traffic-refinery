package main

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"runtime"
	"runtime/pprof"

	"github.com/traffic-refinery/traffic-refinery/internal/config"
	"github.com/traffic-refinery/traffic-refinery/internal/flowstats"
	"github.com/traffic-refinery/traffic-refinery/internal/network"
	"github.com/traffic-refinery/traffic-refinery/internal/servicemap"
	"github.com/traffic-refinery/traffic-refinery/internal/stats"
)

const (
	// Version is the version number of the system
	// MAKE SURE TO INCREMENT AFTER EVERY CHANGE!
	Version = "2.0"
)

func loadConfig() config.TrafficRefineryConfig {
	fname := flag.String("conf", "", "Configuration file to load. If none is provided it looks for trconfig.json in ./ and /etc/traffic_refinery/")
	outFolder := flag.String("out", "", "Folder into which store output files. Defaults to /tmp/")
	hw := flag.String("hw", "", "Replaces hardware address in replay mode")
	cpu := flag.Bool("cpu", false, "Activate CPU profiling")
	mem := flag.Bool("mem", false, "Activate memory profiling")
	debug := flag.Bool("debug", false, "Log at debug level")
	info := flag.Bool("info", false, "Log at info level")
	warn := flag.Bool("warn", false, "Log at warn level")
	error := flag.Bool("error", false, "Log at error level")
	fatal := flag.Bool("fatal", false, "Log at fatal level")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	} else if *info {
		log.SetLevel(log.InfoLevel)
	} else if *warn {
		log.SetLevel(log.WarnLevel)
	} else if *error {
		log.SetLevel(log.ErrorLevel)
	} else if *fatal {
		log.SetLevel(log.FatalLevel)
	} else {
		log.SetLevel(log.FatalLevel)
	}

	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	conf := config.TrafficRefineryConfig{}

	if *fname != "" {
		conf.ImportConfigFromFile(*fname)
	} else {
		conf.ImportConfig()
	}

	if *cpu {
		conf.Sys.CPUProf = true
	}

	if *mem {
		conf.Sys.MemProf = true
	}

	if *outFolder != "" {
		conf.Sys.OutFolder = *outFolder
	}

	if *hw != "" {
		if conf.Parsers.DNSParser.Mode != "replay" {
			log.Warn("-hw option not used as the system is not run in replay mode")
		}
		conf.Parsers.DNSParser.ReplayMAC = *hw
		for i := 0; i < len(conf.Parsers.TrafficParsers); i++ {
			if conf.Parsers.TrafficParsers[i].Mode != "replay" {
				log.Warn("-hw option not used as the system is not run in replay mode")
			}
			conf.Parsers.TrafficParsers[i].ReplayMAC = *hw
		}
	}

	return conf
}

func main() {
	var err error
	var outb []byte
	conf := loadConfig()

	outb, _ = json.Marshal(conf)
	log.Infof("Running traffic refinery with configuration:\n%s\n", outb)

	// Prepare profiling if needed
	cpufname := ""
	memfname := ""
	var memf *os.File
	if conf.Sys.CPUProf {
		cpufname = path.Join(conf.Sys.OutFolder, "cpuprof.out")
		cpuf, err := os.Create(cpufname)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err = pprof.StartCPUProfile(cpuf); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer cpuf.Close()
		defer pprof.StopCPUProfile()
	}
	if conf.Sys.MemProf {
		memfname = path.Join(conf.Sys.OutFolder, "memprof.out")
		memf, err = os.Create(memfname)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
	}

	smapServices := []servicemap.Service{}
	fcacheServices := []flowstats.Service{}
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
		fcacheServices = append(fcacheServices, flowstats.Service{
			Name:    s.Name,
			Collect: s.Collect,
		})
	}

	var smap *servicemap.ServiceMap
	if smap, err = servicemap.NewServiceMap(conf.DNSCache.EvictTime, conf.DNSCache.CleanupTime); err != nil {
		panic(err)
	}
	smap.ConfigServiceMap(smapServices)

	log.Infof("Running the DNS parser on interface %s", conf.Parsers.DNSParser.Ifname)

	dnsni := new(network.NetworkInterface)
	ifconf := network.NetworkInterfaceConfiguration{
		Driver:    conf.Parsers.DNSParser.Driver,
		Name:      conf.Parsers.DNSParser.Ifname,
		Mode:      conf.Parsers.DNSParser.Mode,
		Filter:    network.DNSFilter,
		SnapLen:   1500,
		Clustered: conf.Parsers.DNSParser.Clustered,
		ClusterID: conf.Parsers.DNSParser.ClusterID,
		Replay:    conf.Parsers.DNSParser.Replay,
		ReplayMAC: conf.Parsers.DNSParser.ReplayMAC,
		ZeroCopy:  conf.Parsers.DNSParser.ZeroCopy,
		FanOut:    conf.Parsers.DNSParser.FanOut,
	}
	dnsni.NewNetworkInterface(ifconf)

	dp := new(network.DNSParser)
	dp.NewDNSParser(dnsni, smap)

	stop := make(chan struct{})
	go dp.Parse(nil, stop)

	flowcache, err := flowstats.NewFlowCache(conf.FlowCache.CacheType, smap, conf.FlowCache.EvictTime, conf.FlowCache.CleanupTime, uint32(conf.FlowCache.ShardsCount), conf.FlowCache.Anonymize)
	if err != nil {
		panic(err)
	}
	flowcache.AddServices(fcacheServices)

	log.Debugf("Initializing %d parsers", len(conf.Parsers.TrafficParsers))
	interfaces := []*network.NetworkInterface{}
	for i := 0; i < len(conf.Parsers.TrafficParsers); i++ {
		// In case no value was assigned to the replicas entry, assumes it's 1
		if conf.Parsers.TrafficParsers[i].Replicas == 0 {
			conf.Parsers.TrafficParsers[i].Replicas = 1
		}
		// TODO double check how replicas are supposed to work
		for j := 0; j < conf.Parsers.TrafficParsers[i].Replicas; j++ {
			log.Infof("Running traffic parser %d on interface %s", i+j, conf.Parsers.TrafficParsers[i].Ifname)
			trafficni := new(network.NetworkInterface)
			// Prepare the conf struct
			ifconf := network.NetworkInterfaceConfiguration{
				Driver:    conf.Parsers.TrafficParsers[i].Driver,
				Name:      conf.Parsers.TrafficParsers[i].Ifname,
				Mode:      conf.Parsers.TrafficParsers[i].Mode,
				Filter:    network.NotDNSFilter,
				SnapLen:   1500,
				Clustered: conf.Parsers.TrafficParsers[i].Clustered,
				ClusterID: conf.Parsers.TrafficParsers[i].ClusterID,
				Replay:    conf.Parsers.TrafficParsers[i].Replay,
				ReplayMAC: conf.Parsers.TrafficParsers[i].ReplayMAC,
				ZeroCopy:  conf.Parsers.TrafficParsers[i].ZeroCopy,
				FanOut:    conf.Parsers.TrafficParsers[i].FanOut,
			}
			// Create interface
			trafficni.NewNetworkInterface(ifconf)
			interfaces = append(interfaces, trafficni)
			tp := new(network.TrafficParser)
			tp.NewTrafficParser(trafficni, flowcache)
			stop2 := make(chan struct{})
			go tp.Parse(nil, stop2)
		}
	}

	var printer *stats.Printer

	if conf.Stats.Run {
		// TODO refactor to emit times and services
		if conf.Stats.Mode == "dump" {
			printer = stats.NewPrinter(conf.Stats.Append, 60*time.Minute, conf.Sys.OutFolder, "tr")

			ifcollector := stats.IfStatsPrinter{
				Interfaces: interfaces,
			}
			printer.AddCollector(&stats.StatsCollector{
				Period:    10 * time.Second,
				Collector: &ifcollector,
			})

			// TODO for the time being we support only a 10 second emit
			cachedump := stats.CacheDump{
				Fc: flowcache,
			}
			printer.AddCollector(&stats.StatsCollector{
				Period:    10 * time.Second,
				Collector: &cachedump,
			})

			go printer.Run()
		} else {
			panic(errors.New("unknown printer mode"))
		}
	}

	c := make(chan os.Signal, 5)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	log.Infof("Traffic Refinery running")
	<-c
	log.Infof("Traffic Refinery stopping")
	if printer != nil {
		log.Infof("Captured close signal, waiting for clean up of output...")
		printer.Stop()
	}

	if conf.Sys.MemProf {
		runtime.GC() // get up-to-date heap statistics
		if err := pprof.WriteHeapProfile(memf); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		defer memf.Close()
	}
}
