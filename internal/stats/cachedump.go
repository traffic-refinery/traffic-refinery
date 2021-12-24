// Package stats implements different methods to print various statistics
package stats

import (
	"encoding/json"
	"time"

	"github.com/traffic-refinery/traffic-refinery/internal/flowstats"
)

type OutTraffic struct {
	Flows []json.RawMessage
}

type CacheDump struct {
	Fc       *flowstats.FlowCache
	lastTime int64
}

func NewCacheDump(fc *flowstats.FlowCache) *CacheDump {
	cp := new(CacheDump)
	cp.Fc = fc
	return cp
}

func (cp *CacheDump) Type() string {
	return "CacheDump"
}

func (cp *CacheDump) Init() error {
	cp.lastTime = time.Now().Unix()
	return nil
}

func (cp *CacheDump) Run() []byte {
	endTime := time.Now().Unix()

	outJson := OutJson{
		Version: "3.0",
		Conf:    "--",
		Type:    cp.Type(),
		TsStart: cp.lastTime,
		TsEnd:   endTime,
		Data:    nil,
	}

	trafficData := cp.Fc.DumpToString()
	cp.lastTime = endTime

	data, _ := json.Marshal(trafficData)
	outJson.Data = data
	data, _ = json.Marshal(outJson)
	return data
}
