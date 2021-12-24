//go:build afpacket
// +build afpacket

package network

import (
	"fmt"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/bpf"
)

const (
	// Interface buffersize (MB) for AFPacket
	bufferSize uint32 = 1024
	// Fanout group ID
	fanoutGroup uint16 = 1
	promisc     bool   = true
)

type AFHandle struct {
	Name      string
	Filter    string
	SnapLen   uint32
	ZeroCopy  bool
	Clustered bool
	ClusterID int
	FanOut    bool
	TPacket   *afpacket.TPacket
}

// InitAFPacket builds the TPacket on the given device with the given snaplength
func initAFPacket(device string, snaplen uint32, block_size uint32, num_blocks uint32) (*afpacket.TPacket, error) {

	if tp, err := afpacket.NewTPacket(
		afpacket.OptInterface(device),
		afpacket.OptFrameSize(snaplen),
		afpacket.OptBlockSize(block_size),
		afpacket.OptNumBlocks(num_blocks),
		afpacket.SocketRaw,
		afpacket.TPacketVersion3); err != nil {
		return nil, err
	} else {
		return tp, nil
	}

}

// SetBPFFilter translates a BPF filter string into BPF RawInstruction and applies them.
func (h *AFHandle) setBPFFilter(filter string, snaplen uint32) (err error) {
	pcapBPF, err := pcap.CompileBPFFilter(layers.LinkTypeEthernet, int(snaplen), filter)
	if err != nil {
		return err
	}
	bpfIns := []bpf.RawInstruction{}
	for _, ins := range pcapBPF {
		bpfIns2 := bpf.RawInstruction{
			Op: ins.Code,
			Jt: ins.Jt,
			Jf: ins.Jf,
			K:  ins.K,
		}
		bpfIns = append(bpfIns, bpfIns2)
	}
	if h.TPacket.SetBPF(bpfIns); err != nil {
		return err
	}
	return h.TPacket.SetBPF(bpfIns)
}

func (h *AFHandle) newZeroCopyAFPacketInterface() {
	h.ZeroCopy = true
	var err error
	szFrame, szBlock, numBlocks, err := afpacketComputeSize(bufferSize, h.SnapLen, uint32(os.Getpagesize()))
	if err != nil {
		panic(err)
	}
	if h.TPacket, err = initAFPacket(h.Name, szFrame, szBlock, numBlocks); err != nil {
		panic(err)
	}
	if err = h.setBPFFilter(h.Filter, h.SnapLen); err != nil {
		panic(err)
	}
}

func (h *AFHandle) newAFPacketInterface() {
	var err error
	szFrame, szBlock, numBlocks, err := afpacketComputeSize(bufferSize, h.SnapLen, uint32(os.Getpagesize()))
	if err != nil {
		panic(err)
	}
	if h.TPacket, err = initAFPacket(h.Name, szFrame, szBlock, numBlocks); err != nil {
		panic(err)
	}
	if err = h.setBPFFilter(h.Filter, h.SnapLen); err != nil {
		panic(err)
	}
}

// afpacketComputeSize computes the block_size and the num_blocks in such a way that the
// allocated mmap buffer is close to but smaller than target_size_mb.
// The restriction is that the block_size must be divisible by both the
// frame size and page size.
func afpacketComputeSize(targetSizeMb uint32, snaplen uint32, pageSize uint32) (
	frameSize uint32, blockSize uint32, numBlocks uint32, err error) {

	if snaplen < pageSize {
		frameSize = pageSize / (pageSize / snaplen)
	} else {
		frameSize = (snaplen/pageSize + 1) * pageSize
	}

	// 128 is the default from the gopacket library so just use that
	blockSize = frameSize * 128
	numBlocks = (targetSizeMb * 1024 * 1024) / blockSize

	if numBlocks == 0 {
		return 0, 0, 0, fmt.Errorf("Interface buffersize is too small")
	}

	return frameSize, blockSize, numBlocks, nil
}

func (h *AFHandle) Init(conf *HandleConfig) error {
	h.Name = conf.Name
	h.SnapLen = conf.SnapLen
	h.Filter = conf.Filter
	h.FanOut = conf.FanOut
	if conf.ZeroCopy {
		h.newZeroCopyAFPacketInterface()
	} else {
		h.newAFPacketInterface()
	}
	if h.FanOut {
		err := h.TPacket.SetFanout(afpacket.FanoutHashWithDefrag, uint16(fanoutGroup))
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	return nil
}

func (h *AFHandle) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	if h.ZeroCopy {
		return h.TPacket.ZeroCopyReadPacketData()
		// return nil, gopacket.CaptureInfo{}, errors.New("ZeroCopyReadPacketData is not defined for your system")
	} else {
		return h.TPacket.ReadPacketData()
	}
}

func (h *AFHandle) Stats() IfStats {
	_, s, _ := h.TPacket.SocketStats()
	return IfStats{
		PktRecv: uint64(s.Packets()),
		PktDrop: uint64(s.Drops()),
	}
}
