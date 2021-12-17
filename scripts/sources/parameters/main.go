/*
Based on: https://github.com/awsdocs/aws-doc-sdk-examples/blob/main/gov2/ssm/GetParameter/GetParameterv2.go
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// SSMGetParametersAPI defines the interface for the GetParameters function.
// We use this interface to test the function using a mocked service.
type SSMGetParametersAPI interface {
	GetParameters(
		ctx context.Context,
		params *ssm.GetParametersInput,
		optFns ...func(*ssm.Options),
	) (*ssm.GetParametersOutput, error)
}

type SSMDescribeParametersAPI interface {
	DescribeParameters(
		ctx context.Context,
		params *ssm.DescribeParametersInput,
		optFns ...func(*ssm.Options),
	) (*ssm.DescribeParametersOutput, error)
}

// FindParameters retrieves an AWS Systems Manager string parameter
// Inputs:
//     c is the context of the method call, which includes the AWS Region
//     api is the interface that defines the method call
//     input defines the input arguments to the service call.
// Output:
//     If success, a GetParametersOutput object containing the result of the service call and nil
//     Otherwise, nil and an error from the call to GetParameter
func FindParameters(c context.Context, api SSMGetParametersAPI, input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	return api.GetParameters(c, input)
}

func FindParameterKeys(c context.Context, api SSMDescribeParametersAPI, input *ssm.DescribeParametersInput) (*ssm.DescribeParametersOutput, error) {
	return api.DescribeParameters(c, input)
}

// Split list of reports on nested lists
func generateChunks(flatSlice []string, chunkSize int) [][]string {
	if len(flatSlice) == 0 {
		return nil
	}

	chunks := make([][]string, 0, len(flatSlice)/chunkSize+1)

	for i, v := range flatSlice {
		if i%chunkSize == 0 {
			chunks = append(chunks, make([]string, 0, chunkSize))
		}
		chunks[len(chunks)-1] = append(chunks[len(chunks)-1], v)
	}

	return chunks
}

type Parameter struct {
	Name  string
	Value string
}

func main() {
	var maxResults int
	var productTag string
	flag.IntVar(&maxResults, "max", 3, "The maximum number of items to return for call to AWS")
	flag.StringVar(&productTag, "product", "", "Product tag")
	flag.Parse()

	if productTag == "" {
		log.Fatalln("Please specify the tag of product")
	}

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	client := ssm.NewFromConfig(cfg)

	// Set parameter filters
	filterKey := "tag:Product"
	parameterFilters := []types.ParameterStringFilter{
		{
			Key:    &filterKey,
			Values: []string{productTag},
		},
	}
	describeInput := &ssm.DescribeParametersInput{
		MaxResults:       int32(maxResults),
		ParameterFilters: parameterFilters,
	}

	var parameterKeys []string

	n := 0
	for {
		// Fetch list of parameter keys
		results, err := FindParameterKeys(context.TODO(), client, describeInput)
		if err != nil {
			log.Fatal(err)
		}
		for _, p := range results.Parameters {
			parameterKeys = append(parameterKeys, *p.Name)
		}

		// If there are no more parameters break
		if *&results.NextToken == nil {
			break
		}
		describeInput.NextToken = *&results.NextToken

		n++
		if n >= 10 {
			log.Fatal("To many iterations over DescribeParameters loop")
		}
	}

	var parameters []Parameter

	// Split slice of parameter keys to chunks by 10 (max len allowed by AWS)
	// and fetch values for required parameters
	parameterKeyChunks := generateChunks(parameterKeys, 10)
	for _, chunk := range parameterKeyChunks {
		getInput := &ssm.GetParametersInput{
			Names: chunk,
		}
		results, err := FindParameters(context.TODO(), client, getInput)
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range results.Parameters {
			parameters = append(parameters, Parameter{Name: *p.Name, Value: *p.Value})
		}
	}
	fmt.Println(parameters)
}
