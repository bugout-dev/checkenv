package aws_ssm

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// Fetch values for parameters
// Inputs:
//		chunks: list of lists with parameter key values
func FetchParameters(ctx context.Context, api AWSSystemsManagerParameterStoreAPI, chunks [][]string) []Parameter {
	var parameters []Parameter

	for _, chunk := range chunks {
		getInput := &ssm.GetParametersInput{
			Names: chunk,
		}
		results, err := ExecGetParameters(ctx, api, getInput)
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range results.Parameters {
			parameter := Parameter{
				Name: *p.Name, Value: *p.Value,
			}
			parameters = append(parameters, parameter)
		}
	}
	log.Println("Retrieved values for parameters")

	return parameters
}

// Fetch list of parameter keys from AWS with defined filters
func FetchKeysOfParameters(
	ctx context.Context,
	api AWSSystemsManagerParameterStoreAPI,
	filterTags []FilterTag,
) []string {
	var parameters []string

	// Set parameter filters
	parameterFilters := []types.ParameterStringFilter{}
	for _, ft := range filterTags {
		filterKey := fmt.Sprintf("tag:%s", ft.Name)
		parameterFilters = append(parameterFilters, types.ParameterStringFilter{
			Key:    &filterKey,
			Values: []string{ft.Value},
		})
	}
	describeInput := &ssm.DescribeParametersInput{
		MaxResults:       10,
		ParameterFilters: parameterFilters,
	}

	// CHECKENV_AWS_FETCH_LOOP_LIMIT by default set to 10,
	// it is allows to load 100 parameters from AWS and it is
	// a limiter to prevent loading too many parameters without
	// control during passing erroneous filters
	var err error
	var fetchLoopLimit int
	fetchLoopLimitStr := os.Getenv("CHECKENV_AWS_FETCH_LOOP_LIMIT")
	if fetchLoopLimitStr != "" {
		fetchLoopLimit, err = strconv.Atoi(fetchLoopLimitStr)
	}
	if fetchLoopLimitStr == "" || err != nil {
		fetchLoopLimit = 10
	}

	n := 0
	for {
		// Fetch list of parameter keys
		results, err := ExecDescribeParameters(ctx, api, describeInput)
		if err != nil {
			log.Fatal(err)
		}
		for _, p := range results.Parameters {
			parameters = append(parameters, *p.Name)
		}

		// If there are no more parameters break
		if results.NextToken == nil {
			break
		}
		describeInput.NextToken = results.NextToken

		n++
		if n >= fetchLoopLimit {
			log.Fatal("To many iterations over DescribeParameters loop")
		}
	}
	log.Printf("Retrieved %d parameters", len(parameters))

	return parameters
}

// Split list of reports on nested lists
func GenerateChunks(flatSlice []string, chunkSize int) [][]string {
	if len(flatSlice) == 0 {
		return [][]string{}
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

// ParseFilterTags convert string from user input to key value structure
func ParseFilterTags(filterTagsStr string) []FilterTag {
	var filterTags []FilterTag

	filterTagsSlice := strings.Split(filterTagsStr, ",")
	for _, t := range filterTagsSlice {
		tagNameValue := strings.Split(t, ":")
		if len(tagNameValue) != 2 || len(tagNameValue[0]) == 0 || len(tagNameValue[1]) == 0 {
			log.Printf("Unable to parse tag name and value: %s", t)
			continue
		}
		filterTags = append(filterTags, FilterTag{
			Name:  tagNameValue[0],
			Value: tagNameValue[1],
		})
	}

	return filterTags
}
