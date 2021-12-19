package parameters

// Parameter structure for storing final result from AWS SSM
type Parameter struct {
	Name   string
	Value  string
	Export string
}

// Contains command-line flags defined by user
type Flags struct {
	Export     bool
	MaxResults int
	Outfile    string
	ProductTag string
	Update     bool
}
