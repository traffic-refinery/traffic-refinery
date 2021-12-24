package welford

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestWelfordSingleAddValue(t *testing.T) {
  w := Welford{}
	w.AddValue(1.0)
}

func TestWelfordSingleCheckAndAddValue(t *testing.T) {
	w := Welford{}
	w.CheckAndAddValue(1.0, 0, 0)
}

func TestWelford10000AddValue(t *testing.T) {
	w := Welford{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	iats := make([]float64, 10000)
	for i := 0; i < 10000; i++ {
		iats[i] = r.Float64()
	}

	for i := 0; i < 10000; i++ {
		w.AddValue(iats[i])
	}
	assert.Equal(t, biggerThan(w.Avg, 0), true)
}

func TestWelford10000CheckAndAddValue(t *testing.T) {
		w := Welford{}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		iats := make([]float64, 10000)
		for i := 0; i < 10000; i++ {
			iats[i] = r.Float64()
		}

		for i := 0; i < 10000; i++ {
			w.CheckAndAddValue(iats[i], 1.0, 1.0)
		}
		assert.Equal(t, biggerThan(w.Avg, 0), true)
}

func BenchmarkWelfordAddValue(b *testing.B) {
	w := Welford{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	iats := make([]float64, b.N)
	for i := 0; i < b.N; i++ {
		iats[i] = r.Float64()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.AddValue(iats[i])
	}
}

func BenchmarkWelfordCheckAndAddValue(b *testing.B) {
		w := Welford{}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		iats := make([]float64, b.N)
		for i := 0; i < b.N; i++ {
			iats[i] = r.Float64()
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w.CheckAndAddValue(iats[i], 1.0, 1.0)
		}
}

// func BenchmarkWelfordPacketsAddValue(b *testing.B) {
// 	if b.N < 2 {
// 		return
// 	}
// 	w := Welford{}
// 	filename := os.TempDir() + string(os.PathSeparator) + "gopacket_benchmark.pcap"
// 	if _, err := os.Stat(filename); err != nil {
// 		fmt.Println("Local pcap file", filename, "doesn't exist, reading from", *url)
// 		if resp, err := http.Get(*url); err != nil {
// 			panic(err)
// 		} else if out, err := os.Create(filename); err != nil {
// 			panic(err)
// 			//i//} else if content, err := ioutil.ReadAll(resp.Body); err != nil {
// 			//	panic(err)
// 		} else if _, err := io.Copy(out, resp.Body); err != nil {
// 			panic(err)
// 		} else if err := out.Close(); err != nil {
// 			panic(err)
// 		}
// 	}
// 	if f, err := os.Open(filename); err != nil {
// 		panic(err)
// 	} else if _, err := io.Copy(ioutil.Discard, f); err != nil {
// 		panic(err)
// 	} else if err := f.Close(); err != nil {
// 		panic(err)
// 	}
//
// 	var packetDataSource *BufferPacketSource
// 	var packetSource *gopacket.PacketSource
// 	if h, err := pcap.OpenOffline(filename); err != nil {
// 		panic(err)
// 	} else {
// 		packetDataSource = NewBufferPacketSource(h)
// 		packetSource = gopacket.NewPacketSource(packetDataSource, h.LinkType())
// 	}
//
// 	packetDataSource.Reset()
// 	count := 0
// 	for _, err := packetSource.NextPacket(); err != io.EOF; _, err = packetSource.NextPacket() {
// 		count++
// 	}
//
// 	lastPkt := int64(0)
// 	resetTs := int64(0)
// 	firstTs := int64(0)
// 	iats := make([]float64, 0)
//
// 	for i := 0; i < b.N + 1; {
// 		packetDataSource.Reset()
// 		if i != 0 {
// 			resetTs = lastPkt - firstTs
// 		}
// 		for packet, err := packetSource.NextPacket(); err != io.EOF; packet, err = packetSource.NextPacket() {
// 			ts := packet.Metadata().Timestamp.UnixNano() / 1000 + resetTs
// 			if i == 0 {
// 				firstTs = ts
// 			}
//
// 			// fmt.Printf("Adding packet with ts %v", ts)
//
// 			if lastPkt != 0 {
// 				iat := float64(ts) - float64(lastPkt)
// 				iats = append(iats, iat)
// 			}
// 			lastPkt = ts
//
// 			i++
// 			if i >= b.N + 1 {
// 				break
// 			}
// 		}
// 	}
//
// 	b.ResetTimer()
//
// 	// fmt.Printf("The size of the test is %v\n", b.N)
// 	for i := 0; i < b.N; i++ {
// 		// fmt.Printf("%d\n", i)
// 		w.AddValue(iats[i])
// 	}
// }
//
// func BenchmarkWelfordPacketsCheckAndAddValue(b *testing.B) {
// 	if b.N < 2 {
// 		return
// 	}
// 	w := Welford{}
// 	filename := os.TempDir() + string(os.PathSeparator) + "gopacket_benchmark.pcap"
// 	if _, err := os.Stat(filename); err != nil {
// 		fmt.Println("Local pcap file", filename, "doesn't exist, reading from", *url)
// 		if resp, err := http.Get(*url); err != nil {
// 			panic(err)
// 		} else if out, err := os.Create(filename); err != nil {
// 			panic(err)
// 			//i//} else if content, err := ioutil.ReadAll(resp.Body); err != nil {
// 			//	panic(err)
// 		} else if _, err := io.Copy(out, resp.Body); err != nil {
// 			panic(err)
// 		} else if err := out.Close(); err != nil {
// 			panic(err)
// 		}
// 	}
// 	if f, err := os.Open(filename); err != nil {
// 		panic(err)
// 	} else if _, err := io.Copy(ioutil.Discard, f); err != nil {
// 		panic(err)
// 	} else if err := f.Close(); err != nil {
// 		panic(err)
// 	}
//
// 	var packetDataSource *BufferPacketSource
// 	var packetSource *gopacket.PacketSource
// 	if h, err := pcap.OpenOffline(filename); err != nil {
// 		panic(err)
// 	} else {
// 		packetDataSource = NewBufferPacketSource(h)
// 		packetSource = gopacket.NewPacketSource(packetDataSource, h.LinkType())
// 	}
//
// 	packetDataSource.Reset()
// 	count := 0
// 	for _, err := packetSource.NextPacket(); err != io.EOF; _, err = packetSource.NextPacket() {
// 		count++
// 	}
//
// 	lastPkt := int64(0)
// 	resetTs := int64(0)
// 	firstTs := int64(0)
// 	iats := make([]float64, 0)
//
// 	for i := 0; i < b.N + 1; {
// 		packetDataSource.Reset()
// 		if i != 0 {
// 			resetTs = lastPkt - firstTs
// 		}
// 		for packet, err := packetSource.NextPacket(); err != io.EOF; packet, err = packetSource.NextPacket() {
// 			ts := packet.Metadata().Timestamp.UnixNano() / 1000 + resetTs
// 			if i == 0 {
// 				firstTs = ts
// 			}
//
// 			if lastPkt != 0 {
// 				iat := float64(ts) - float64(lastPkt)
// 				iats = append(iats, iat)
// 			}
// 			lastPkt = ts
//
// 			i++
// 			if i >= b.N + 1 {
// 				break
// 			}
// 		}
// 	}
//
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		//TODO add similar process to what done in TA
// 		w.CheckAndAddValue(iats[i], 1.0, 1.0)
// 	}
// }

func biggerThan(first float64, second float64) bool {
	return first > second
}
