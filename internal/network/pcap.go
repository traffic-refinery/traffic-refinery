package network

import (
	"errors"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	log "github.com/sirupsen/logrus"
)

type PcapHandle struct {
	Name      string
	Filter    string
	SnapLen   uint32
	ZeroCopy  bool
	Clustered bool
	ClusterID int
	FanOut    bool
	PHandle   *pcap.Handle
}

func initPcap(device, filter string, snaplen uint32) (*pcap.Handle, error) {
	inactiveHandle, err := pcap.NewInactiveHandle(device)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	inactiveHandle.SetSnapLen(int(snaplen))
	inactiveHandle.SetPromisc(true)
	inactiveHandle.SetTimeout(pcap.BlockForever)
	// inactiveHandle.SetBufferSize(1000000 * ph.BufferMb)

	handle, err := inactiveHandle.Activate()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return handle, nil
}

func (h *PcapHandle) NewPcapInterface() {
	var err error
	if h.PHandle, err = initPcap(h.Name, h.Filter, h.SnapLen); err != nil {
		panic(err)
	}

}

func (h *PcapHandle) Init(conf *HandleConfig) error {
	h.Name = conf.Name
	h.SnapLen = conf.SnapLen
	h.Filter = conf.Filter
	h.NewPcapInterface()
	return nil
}

func (h *PcapHandle) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	if h.ZeroCopy {
		log.Fatal("You can not read zero copy from pcap")
		return nil, gopacket.CaptureInfo{}, errors.New("You can not read zero copy from pcap")
	} else {
		return h.PHandle.ReadPacketData()
	}
}

func (h *PcapHandle) Stats() IfStats {
	s, _ := h.PHandle.Stats()
	return IfStats{
		PktRecv: uint64(s.PacketsReceived),
		PktDrop: uint64(s.PacketsDropped),
	}
}
