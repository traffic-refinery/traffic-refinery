package servicemap

import (
	"net"
)

// Prefix
type Prefix struct {
	prefix   *net.IPNet
	services []ServiceID
}

// IPMap contains the saved entries extracted from DNS queries
type IPMap struct {
	prefixes []Prefix // List of network prefixes to match
}

// NewIPData generates a new DNSCache structure
func NewIPMap() (*IPMap, error) {
	// No need for concurrency, IP prefixes are forever
	dc := &IPMap{}

	return dc, nil
}

//
func (dc *IPMap) addService(code ServiceID, prefixes []string) error {
	for _, p := range prefixes {
		if _, n, err := net.ParseCIDR(p); err == nil {
			added := false
			for i := range dc.prefixes {
				if dc.prefixes[i].prefix.Network() == n.Network() {
					// todo add check that is not in there already
					dc.prefixes[i].services = append(dc.prefixes[i].services, code)
					added = true
				}
			}
			if !added {
				newEntry := Prefix{
					prefix:   n,
					services: []ServiceID{code},
				}
				// todo Append services
				dc.prefixes = append(dc.prefixes, newEntry)
			}
		} else {
			return err
		}
	}

	return nil
}

func (dc *IPMap) addServices(services []Service) error {
	for _, s := range services {
		if err := dc.addService(s.Code, s.ServiceFilter.Prefixes); err != nil {
			return err
		}
	}
	return nil
}

// CheckPrefixFirstMatch lookups the cached DNS entries for possible matches
// Returns first matching service
func (dc *IPMap) checkPrefixFirstMatch(sIP string) ([]ServiceID, bool) {
	ip := net.ParseIP(sIP)
	for _, entry := range dc.prefixes {
		if entry.prefix.Contains(ip) {
			return entry.services, true
		}
	}
	return nil, false
}

// CheckPrefixAllMatches lookups the cached DNS entries for possible matches
// Returns all matching services
func (dc *IPMap) checkPrefixAllMatches(sIP string) ([]ServiceID, bool) {
	services := []ServiceID{}
	ip := net.ParseIP(sIP)
	found := false
	for _, entry := range dc.prefixes {
		if entry.prefix.Contains(ip) {
			services = append(services, entry.services...)
			found = true
		}
	}
	return services, found
}
