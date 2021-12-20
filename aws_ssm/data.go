package aws_ssm

// Parameter structure for storing final result from AWS SSM
type Parameter struct {
	Name  string
	Value string
}

// Tags for filter defined by user
type FilterTag struct {
	Name  string
	Value string
}
