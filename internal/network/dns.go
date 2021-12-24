package network

import (
	"io"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
	"github.com/traffic-refinery/traffic-refinery/internal/servicemap"
)

// DNSParser
type DNSParser struct {
	netif *NetworkInterface
	sm    *servicemap.ServiceMap
}

func (dp *DNSParser) NewDNSParser(netif *NetworkInterface, sm *servicemap.ServiceMap) {
	dp.netif = netif
	dp.sm = sm
}

// DNSParser is the worker function for parsing network traffic, focusing on dns traffic.
// Reads directly from the NetworkInterface it has been assigned
// The waitgroup is used to cleanly shut down.
func (dp *DNSParser) Parse(wg *sync.WaitGroup, stop chan struct{}) {
	// We use decodinglayerparser, so we set up variables for the dns layer
	// which is the only one to be parsed
	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6
	var tcp layers.TCP
	var udp layers.UDP
	var dns layers.DNS

	// Set up the logrus formatting to use timestamps
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &udp, &tcp, &dns)

	decoded := []gopacket.LayerType{}
	if wg != nil {
		defer wg.Done()
	}

loop:
	for {
		select {
		// signal from main.go has been caught (user shutting down daemon)
		case <-stop:
			break loop
		// process data from ring
		default:
			// Read raw bytes from ring - NOT a gopacket.packet
			pkt, _, err := dp.netif.ReadPacketData()

			if err == io.EOF {
				break
			} else if err != nil {
				continue
			}

			err = parser.DecodeLayers(pkt, &decoded)
			for _, typ := range decoded {
				switch typ {
				case layers.LayerTypeDNS:
					dp.sm.ParseDNSResponse(dns)
				default:
					continue
				}
			}

			if err != nil {
				log.Warnf("Error parsing DNS packet: %s", err)
				continue
			}
		}
	}
}
