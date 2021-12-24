//go:build !afpacket
// +build !afpacket

package network

import "github.com/google/gopacket"

type AFHandle struct {
}

func (h *AFHandle) Init(conf *HandleConfig) error {
	panic("No afpacket package available")
}

func (h *AFHandle) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	panic("No afpacket package available")
}

func (h *AFHandle) Stats() IfStats {
	panic("No afpacket package available")
}
