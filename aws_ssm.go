// checkenv plugin that provides the environment variables defined in the checkenv process.

package main

import (
	"context"

	"github.com/bugout-dev/checkenv/aws_ssm"
)

func AWSSystemsManagerParameterStoreProvider(filter string) (map[string]string, error) {
	environment := make(map[string]string)

	AWSSystemsManagerFlags := aws_ssm.Flags{
		Export:     false,
		MaxResults: 10,
		Outfile:    "",
		Update:     false,
	}
	AWSSystemsManagerFlags.FilterTags = aws_ssm.ParseFilterTags(filter)

	ctx := context.Background()
	api := aws_ssm.InitAWSClient(ctx)
	keys := aws_ssm.FetchKeysOfParameters(ctx, api, AWSSystemsManagerFlags)
	keyChunks := aws_ssm.GenerateChunks(keys, 10)
	parameters := aws_ssm.FetchParameters(ctx, api, keyChunks, AWSSystemsManagerFlags)
	for _, parameter := range parameters {
		environment[parameter.Name] = parameter.Value
	}

	return environment, nil
}

func init() {
	helpString := "Provides environment variables defined in AWS Systems Manager Parameter Store."
	RegisterPlugin("aws_ssm", helpString, noop, AWSSystemsManagerParameterStoreProvider)
}
