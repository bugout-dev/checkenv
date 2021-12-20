/*
Based on: https://github.com/awsdocs/aws-doc-sdk-examples/blob/main/gov2/ssm/GetParameter/GetParameterv2.go
*/
package aws_ssm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type AWSSystemsManagerParameterStore interface {
	GetParameters(
		ctx context.Context,
		params *ssm.GetParametersInput,
		optFns ...func(*ssm.Options),
	) (*ssm.GetParametersOutput, error)

	DescribeParameters(
		ctx context.Context,
		params *ssm.DescribeParametersInput,
		optFns ...func(*ssm.Options),
	) (*ssm.DescribeParametersOutput, error)
}

// ExecGetParameters and ExecDescribeParameters retrieves an AWS Systems Manager string parameter
// Inputs:
// 		c: is the context of the method call, which includes the AWS Region
// 		api: is the interface that defines the method call
// 		input: defines the input arguments to the service call
// Output:
// 		If success, a GetParametersOutput object containing the result of the service call and nil
// 		Otherwise, nil and an error from the call to GetParameters
func ExecGetParameters(c context.Context, api AWSSystemsManagerParameterStore, input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	return api.GetParameters(c, input)
}

func ExecDescribeParameters(c context.Context, api AWSSystemsManagerParameterStore, input *ssm.DescribeParametersInput) (*ssm.DescribeParametersOutput, error) {
	return api.DescribeParameters(c, input)
}

// Load the Shared AWS Configuration (~/.aws/config)
func InitAWSClient(ctx context.Context) *ssm.Client {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalln("Failed loading AWS Configuration", err)
	}
	client := ssm.NewFromConfig(cfg)

	return client
}
