package network

// "errors"
// "github.com/stretchr/testify/assert"

// //Tests pfring initialization.
// //Assumes that an interface (other than loopback) is available.
// func TestGeneralPFRingInitialization(t *testing.T) {
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

// 	ni.NewRingInterface(iface.Name, "", "host", 1500)
// }

// func TestDNSPFRingInitialization(t *testing.T) {
// 	ni := new(NetworkInterface)

// 	if name, err := GetFirstInterface(); err != nil {
// 		t.Error(err)
// 	} else {
// 		ni.NewRingInterface(name, DNSFilter, "host", 1500)
// 	}

// }

// func TestTrafficPFRingInitialization(t *testing.T) {
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

// 	ni.NewRingInterface(iface.Name, NotDNSFilter, "host", 1500)
// }

// func TestClusteredPFRingInitialization(t *testing.T) {
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

// 	ni.NewClusteredRingInterface(iface.Name, "", "host", 1500, 1, pfring.ClusterPerFlow5Tuple)
// }

// //Tests pfring initialization.
// //Assumes that an interface (other than loopback) is available.
// func TestGeneralPcapInitialization(t *testing.T) {
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

// 	ni.NewPcapInterface(iface.Name, "", "host", 1500)
// }

// func TestDNSPcapInitialization(t *testing.T) {
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

// 	ni.NewPcapInterface(iface.Name, DNSFilter, "host", 1500)
// }

// func TestTrafficPcapInitialization(t *testing.T) {
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

// 	ni.NewPcapInterface(iface.Name, NotDNSFilter, "host", 1500)
// }
