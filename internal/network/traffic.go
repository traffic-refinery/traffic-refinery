package network

import (
	"io"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
)

type TrafficParser struct {
	netif           *NetworkInterface
	packetProcessor PacketProcessor
}

func (tp *TrafficParser) NewTrafficParser(netif *NetworkInterface, packetProcessor PacketProcessor) {
	tp.netif = netif
	tp.packetProcessor = packetProcessor
}

func (tp *TrafficParser) parseUdpLayer(udp *layers.UDP, dir int) (int64, uint16, uint16, error) {
	sPort := udp.SrcPort
	lPort := udp.DstPort
	if dir == TrafficOut {
		sPort = udp.DstPort
		lPort = udp.SrcPort
	}
	return int64(udp.Length - 8), uint16(sPort), uint16(lPort), nil
}

func (tp *TrafficParser) parseTcpLayer(tcp *layers.TCP, ipDataLen int64, dir int) (int64, uint16, uint16, uint32, error) {
	sPort := tcp.SrcPort
	lPort := tcp.DstPort
	if dir == TrafficOut {
		sPort = tcp.DstPort
		lPort = tcp.SrcPort
	}
	seq := tcp.Seq
	return ipDataLen - 4*int64(tcp.DataOffset), uint16(sPort), uint16(lPort), seq, nil
}

func (tp *TrafficParser) parseIpV4Layer(ip *layers.IPv4, dir int) (int64, string, string, bool, error) {
	var isLocal bool
	var ipDataLen int64
	var sIp, mIp string

	ipDataLen = int64(ip.Length - 4*uint16(ip.IHL))
	sIp = ip.SrcIP.String()
	if dir == TrafficOut {
		sIp = ip.DstIP.String()
		mIp = ip.SrcIP.String()
		isLocal = tp.netif.LocalNetv4.Contains(ip.DstIP)
	} else {
		mIp = ip.DstIP.String()
		isLocal = tp.netif.LocalNetv4.Contains(ip.SrcIP)
	}
	if !isLocal && IsRFC1918(ip.DstIP) && IsRFC1918(ip.SrcIP) {
		isLocal = true
	}

	return ipDataLen, sIp, mIp, isLocal, nil
}

func (tp *TrafficParser) parseIpV6Layer(ip *layers.IPv6, dir int) (int64, string, string, bool, error) {
	var isLocal bool
	var ipDataLen int64
	var sIp, mIp string

	ipDataLen = int64(ip.Length)
	sIp = ip.SrcIP.String()
	if dir == TrafficOut {
		sIp = ip.DstIP.String()
		mIp = ip.SrcIP.String()
		isLocal = tp.netif.LocalNetv6.Contains(ip.DstIP)
	} else {
		mIp = ip.DstIP.String()
		isLocal = tp.netif.LocalNetv6.Contains(ip.SrcIP)
	}

	return ipDataLen, sIp, mIp, isLocal, nil
}

func (tp *TrafficParser) parseEthLayer(eth *layers.Ethernet, dir int) (string, error) {
	var hwAddr string
	if dir == TrafficOut {
		hwAddr = eth.SrcMAC.String()
	} else {
		hwAddr = eth.DstMAC.String()
	}
	return hwAddr, nil
}

// TrafficParser is the worker function for parsing network traffic. Each worker reads directly from the ring that is passed
// The waitgroup is used to cleanly shut down. Each worker listen on the stop chan to know when to stop processing
func (tp *TrafficParser) Parse(wg *sync.WaitGroup, stop chan struct{}) {
	// We use decodinglayerparser, so we set up variables for the layers we intend to parse
	pkt := NewPacket()
	var vlantag *layers.Dot1Q
	vlantag = new(layers.Dot1Q)

	// We use Flows to access the network and transport endpoints when building the 4-tuple flow
	// var netFlow, tranFlow gopacket.Flow
	// isValid is a flag used to tell the worker whether or not to process the information in a packet
	var isValid bool
	var parsingErr error

	// initialize the CIDR IP range slice
	CIDRinit()

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, pkt.Eth, vlantag, pkt.Ip4, pkt.Ip6, pkt.Tcp, pkt.Udp)
	decoded := []gopacket.LayerType{}
	if wg != nil {
		defer wg.Done()
	}
loop:
	for {
		pkt.Clear()
		select {
		// signal from main.go has been caught (user shutting down daemon)
		case <-stop:
			break loop
		// process data from ring
		default:
			// Read raw bytes from ring - NOT a gopacket.packet
			data, ci, err := tp.netif.ReadPacketData()
			pkt.TStamp = ci.Timestamp.UnixNano()
			pkt.RawData = data

			// reset variables
			isValid = false

			if err == io.EOF {
				break
			} else if err != nil {
				continue
			}

			err = parser.DecodeLayers(data, &decoded)

			//TODO handle the fact that there are case of errors even when it should not be interrupted
			if err != nil {
				log.Debugln(err)
			}

		decoding_loop:
			for _, typ := range decoded {
				switch typ {
				case layers.LayerTypeEthernet:
					pkt.Dir, parsingErr = tp.netif.getDirection(pkt.Eth)
					if pkt.Dir == -1 {
						break decoding_loop
					}
					pkt.HwAddr, parsingErr = tp.parseEthLayer(pkt.Eth, pkt.Dir)
				case layers.LayerTypeIPv4:
					pkt.Length, pkt.ServiceIP, pkt.MyIP, pkt.IsLocal, parsingErr = tp.parseIpV4Layer(pkt.Ip4, pkt.Dir)
					pkt.IsIPv4 = true
				case layers.LayerTypeTCP:
					pkt.DataLength, pkt.ServicePort, pkt.MyPort, pkt.SeqNumber, parsingErr = tp.parseTcpLayer(pkt.Tcp, pkt.Length, pkt.Dir)
					pkt.IsTCP = true
					isValid = true
				case layers.LayerTypeUDP:
					pkt.DataLength, pkt.ServicePort, pkt.MyPort, parsingErr = tp.parseUdpLayer(pkt.Udp, pkt.Dir)
					pkt.IsTCP = false
					isValid = true
				case layers.LayerTypeIPv6:
					pkt.Length, pkt.ServiceIP, pkt.MyIP, pkt.IsLocal, parsingErr = tp.parseIpV6Layer(pkt.Ip6, pkt.Dir)
					pkt.IsIPv4 = false
				}
			}

			if parsingErr != nil {
				log.Warnln(err)
				continue
			}
			if !isValid {
				log.Debugf("Read packet without required layers or with wrong direction")
				continue
			}

			tp.packetProcessor.ProcessPacket(pkt)
		}
	}
}
