package servicemap

type ServiceID uint16

type Filter struct {
	// DomainsString is the list of domains to match
	DomainsString []string
	// DomainsRegex is the list of regexes to match
	DomainsRegex []string
	// Prefixes is the list of subnets to match
	Prefixes []string
}

type Service struct {
	// Name of the service
	Name string
	// Code of the service
	Code ServiceID
	// ServiceFilter is the ensemble of filters to use for matching
	ServiceFilter Filter
}
