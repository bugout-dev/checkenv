// checkenv plugin that provides environment variables defined in AWS System Manager Parameter Store.

package main

import (
	"context"

	"github.com/bugout-dev/checkenv/aws_ssm"
)

func AWSSystemsManagerParameterStoreProvider(filter string) (map[string]string, error) {
	environment := make(map[string]string)

	// Convert string of tags for filter to key:value structure
	filterTags := aws_ssm.ParseFilterTags(filter)

	ctx := context.Background()

	api := aws_ssm.InitAWSClient(ctx)

	keys := aws_ssm.FetchKeysOfParameters(ctx, api, filterTags)

	// Split slice of parameter keys to chunks by 10 (max len allowed by AWS)
	// and fetch values for required parameters
	keyChunks := aws_ssm.GenerateChunks(keys, 10)
	parameters := aws_ssm.FetchParameters(ctx, api, keyChunks)

	for _, parameter := range parameters {
		environment[parameter.Name] = parameter.Value
	}

	return environment, nil
}

func init() {
	helpString := "Provides environment variables defined in AWS Systems Manager Parameter Store."
	RegisterPlugin("aws_ssm", helpString, AWSSystemsManagerParameterStoreProvider)
}
