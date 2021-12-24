package servicemap

import (
	"regexp"

	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
	"github.com/traffic-refinery/traffic-refinery/internal/aho_corasick"
)

// CName contains the timeout information
type CName struct {
	Expire int64
}

type Domain struct {
	// List of string domains to match
	match *aho_corasick.AhoCorasick
	//
	services []ServiceID
}

type Pattern struct {
	regex *regexp.Regexp
	//
	services []ServiceID
}

// DNSData contains the saved entries extracted from DNS queries
type DNSMap struct {
	patterns []Pattern
	domains  []Domain
}

// NewDNSData generates a new DNSData structure
func NewDNSMap() (*DNSMap, error) {
	dc := &DNSMap{}
	return dc, nil
}

func (dc *DNSMap) addService(Code ServiceID, DomainsString, DomainsRegex []string) error {

	if len(DomainsString) > 0 {
		stringMatch := new(aho_corasick.AhoCorasick)
		stringMatch.NewAhoCorasick()
		for _, ds := range DomainsString {
			stringMatch.AddString(ds, ds)
		}
		stringMatch.Failure()
		dc.domains = append(dc.domains, Domain{match: stringMatch, services: []ServiceID{Code}})
	}

	if len(DomainsRegex) > 0 {
		for _, dr := range DomainsRegex {
			if r, err := regexp.Compile(dr); err == nil {
				dc.patterns = append(dc.patterns, Pattern{regex: r, services: []ServiceID{Code}})
			} else {
				return err
			}
		}
	}

	return nil
}

func (dc *DNSMap) addServices(services []Service) error {
	for _, s := range services {
		if err := dc.addService(s.Code, s.ServiceFilter.DomainsString, s.ServiceFilter.DomainsRegex); err != nil {
			log.Errorf("DNSMap error: %s\n", err)
			return err
		}
	}
	return nil
}

// ParseDNSResponseFirstMatch matches a DNS response to the configured services.
// Returns the first matching entry.
// First tries to match by domain, then by regex, and finally by IP address.
func (dc *DNSMap) ParseDNSResponseFirstMatch(dns layers.DNS) (string, string, []ServiceID, bool, int64) {
	ip := ""
	domain := ""
	services := []ServiceID{}
	ttl := int64(0)
	found := false

	// TODO: handle all answers
	for _, a := range dns.Answers {
		if a.IP != nil {
			log.Debugf("Adding DNS entry for IP %s\n", a.IP.String())
			ip = a.IP.String()
			ttl = int64(a.TTL)
			break
		}
	}

	if ip == "" {
		log.Debugf("No IP in DNS answer\n")
		return ip, domain, services, found, ttl
	}

	for _, q := range dns.Questions[:1] { // Assuming there's only one query.
		for _, s := range dc.domains {
			acMatch := s.match.FirstMatch(string(q.Name))
			if len(acMatch) > 0 {
				found = true
				services = append(services, s.services...)
				domain = acMatch[0]
				log.Debugf("Adding ip %s for service %d\n", ip, services[0])
				return ip, domain, services, found, ttl
			}
		}

		for _, r := range dc.patterns {
			if r.regex.MatchString(string(q.Name)) {
				found = true
				services = append(services, r.services...)
				domain = r.regex.String()
				log.Debugf("Adding ip %s for service %d\n", ip, services[0])
				return ip, domain, services, found, ttl
			}
		}

		// TODO If not found we can search for the canonical name too...
	}
	log.Debugf("IP %s has no service match\n", ip)
	return ip, domain, services, found, ttl
}

// ParseDNSResponseAllMatches matches a DNS response to the configured services.
// Returns all matching entries.
// First tries to match by domain, then by regex, and finally by IP address.
func (dc *DNSMap) ParseDNSResponseAllMatches(dns layers.DNS, pTs int64) (string, []string, []ServiceID, bool, int64) {
	ip := ""
	domain := []string{}
	services := []ServiceID{}
	ttl := int64(0)
	found := false

	// TODO: handle all answers
	for _, a := range dns.Answers {
		if a.IP != nil {
			log.Debugf("Adding DNS entry for IP %s", a.IP.String())
			ip = a.IP.String()
			ttl = int64(a.TTL)
			break
		}
	}

	if ip == "" {
		return ip, domain, services, found, ttl
	}

	for _, q := range dns.Questions[:1] { // Assuming there's only one query.
		for _, s := range dc.domains {
			acMatch := s.match.FirstMatch(string(q.Name))
			if len(acMatch) > 0 {
				found = true
				services = append(services, s.services...)
				domain = append(domain, acMatch[0])
			}
		}

		for _, r := range dc.patterns {
			if r.regex.MatchString(string(q.Name)) {
				found = true
				services = append(services, r.services...)
				domain = append(domain, r.regex.String())
				return ip, domain, services, found, ttl
			}
		}

		// TODO If not found we can search for the canonical name too...
	}
	return ip, domain, services, found, ttl
}
