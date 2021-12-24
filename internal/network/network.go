package network

import (
	"bytes"
	"errors"
	"net"
	"os/exec"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	log "github.com/sirupsen/logrus"
)

// Configuration for network interface.
const (
	// hostMode should be used when tracking traffic from a network end point.
	hostMode   = "host"
	apMode     = "router"
	mirrorMode = "mirror"
	replayMode = "replay"
)

const (
	TrafficIn  = 0
	TrafficOut = 1
)

const (
	HandleTypePFRing   = 0
	HandleTypePcap     = 1
	HandleTypeAFPacket = 2
)

// BPF Filter for capturing DNS traffic only
const DNSFilter = "udp and port 53"

// BPF Filter for capturing DNS all traffic but DNS
// const NotDNSFilter = "tcp or (udp and not port 53)"
const NotDNSFilter = "tcp or (udp and not port 53)"

// NetworkInterfaceConfiguration is a support structure used to configure an interface
type NetworkInterfaceConfiguration struct {
	// name, filter, mode string, snaplen uint32
	Driver    string
	Name      string
	Mode      string
	Filter    string
	SnapLen   uint32
	Clustered bool
	ClusterID int
	Replay    bool
	ReplayMAC string
	ZeroCopy  bool
	FanOut    bool
}

// NetworkInterface is a structure that carries information on the interface it maps to
// and pointers to the underlying packet processing tool (PFRing or Pcap)
type NetworkInterface struct {
	Mode       string
	Name       string
	HwAddr     net.HardwareAddr
	LocalNetv4 net.IPNet
	LocalNetv6 net.IPNet
	HandleType uint8
	IfHandle   Handle
}

func getMirrorMac(iface string) (net.HardwareAddr, error) {
	cmd := exec.Command("arp", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	arpOut := strings.Split(out.String(), "\n")
	for _, arpLine := range arpOut {
		arpSplit := strings.Split(arpLine, " ")
		if len(arpSplit) > 6 {
			if iface == arpSplit[6] || iface == arpSplit[5] {
				mac, err := net.ParseMAC(arpSplit[3])
				if err == nil {
					return mac, nil
				}
			}
		}
	}
	return nil, errors.New("Could not find mac address of mirror switch on " + iface)
}

func getMacFromName(name string) (net.HardwareAddr, net.IPNet, net.IPNet) {
	var hardwareAddr net.HardwareAddr
	var localNetv4, localNetv6 net.IPNet
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	// handle err
	for _, i := range ifaces {
		if i.Name == name {
			addrs, _ := i.Addrs()
			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPNet:
					if v.IP.IsGlobalUnicast() && v.IP.To4() != nil {
						localNetv4 = *v
					} else if v.IP.IsGlobalUnicast() && v.IP.To16() != nil {
						localNetv6 = *v
					}
				}
			}

			hardwareAddr = i.HardwareAddr
			break
		}
	}
	return hardwareAddr, localNetv4, localNetv6
}

func (ni *NetworkInterface) getDirection(eth *layers.Ethernet) (int, error) {
	dir := -1
	if ni.Mode == apMode || ni.Mode == mirrorMode {
		if eth.DstMAC.String() == ni.HwAddr.String() {
			dir = TrafficOut
		} else if eth.SrcMAC.String() == ni.HwAddr.String() {
			dir = TrafficIn
		}
	} else if ni.Mode == hostMode {
		// log.Debugf("src %s dst %s hw %s", eth.SrcMAC.String(), eth.DstMAC.String(), ni.HwAddr.String())
		if eth.SrcMAC.String() == ni.HwAddr.String() {
			dir = TrafficOut
		} else if eth.DstMAC.String() == ni.HwAddr.String() {
			dir = TrafficIn
		}
	} else {
		panic(errors.New("interface mode not set"))
	}
	return dir, nil
}

func (ni *NetworkInterface) NewNetworkInterface(conf NetworkInterfaceConfiguration) {
	ni.Name = conf.Name
	ni.Mode = conf.Mode

	// Get MAC address of interface in use
	var err error
	if conf.Replay {
		if hwAddr, err := net.ParseMAC(conf.ReplayMAC); err != nil {
			panic(err)
		} else {
			ni.HwAddr = hwAddr
		}
	} else if conf.Mode == "mirror" {
		ni.HwAddr, err = getMirrorMac(ni.Name)
		if err != nil {
			panic(err)
		}
	} else {
		ni.HwAddr, ni.LocalNetv4, ni.LocalNetv6 = getMacFromName(ni.Name)
	}

	// Initiate the interface based on type
	if conf.Driver == "pcap" {
		ni.HandleType = HandleTypePcap
		ni.IfHandle = &PcapHandle{}
	} else if conf.Driver == "ring" {
		ni.HandleType = HandleTypePFRing
		ni.IfHandle = &RingHandle{}
	} else if conf.Driver == "afpacket" {
		ni.HandleType = HandleTypeAFPacket
		ni.IfHandle = &AFHandle{}
	} else {
		panic(errors.New("wrong interface driver type"))
	}
	hc := HandleConfig{
		Name:      conf.Name,
		Filter:    conf.Filter,
		SnapLen:   conf.SnapLen,
		Clustered: conf.Clustered,
		ClusterID: conf.ClusterID,
		ZeroCopy:  conf.ZeroCopy,
		FanOut:    conf.FanOut,
	}
	ni.IfHandle.Init(&hc)

}

func (ni *NetworkInterface) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	return ni.IfHandle.ReadPacketData()
}
