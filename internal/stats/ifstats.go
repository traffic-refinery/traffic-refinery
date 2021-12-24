// Package stats implements different methods to print various statistics
package stats

import (
	"encoding/json"
	"time"

	"github.com/traffic-refinery/traffic-refinery/internal/network"
)

type IfStatsPrinter struct {
	Interfaces []*network.NetworkInterface
	lastTime   int64
}

type ParserStats struct {
	Name    string
	PktRecv uint64
	PktDrop uint64
}

func NewIfStatsPrinter(inter []*network.NetworkInterface) *IfStatsPrinter {
	cp := new(IfStatsPrinter)
	cp.Interfaces = inter
	return cp
}

func (cp *IfStatsPrinter) Type() string {
	return "IfStatsPrinter"
}

func (cp *IfStatsPrinter) Init() error {
	cp.lastTime = time.Now().Unix()
	return nil
}

func (cp *IfStatsPrinter) Run() []byte {
	endTime := time.Now().Unix()
	parsers := make([]ParserStats, len(cp.Interfaces))

	for i, iface := range cp.Interfaces {
		s := iface.IfHandle.Stats()
		parsers[i].PktRecv = s.PktRecv
		parsers[i].PktDrop = s.PktDrop
	}

	parsersData, _ := json.Marshal(parsers)

	outJson := OutJson{
		Version: "3.0",
		Conf:    "--",
		Type:    cp.Type(),
		TsStart: cp.lastTime,
		TsEnd:   endTime,
		Data:    parsersData,
	}

	cp.lastTime = endTime

	b, _ := json.Marshal(outJson)
	return b
}
