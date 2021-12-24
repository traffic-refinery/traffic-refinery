// Package stats implements different methods to print various statistics
package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type OutJson struct {
	Version string
	Conf    string
	Type    string
	TsStart int64
	TsEnd   int64
	Data    json.RawMessage
}

type StatsCollector struct {
	Period    time.Duration
	Collector Stats
	ticker    *time.Ticker
	end       chan bool
}

type Printer struct {
	outDir     string
	baseName   string
	period     time.Duration
	app        bool
	End        chan bool
	f          *os.File
	wTime      int64
	collectors []*StatsCollector
}

func NewPrinter(app bool, period time.Duration, outDir, baseName string) *Printer {
	cp := new(Printer)
	cp.app = app
	cp.End = make(chan bool, 1)
	cp.outDir = outDir
	cp.baseName = baseName
	cp.period = period
	return cp
}

func (cp *Printer) AddCollector(collector *StatsCollector) {
	cp.collectors = append(cp.collectors, collector)
}

func (cp *Printer) clean() {
	for nc := range cp.collectors {
		cp.collectors[nc].end <- true
	}
	if cp.app {
		os.Rename(cp.f.Name(), fmt.Sprintf("%s/%s.%d.out", cp.outDir, cp.baseName, cp.wTime))
	}
}

func (cp *Printer) Stop() {
	cp.End <- true
	cp.clean()
}

func (cp *Printer) Run() {
	var err error

	if cp.app {
		cp.f, err = ioutil.TempFile(cp.outDir, fmt.Sprintf("tmp.%s.", cp.baseName))
		if err != nil {
			panic("Could not create tmp output file")
		}
	}

	ticker := time.NewTicker(time.Duration(cp.period))
	cp.wTime = time.Now().Unix()

	// Start running all stats collectors
	for _, collector := range cp.collectors {
		go func(sc *StatsCollector) {
			sc.end = make(chan bool, 1)
			sc.Collector.Init()
			sc.ticker = time.NewTicker(sc.Period)
			for {
				select {
				case <-sc.end:
					return
				case <-sc.ticker.C:
					s := sc.Collector.Run()
					if cp.app {
						cp.f.WriteString(fmt.Sprintf("%s\n", s))
					} else {
						err = ioutil.WriteFile(fmt.Sprintf("%s/%s.out", cp.outDir, cp.baseName), s, 0644)
					}
				}
			}
		}(collector)
	}

	for {
		select {

		case <-cp.End:
			return

		case <-ticker.C:
			log.Infoln("Printing out flow stats to file")
			cTime := time.Now().Unix()
			if cp.app {
				log.Debugln("Wrapping up out file")
				err = os.Rename(cp.f.Name(), fmt.Sprintf("%s/%s.%d.out", cp.outDir, cp.baseName, cp.wTime))
				cp.wTime = cTime
				if err != nil {
					panic("Could not move tmp file to output file")
				}
				cp.f, err = ioutil.TempFile(cp.outDir, fmt.Sprintf("tmp.%s.", cp.baseName))
				if err != nil {
					panic("Could not create tmp output file")
				}
			}
		}
	}

}
