package flowstats

type Service struct {
	// Name of the service
	Name string
	// Collect is the list of counters to collect in string format
	Collect []string
}
