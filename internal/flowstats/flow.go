package flowstats

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/traffic-refinery/traffic-refinery/internal/counters"
	"github.com/traffic-refinery/traffic-refinery/internal/network"
)

// Flow is a general flow interface.
// Functions that all flow type structures have to implement.

// Flow is a general flow structure that contains flow information as well as
// the counters it needs to collect
type Flow struct {
	Id          string
	Service     string
	DomainName  string
	ServiceIP   string
	LocalIP     string
	Protocol    string
	LocalPort   string
	ServicePort string

	Cntrs []counters.Counter
}

func CreateFlow() *Flow {
	return &Flow{}
}

// AddPacket updates the flow states based on the packet pkt
func (f *Flow) AddPacket(pkt *network.Packet) error {

	if pkt == nil {
		return errors.New("packet can not be nil")
	}

	for _, counter := range f.Cntrs {
		log.Debugf("Updating counter of type %s for flow %s", counter.Type(), f.Id)
		counter.AddPacket(pkt)
	}

	log.Debugf("Updated flow %s with Service %s", f.DomainName, f.Service)
	return nil
}

// Reset resets the flow statistics
func (f *Flow) Reset() error {
	for _, counter := range f.Cntrs {
		counter.Reset()
	}
	return nil
}

// Clear the flow statistics
func (f *Flow) Clear() error {
	for _, counter := range f.Cntrs {
		counter.Clear()
	}
	return nil
}

type OutCounter struct {
	CType string
	Data  json.RawMessage
}

type OutFlow struct {
	Id          string
	Service     string
	DomainName  string
	ServiceIP   string
	LocalIP     string
	Protocol    string
	LocalPort   string
	ServicePort string

	Cntrs []OutCounter
}

// Collect converts a flow into JSON form
func (f *Flow) Collect() []byte {
	of := OutFlow{
		Id:          f.Id,
		Service:     f.Service,
		DomainName:  f.DomainName,
		ServiceIP:   f.ServiceIP,
		LocalIP:     f.LocalIP,
		Protocol:    f.Protocol,
		LocalPort:   f.LocalPort,
		ServicePort: f.ServicePort,
	}
	for _, c := range f.Cntrs {
		of.Cntrs = append(of.Cntrs, OutCounter{
			CType: c.Type(),
			Data:  c.Collect(),
		})
	}
	b, _ := json.Marshal(of)
	return b
}
