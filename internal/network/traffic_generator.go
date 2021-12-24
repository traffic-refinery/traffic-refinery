package network

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/traffic-refinery/traffic-refinery/internal/servicemap"
)

// PacketData contains packet and its metadata
type PacketData struct {
	FlowID  string
	Service string
	Pkt     Packet
}

// PacketTrace is a container of ordered packets
type PacketTrace struct {
	Trace []*PacketData
	Count int64
}

type DNSPacketData struct {
	Data  *layers.DNS
	PktTs int64
}

// DNSTrace is a container of ordered DNS responses
type DNSTrace struct {
	Trace []*DNSPacketData
	Count int64
}

func (pktData *PacketData) parseTCPLayer(pkt gopacket.Packet) error {
	tcpLayer := pkt.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return errors.New("not a TCP pkt")
	}

	tcp, _ := tcpLayer.(*layers.TCP)
	pktData.Pkt.Tcp = tcpLayer.(*layers.TCP)
	pktData.Pkt.IsTCP = true
	pktData.Pkt.ServicePort = uint16(tcp.SrcPort)
	pktData.Pkt.MyPort = uint16(tcp.DstPort)
	if pktData.Pkt.Dir == TrafficOut {
		pktData.Pkt.ServicePort = uint16(tcp.DstPort)
		pktData.Pkt.MyPort = uint16(tcp.SrcPort)
	}
	pktData.Pkt.SeqNumber = tcp.Seq
	pktData.Pkt.DataLength = pktData.Pkt.Length - 4*int64(tcp.DataOffset)
	return nil
}

func (pktData *PacketData) parseIPLayer(pkt gopacket.Packet) error {
	if ip4Layer := pkt.Layer(layers.LayerTypeIPv4); ip4Layer != nil {
		ip, _ := ip4Layer.(*layers.IPv4)
		pktData.Pkt.Ip4 = ip
		pktData.Pkt.IsIPv4 = true
		pktData.Pkt.Length = int64(ip.Length - 4*uint16(ip.IHL))
		if ip.DstIP.IsPrivate() {
			pktData.Pkt.MyIP = ip.DstIP.String()
			pktData.Pkt.ServiceIP = ip.SrcIP.String()
			pktData.Pkt.Dir = TrafficIn
		} else {
			pktData.Pkt.MyIP = ip.SrcIP.String()
			pktData.Pkt.ServiceIP = ip.DstIP.String()
			pktData.Pkt.Dir = TrafficOut
		}
		return nil
	} else {
		return errors.New("not an IP pkt")
	}
}

func (pktData *PacketData) parseEthLayer(pkt gopacket.Packet) error {
	ethLayer := pkt.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		return errors.New("not an Ethernet pkt")
	}
	eth, _ := ethLayer.(*layers.Ethernet)
	pktData.Pkt.Eth = eth
	if pktData.Pkt.Dir == TrafficIn {
		pktData.Pkt.HwAddr = eth.DstMAC.String()
	} else {
		pktData.Pkt.HwAddr = eth.SrcMAC.String()
	}
	return nil
}

func populatePacket(pktData *PacketData, pkt *gopacket.Packet) error {
	var err error
	pktData.Pkt.TStamp = (*pkt).Metadata().CaptureInfo.Timestamp.UnixNano()
	pktData.Pkt.RawData = (*pkt).Data()
	// Populate service
	err = pktData.parseIPLayer(*pkt)
	if err != nil {
		return err
	}
	err = pktData.parseEthLayer(*pkt)
	if err != nil {
		return err
	}
	err = pktData.parseTCPLayer(*pkt)
	if err != nil {
		return err
	}
	return nil
}

// GetTraceWithServices preparse a list of packets to process in sequence for testing
func GetTraceWithServices(pcapfile string, sm *servicemap.ServiceMap) *PacketTrace {
	trace := &PacketTrace{}

	if handle, err := pcap.OpenOffline(pcapfile); err != nil {
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for pkt := range packetSource.Packets() {
			pktData := &PacketData{}
			populatePacket(pktData, &pkt)
			if services, found := sm.LookupIP(pktData.Pkt.ServiceIP); found {
				pktData.Service, _ = sm.GetName(services[0])
			} else {
				pktData.Service = "Unknown"
			}
			pktData.FlowID = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d-%d", pktData.Pkt.ServiceIP, pktData.Pkt.MyIP, pktData.Pkt.ServicePort, pktData.Pkt.MyPort))))

			trace.Trace = append(trace.Trace, pktData)
			trace.Count++
		}
	}

	return trace
}

