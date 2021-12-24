package servicemap

import (
	"errors"
	"time"

	"github.com/google/gopacket/layers"
)

const (
	// NotFoundEntryTimeout is the expire time to recheck for IPs not found
	NotFoundEntryTimeout int64 = 60 * 60
)

// ServiceMap contains all the data structures required to support service mappings
type ServiceMap struct {
	// services
	services []*Service
	// idToService
	idToService map[ServiceID]*Service
	// nameToService
	nameToService map[string]*Service
	// ipCache contains cached IP to Service mappings
	ipCache *IPCache
	// ipMap is the map from network prefixes to services
	ipMap *IPMap
	// ipMap is the map from dns domains to services
	dnsMap *DNSMap
}

// NewServiceMap generates a new ServiceMap structure
func NewServiceMap(cleanupTime, evictTime time.Duration) (*ServiceMap, error) {
	sm := &ServiceMap{}
	var err error

	// Initialize maps

	if sm.ipMap, err = NewIPMap(); err != nil {
		return nil, err
	}

	if sm.dnsMap, err = NewDNSMap(); err != nil {
		return nil, err
	}

	if sm.ipCache, err = NewIPCache(cleanupTime, evictTime); err != nil {
		return nil, err
	}

	sm.idToService = make(map[ServiceID]*Service)
	sm.nameToService = make(map[string]*Service)

	return sm, nil
}

//
func (sm *ServiceMap) ConfigServiceMap(services []Service) error {

	for _, service := range services {
		s := service
		sm.services = append(sm.services, &s)
		if _, found := sm.idToService[s.Code]; found {
			return errors.New("can not use twice the same service ID")
		}
		sm.idToService[s.Code] = &s
		if _, found := sm.nameToService[s.Name]; found {
			return errors.New("can not use twice the same service name")
		}
		sm.nameToService[service.Name] = &s
	}

	if err := sm.ipMap.addServices(services); err != nil {
		return err
	}

	if err := sm.dnsMap.addServices(services); err != nil {
		return err
	}

	return nil
}

// ParseDNSResponse matches a DNS response to the configured services.
// First tries to match by domain, then by regex, and finally by IP address.
func (sm *ServiceMap) ParseDNSResponse(dns layers.DNS) {
	if ip, _, services, found, ttl := sm.dnsMap.ParseDNSResponseFirstMatch(dns); found {
		sm.ipCache.Insert(ip, services, ttl)
	}
}

// Lookup allows to lookup entries in the cache map
func (sm *ServiceMap) LookupIP(ip string) ([]ServiceID, bool) {
	// If not, check if in the prefixes
	if services, ok := sm.ipCache.Lookup(ip); ok {
		if len(services) > 0 {
			return services, true
		} else {
			return services, false
		}

	} else if services, found := sm.ipMap.checkPrefixFirstMatch(ip); found {
		sm.ipCache.Insert(ip, services, 0)
		return services, true
	} else {
		sm.ipCache.Insert(ip, services, 0)
		return services, false
	}
}

func (sm *ServiceMap) GetName(id ServiceID) (string, bool) {
	if s, ok := sm.idToService[id]; ok {
		return s.Name, true
	} else {
		return "", false
	}
}

func (sm *ServiceMap) GetId(name string) (ServiceID, bool) {
	if s, ok := sm.nameToService[name]; ok {
		return s.Code, true
	} else {
		return 0, false
	}
}

func (sm *ServiceMap) GetService(id ServiceID) (*Service, bool) {
	if s, ok := sm.idToService[id]; ok {
		return s, true
	} else {
		return nil, false
	}
}
