//go:build pfring
// +build pfring

package network

import (
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pfring"
)

type RingHandle struct {
	Name      string
	Filter    string
	SnapLen   uint32
	ZeroCopy  bool
	Clustered bool
	ClusterID int
	Ring      *pfring.Ring
}

// InitRing builds the pfring on the given device with the given snaplength
// TODO check that SetBPFFilter works with an empty filter
func initRing(device, filter string, snaplen uint32) (*pfring.Ring, error) {

	if ring, err := pfring.NewRing(device, snaplen, pfring.FlagPromisc); err != nil {
		return nil, err
	} else if err := ring.SetBPFFilter(filter); err != nil {
		return nil, err
	} else if err := ring.Enable(); err != nil {
		return nil, err
	} else {
		//TODO should we optimize this?
		ring.SetPollDuration(0) // default is 500ms - too slow for real-time
		ring.SetPollWatermark(0)
		ring.SetSocketMode(pfring.ReadOnly)
		return ring, nil
	}
}

func addClusterToRing(ring *pfring.Ring, clusterID int, t pfring.ClusterType) error {
	return ring.SetCluster(clusterID, t)
}

func (h *RingHandle) newZeroCopyRingInterface() {
	h.ZeroCopy = true
	var err error
	if h.Ring, err = initRing(h.Name, h.Filter, h.SnapLen); err != nil {
		panic(err)
	}
}

func (h *RingHandle) newClusteredRingInterface(clusterID int, t pfring.ClusterType) {
	h.Clustered = true
	h.ClusterID = clusterID
	var err error
	if h.Ring, err = initRing(h.Name, h.Filter, h.SnapLen); err != nil {
		panic(err)
	}

	if err := addClusterToRing(h.Ring, clusterID, t); err != nil {
		log.Fatalln(err.Error())
		panic(err)
	}
}

func (h *RingHandle) newRingInterface() {
	var err error
	if h.Ring, err = initRing(h.Name, h.Filter, h.SnapLen); err != nil {
		panic(err)
	}
}

func (h *RingHandle) Init(conf *HandleConfig) error {
	h.Name = conf.Name
	h.SnapLen = conf.SnapLen
	h.Filter = conf.Filter
	if conf.Clustered {
		h.newClusteredRingInterface(conf.ClusterID, pfring.ClusterPerFlow5Tuple)
	} else if conf.ZeroCopy {
		h.newZeroCopyRingInterface()
	} else {
		h.newRingInterface()
	}
	return nil
}

func (h *RingHandle) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	if h.ZeroCopy {
		return h.Ring.ZeroCopyReadPacketData()
	} else {
		return h.Ring.ReadPacketData()
	}
}

func (h *RingHandle) Stats() IfStats {
	s, _ := h.Ring.Stats()
	return IfStats{
		PktRecv: uint64(s.Received),
		PktDrop: uint64(s.Dropped),
	}
}
