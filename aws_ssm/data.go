package aws_ssm

// Parameter structure for storing final result from AWS SSM
type Parameter struct {
	Name   string
	Value  string
	Export string
}

// Tags for filter defined by user
type FilterTag struct {
	Name  string
	Value string
}

// Contains command-line flags defined by user
type Flags struct {
	Export     bool
	MaxResults int
	Outfile    string
	FilterTags []FilterTag
	Update     bool
}
