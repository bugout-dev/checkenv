package aws_ssm

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// Fetch values for parameters
// Inputs:
//		chunks: list of lists with parameter key values
func FetchParameters(ctx context.Context, api SSMGetParametersAPI, chunks [][]string, flags Flags) []Parameter {
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
			if flags.Export {
				parameter.Export = "export "
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
	api SSMDescribeParametersAPI,
	flags Flags,
) []string {
	var parameters []string

	// Set parameter filters
	parameterFilters := []types.ParameterStringFilter{}
	for _, ft := range flags.FilterTags {
		filterKey := fmt.Sprintf("tag:%s", ft.Name)
		parameterFilters = append(parameterFilters, types.ParameterStringFilter{
			Key:    &filterKey,
			Values: []string{ft.Value},
		})
	}
	describeInput := &ssm.DescribeParametersInput{
		MaxResults:       int32(flags.MaxResults),
		ParameterFilters: parameterFilters,
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
		describeInput.NextToken = *&results.NextToken

		n++
		if n >= 50 {
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

// WriteToFile generate or update existing file and
// flash to it environment variables
func WriteToFile(parameters []Parameter, outfile string, update bool, export bool) {
	flag := os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	if update {
		flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	}

	f, err := os.OpenFile(
		outfile,
		flag,
		0644,
	)
	if err != nil {
		log.Fatalf("Unable to open file %s, error: %s", outfile, err)
	}
	defer f.Close()

	parametersStr := ""
	for _, p := range parameters {
		parametersStr += fmt.Sprintf("%s%s=%s\n", p.Export, p.Name, p.Value)
	}
	if _, err := f.WriteString(parametersStr); err != nil {
		log.Fatalf("Unable to write to file %s, error: %s", outfile, err)
	}
}

// HandleSignals process Ctrl+C and all script interruptions
func HandleSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	for {
		sig := <-sigCh
		switch sig {
		case os.Interrupt:
			cancel()
			return
		}
	}
}

func Extract() {
	var flags Flags
	var filterTagsStr string
	flag.BoolVar(&flags.Export, "export", false, "Add prefix 'export' to each parameter")
	flag.IntVar(&flags.MaxResults, "max", 3, "The maximum number of items to return for call to AWS")
	flag.StringVar(&flags.Outfile, "outfile", "", "Output file where parameters will be saved")
	flag.StringVar(&filterTagsStr, "tags", "", "Product tags for filter separated by comma in format 'tagName1:tagValue1,tagName2:tagValue2'")
	flag.BoolVar(&flags.Update, "update", false, "Update existing file if exists (by default the file will be overwritten)")
	flag.Parse()

	if filterTagsStr == "" {
		log.Fatalln("Please specify the tags for filter, at least Product tag")
	}

	// Convert string of tags for filter to key:value structure
	filterTags := ParseFilterTags(filterTagsStr)
	flags.FilterTags = filterTags

	ctx, cancel := context.WithCancel(context.Background())
	go HandleSignals(cancel)

	client := InitAWSClient(ctx)

	parameterKeys := FetchKeysOfParameters(ctx, client, flags)

	// Split slice of parameter keys to chunks by 10 (max len allowed by AWS)
	// and fetch values for required parameters
	parameterKeyChunks := GenerateChunks(parameterKeys, 10)
	if len(parameterKeyChunks) == 0 {
		log.Fatalln("Nothing to generate, empty slice provided")
	}
	parameters := FetchParameters(ctx, client, parameterKeyChunks, flags)

	if flags.Outfile != "" {
		WriteToFile(parameters, flags.Outfile, flags.Update, flags.Export)
	} else {
		for _, p := range parameters {
			fmt.Printf("%s%s=%s\n", p.Export, p.Name, p.Value)
		}
	}
}
