package network

import "github.com/google/gopacket"

type HandleConfig struct {
	Name      string
	Filter    string
	SnapLen   uint32
	Clustered bool
	ClusterID int
	ZeroCopy  bool
	FanOut    bool
}

type Handle interface {
	Init(conf *HandleConfig) error
	ReadPacketData() ([]byte, gopacket.CaptureInfo, error)
	Stats() IfStats
}
