package network

import (
	"github.com/google/gopacket/layers"
)

type Packet struct {
	RawData     []byte
	Eth         *layers.Ethernet
	Ip4         *layers.IPv4
	Ip6         *layers.IPv6
	Tcp         *layers.TCP
	Udp         *layers.UDP
	Dns         *layers.DNS
	TStamp      int64
	Dir         int
	HwAddr      string
	IsIPv4      bool
	IsLocal     bool
	Length      int64
	ServiceIP   string
	MyIP        string
	IsTCP       bool
	DataLength  int64
	ServicePort uint16
	MyPort      uint16
	SeqNumber   uint32
	IsDNS       bool
}

func NewPacket() *Packet {
	packet := &Packet{}
	packet.Eth = new(layers.Ethernet)
	packet.Ip4 = new(layers.IPv4)
	packet.Ip6 = new(layers.IPv6)
	packet.Tcp = new(layers.TCP)
	packet.Udp = new(layers.UDP)
	return packet
}

func (packet *Packet) Clear() {
	packet.TStamp = 0
	packet.Dir = 0
	packet.HwAddr = ""
	packet.IsIPv4 = false
	packet.IsLocal = false
	packet.Length = 0
	packet.ServiceIP = ""
	packet.MyIP = ""
	packet.IsTCP = false
	packet.DataLength = 0
	packet.ServicePort = 0
	packet.MyPort = 0
	packet.SeqNumber = 0
	packet.IsDNS = false
}