// GetTrace preparse a list of packets to process in sequence for testing
func GetTrace(pcapfile string) *PacketTrace {
	trace := &PacketTrace{}

	if handle, err := pcap.OpenOffline(pcapfile); err != nil {
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for pkt := range packetSource.Packets() {
			pktData := &PacketData{}
			populatePacket(pktData, &pkt)
			pktData.Service = "Unknown"
			pktData.FlowID = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d-%d", pktData.Pkt.ServiceIP, pktData.Pkt.MyIP, pktData.Pkt.ServicePort, pktData.Pkt.MyPort))))
			trace.Trace = append(trace.Trace, pktData)
			trace.Count++
		}
	}

	return trace
}

func GetRandomMac() net.HardwareAddr {
	buf := make([]byte, 6)
	var mac net.HardwareAddr

	_, err := rand.Read(buf)
	if err != nil {
		panic("random number generation error")
	}

	// Set the local bit
	buf[0] |= 2

	mac = append(mac, buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])

	return mac
}

func GetRandomIP() string {
	result := make(net.IP, 4)
	ip := rand.Uint32()
	result[0] = byte(ip)
	result[1] = byte(ip >> 8)
	result[2] = byte(ip >> 16)
	result[3] = byte(ip >> 24)
	return result.String()
}

func GetRandomPort() uint16 {
	return uint16(rand.Intn(65536))
}

// GenerateRandomPacket creates a random packet of length len with given packet size
func GetRandomPacket(len int) (pktData *PacketData) {
	pktData = &PacketData{}
	pktData.Pkt.TStamp = time.Now().UnixNano()
	//Generate random data
	pktData.Pkt.Dir = rand.Intn(2)
	pktData.Pkt.HwAddr = GetRandomMac().String()
	pktData.Pkt.Length = int64(len)
	pktData.Pkt.ServiceIP = GetRandomIP()
	pktData.Pkt.MyIP = GetRandomIP()
	pktData.Pkt.IsIPv4 = true
	pktData.Pkt.DataLength = int64(len) - 20
	pktData.Pkt.ServicePort = GetRandomPort()
	pktData.Pkt.IsTCP = true
	pktData.Pkt.MyPort = GetRandomPort()
	pktData.FlowID = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d-%d", pktData.Pkt.ServiceIP, pktData.Pkt.MyIP, pktData.Pkt.ServicePort, pktData.Pkt.MyPort))))
	pktData.Service = "Random"
	pktData.Pkt.RawData = make([]byte, len)
	rand.Read(pktData.Pkt.RawData)
	return pktData
}

// GetRandomTrace creates a list of random packets to process in sequence for testing
func GetRandomTrace(n, len int) *PacketTrace {
	trace := &PacketTrace{}
	for i := 0; i < n; i++ {
		pktData := GetRandomPacket(len)
		trace.Trace = append(trace.Trace, pktData)
		trace.Count++
	}

	return trace
}

// GetDNSTrace preparses a list of DNS packets to process in sequence for testing
func GetDNSTrace(pcapfile string) *DNSTrace {
	trace := &DNSTrace{}

	if handle, err := pcap.OpenOffline(pcapfile); err != nil {
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for pkt := range packetSource.Packets() {
			// check if packet is a DNS response
			dnsLayer := pkt.Layer(layers.LayerTypeDNS)
			dns, _ := dnsLayer.(*layers.DNS)
			if dnsLayer == nil {
				// not a dns packet
				continue
			} else if dns.ANCount < 1 {
				// no response in the dns packet. Possibly a question
				continue
			}
			trace.Trace = append(trace.Trace, &DNSPacketData{Data: dns, PktTs: pkt.Metadata().CaptureInfo.Timestamp.UnixNano()})
			trace.Count++
		}
	}

	return trace
}
