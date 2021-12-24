package network

// "errors"
// "github.com/stretchr/testify/assert"

// func TestTrafficParserFromRing(t *testing.T) {
// 	ni := new(NetworkInterface)
// 	ifaces, err := net.Interfaces()
// 	if err != nil {
// 		panic(err)
// 	}
// 	iface := net.Interface{Index: -1}
// 	for _, i := range ifaces {
// 		addrs, _ := i.Addrs()
// 		for _, addr := range addrs {
// 			switch v := addr.(type) {
// 			case *net.IPNet:
// 				if v.IP.IsGlobalUnicast() && v.IP.To4() != nil && v.IP.String() != "1.2.3.4" {
// 					iface = i
// 				}
// 			}
// 		}
// 	}

// 	print("Testing Traffic parser on interface " + iface.Name + "\n")
// 	ni.NewRingInterface(iface.Name, NotDNSFilter, "host", 1500)

// 	dp := new(TrafficParser)
// 	dp.NewTrafficParser(ni, nil)

// 	stop := make(chan struct{})
// 	go dp.Parse(nil, stop)
// 	select {
// 	case <-time.After(10 * time.Second):
// 		print("something" + "\n")
// 		stop <- struct{}{}
// 	}
// }
