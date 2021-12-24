package counters

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBuildEmpty(t *testing.T) {
	ac := AvailableCounters{}
	if _, err := ac.Build(nil); err != nil {
		panic(err)
	}
}

func TestBuildDefaults(t *testing.T) {
	ac := AvailableCounters{}
	if _, err := ac.Build([]string{"LatencyJitterCounter", "PacketCounters", "TCPState", "VideoCounters"}); err != nil {
		panic(err)
	}
}

func TestInitialization(t *testing.T) {
	ac := AvailableCounters{}
	counters := []string{"LatencyJitterCounter", "PacketCounters", "TCPState", "VideoCounters"}
	if _, err := ac.Build(counters); err != nil {
		panic(err)
	}
	counter, _ := ac.InstantiateByName("PacketCounters")
	if reflect.TypeOf(counter) != reflect.TypeOf(&PacketCounters{}) {
		panic(fmt.Sprintf("%s %s", reflect.TypeOf(counter).String(), reflect.TypeOf(&PacketCounters{}).String()))
	}
	counter, _ = ac.InstantiateByName("LatencyJitterCounter")
	if reflect.TypeOf(counter) != reflect.TypeOf(&LatencyJitterCounter{}) {
		panic(fmt.Sprintf("%s %s", reflect.TypeOf(counter).String(), reflect.TypeOf(&PacketCounters{}).String()))
	}
	counter, _ = ac.InstantiateByName("TCPState")
	if reflect.TypeOf(counter) != reflect.TypeOf(&TCPState{}) {
		panic(fmt.Sprintf("%s %s", reflect.TypeOf(counter).String(), reflect.TypeOf(&PacketCounters{}).String()))
	}
	counter, _ = ac.InstantiateByName("VideoCounters")
	if reflect.TypeOf(counter) != reflect.TypeOf(&VideoCounters{}) {
		panic(fmt.Sprintf("%s %s", reflect.TypeOf(counter).String(), reflect.TypeOf(&PacketCounters{}).String()))
	}
}

func BenchmarkInitialization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &PacketCounters{}
	}
}

func BenchmarkInitializationReflection(b *testing.B) {
	ac := AvailableCounters{}
	counters := []string{"LatencyJitterCounter", "PacketCounters", "TCPState", "VideoCounters"}
	if _, err := ac.Build(counters); err != nil {
		panic(err)
	}
	t := ac.registryByName["PacketCounters"]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reflect.New(t).Elem().Addr().Interface()
	}
}

func BenchmarkInitializationReflectionWithRegistry(b *testing.B) {
	ac := AvailableCounters{}
	counters := []string{"LatencyJitterCounter", "PacketCounters", "TCPState", "VideoCounters"}
	if _, err := ac.Build(counters); err != nil {
		panic(err)
	}
	code := ac.nameToId["PacketCounters"]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ac.InstantiateById(code)
	}
}
