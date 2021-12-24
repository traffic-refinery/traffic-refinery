// Package config is used to configure traffic refinery
package config

import (
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// SysConfig provides general configurations for the Traffic Refinery system.
type SysConfig struct {
	// CPU is a boolean determining whether to run CPU profiling
	CPUProf bool
	// MemProf is a boolean determining whether to run MemProf profiling
	MemProf bool
	// InterfacesStats is a boolean determing whether to print out interface
	// statistics
	InterfacesStats bool
	// OutFolder is the path where to store the output files
	OutFolder string
}

// ParserConfig provides configurations for a single parser
type ParserConfig struct {
	// Driver type. Either "ring" (PF_RING) or "pcap" (PCAP) or "afpacket (AF Packet)"
	Driver string
	// Whether to use PF_RING clustering for load balancing across threads
	Clustered bool
	// ID of the cluster to use
	ClusterID int
	// Whether to use PF_RING in Zero Copy mode. Not available if Clustered is true
	ZeroCopy bool
	// Whether to use AFPacket Fanout
	FanOut bool
	// Name of the interface to use
	Ifname string
	// Mode for hte interface. Supports "host"|"router"|"mirror" modes
	Mode string
	// Whether it's a replay session
	Replay bool
	// Gateway MAC address used when in replay mode
	ReplayMAC string
	// How many replicas of the same parser type
	Replicas int
}

// ParsersConfig provides configurations for packet capture and processing.
type ParsersConfig struct {
	// Struct containing the configuration for the DNS parser
	DNSParser ParserConfig
	// Array of struct containing the configurations of the traffic parsers
	TrafficParsers []ParserConfig
}

// DNSCacheConfig provides configurations used by the DNSCache
type DNSCacheConfig struct {
	// Time before eviction from the cache
	EvictTime time.Duration
	// Length of the garbage collection period.
	CleanupTime time.Duration
}

// FlowCacheConfig provides configurations used by the FlowCache
type FlowCacheConfig struct {
	// Cache type. Currently only supports "ConcurrentCacheMap"
	CacheType string
	// Time before eviction from the cache
	EvictTime time.Duration
	// Number of shards used for concurrency
	ShardsCount int
	// Length of the garbage collection period.
	CleanupTime time.Duration
	// Whether to anonymize IP addresses or not
	Anonymize bool
}

// StatsConfig contains basic configurations on how to print statistics
type StatsOutConfig struct {
	// Run determines whether to run a printer or not
	Run bool
	// Mode determines which printer to use. Currently only supports "dump"
	Mode string
	// Append determines whether to append entries to the file or to
	// reflush the same one every cycle
	Append bool
}

// ServiceFilterConfig contains the set of filters used to filter
// traffic classes
type ServiceFilterConfig struct {
	// DomainsString is the list of domains to match
	DomainsString []string
	// DomainsRegex is the list of regexes to match
	DomainsRegex []string
	// Prefixes is the list of subnets to match
	Prefixes []string
}

// ServiceConfig contains the details of a service to track
type ServiceConfig struct {
	// Name of the service
	Name string
	// Filter of the service
	Filter ServiceFilterConfig
	// Collect is the list of features to collect for the service
	Collect []string
	// Emit is the cycle length for stats printouts in ms
	Emit time.Duration
}

// TrafficRefineryConfig contains all configuration structures required by traffic refinery
type TrafficRefineryConfig struct {
	Sys       SysConfig
	Parsers   ParsersConfig
	DNSCache  DNSCacheConfig
	FlowCache FlowCacheConfig
	Stats     StatsOutConfig
	Services  []ServiceConfig
}

func (conf *TrafficRefineryConfig) setDefaults() {
	viper.SetDefault("Sys.CPUProf", false)
	viper.SetDefault("Sys.MemProf", false)
	viper.SetDefault("Sys.InterfaceStats", false)
	viper.SetDefault("Sys.OutFolder", "/tmp/")

	viper.SetDefault("Parsers.DNSParser", ParserConfig{})
	viper.SetDefault("Parsers.TrafficParsers", []ParserConfig{})

	viper.SetDefault("DnsCache.CleanupTime", 5*time.Minute)
	viper.SetDefault("DnsCache.EvictTime", 10*time.Minute)

	viper.SetDefault("FlowCache.CacheType", "ConcurrentCacheMap")
	viper.SetDefault("FlowCache.EvictTime", 10*time.Minute)
	viper.SetDefault("FlowCache.CleanupTime", 5*time.Minute)
	viper.SetDefault("FlowCache.ShardsCount", 32)
	viper.SetDefault("FlowCache.Anonymize", true)

	viper.SetDefault("Stats.Run", false)
	viper.SetDefault("Stats.Mode", "dump")
	viper.SetDefault("Stats.Append", false)

	viper.SetDefault("Services", []ServiceConfig{})
}

// ImportConfig uses a conventional file named tr"config" to load the configuration
func (conf *TrafficRefineryConfig) ImportConfig() {
	conf.setDefaults()
	viper.SetConfigName("trconfig")               // name of config file (without extension)
	viper.AddConfigPath("./")                     // optionally look for config in the working directory
	viper.AddConfigPath("/etc/traffic_refinery/") // path to look for the config file in
	err := viper.ReadInConfig()                   // Find and read the config file
	if err != nil {                               // Handle errors reading the config file
		panic(err)
	}
	conf.loadSystemConfig()
	conf.loadParsersConfig()
	conf.loadDNSCacheConfig()
	conf.loadFlowCacheConfig()
	conf.loadStatsConfig()
	conf.loadServiceConfig()
}

// ImportConfigFromFile uses a conventional file named path/configName" to load the configuration
func (conf *TrafficRefineryConfig) ImportConfigFromFile(fileName string) {
	conf.setDefaults()
	path, name := filepath.Split(fileName)
	if path == "" {
		path = "."
	}
	extensionType := strings.TrimPrefix(filepath.Ext(fileName), ".")
	viper.SetConfigName(name)          // name of config file (without extension)
	viper.AddConfigPath(path)          // optionally look for config in the working directory
	viper.SetConfigType(extensionType) // type of configuration file based on the extension
	err := viper.ReadInConfig()        // Find and read the config file
	if err != nil {                    // Handle errors reading the config file
		panic(err)
	}
	conf.loadSystemConfig()
	conf.loadParsersConfig()
	conf.loadDNSCacheConfig()
	conf.loadFlowCacheConfig()
	conf.loadStatsConfig()
	conf.loadServiceConfig()
}

// PrintConfig prints the current configuration.
func (conf *TrafficRefineryConfig) PrintConfig() {
	log.Debugf("Parsers configuration - Not implemented")
}

// LoadParsersConfig loads the configuration from viper.
func (conf *TrafficRefineryConfig) loadSystemConfig() {
	conf.Sys.CPUProf = viper.GetBool("Sys.CPUProf")
	conf.Sys.MemProf = viper.GetBool("Sys.MemProf")
	conf.Sys.OutFolder = viper.GetString("Sys.OutFolder")
}

// LoadParsersConfig loads the configuration from viper.
func (conf *TrafficRefineryConfig) loadParsersConfig() {
	conf.Parsers.DNSParser.Driver = viper.GetString("Parsers.DNSParser.Driver")
	conf.Parsers.DNSParser.Clustered = viper.GetBool("Parsers.DNSParser.Clustered")
	conf.Parsers.DNSParser.ClusterID = viper.GetInt("Parsers.DNSParser.ClusterID")
	conf.Parsers.DNSParser.ZeroCopy = viper.GetBool("Parsers.DNSParser.ZeroCopy")
	conf.Parsers.DNSParser.Ifname = viper.GetString("Parsers.DNSParser.Ifname")
	conf.Parsers.DNSParser.Mode = viper.GetString("Parsers.DNSParser.Mode")
	conf.Parsers.DNSParser.Replay = viper.GetBool("Parsers.DNSParser.Replay")
	conf.Parsers.DNSParser.ReplayMAC = viper.GetString("Parsers.DNSParser.ReplayMAC")
	if err := viper.UnmarshalKey("Parsers.TrafficParsers", &conf.Parsers.TrafficParsers); err != nil {
		panic(err)
	}
}

func (conf *TrafficRefineryConfig) loadDNSCacheConfig() {
	conf.DNSCache.CleanupTime = viper.GetDuration("DNSCache.CleanupTime")
	conf.DNSCache.EvictTime = viper.GetDuration("DNSCache.EvictTime")

}

func (conf *TrafficRefineryConfig) loadFlowCacheConfig() {
	conf.FlowCache.CacheType = viper.GetString("FlowCache.CacheType")
	conf.FlowCache.EvictTime = viper.GetDuration("FlowCache.EvictTime")
	conf.FlowCache.CleanupTime = viper.GetDuration("FlowCache.CleanupTime")
	conf.FlowCache.ShardsCount = viper.GetInt("FlowCache.ShardsCount")
	conf.FlowCache.Anonymize = viper.GetBool("FlowCache.Anonymize")
}

func (conf *TrafficRefineryConfig) loadStatsConfig() {
	conf.Stats.Run = viper.GetBool("Stats.Run")
	conf.Stats.Mode = viper.GetString("Stats.Mode")
	conf.Stats.Append = viper.GetBool("Stats.Append")
}

func (conf *TrafficRefineryConfig) loadServiceConfig() {
	if err := viper.UnmarshalKey("Services", &conf.Services); err != nil {
		panic(err)
	}
}
