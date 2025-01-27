package aws_ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// Fill fake output data
type MockAWSSystemManagerParameterStore struct{}

func (dt MockAWSSystemManagerParameterStore) GetParameters(
	ctx context.Context,
	params *ssm.GetParametersInput,
	optFns ...func(*ssm.Options),
) (*ssm.GetParametersOutput, error) {
	parameters := []types.Parameter{}
	for _, d := range globalData {
		parameters = append(parameters, types.Parameter{
			Name:  aws.String(d.Name),
			Value: aws.String(d.Value),
		})
	}
	output := &ssm.GetParametersOutput{
		Parameters: parameters,
	}

	return output, nil
}
func (dt MockAWSSystemManagerParameterStore) DescribeParameters(
	ctx context.Context,
	params *ssm.DescribeParametersInput,
	optFns ...func(*ssm.Options),
) (*ssm.DescribeParametersOutput, error) {

	// TODO(kompotkot): How to test filters?
	parameters := []types.ParameterMetadata{
		{Name: aws.String("/test/dev/t1")},
		{Name: aws.String("/test/dev/t2")},
	}
	output := &ssm.DescribeParametersOutput{
		Parameters: parameters,
	}

	return output, nil
}

type DataTags struct {
	Product string `json:"Product"`
}

type Data struct {
	Name  string     `json:"Name"`
	Value string     `json:"Value"`
	Tags  []DataTags `json:"Tags"`
}

var globalData []Data

var globalParameterKeys []string

func populateData(t *testing.T) error {
	content, err := ioutil.ReadFile("./data_test.json")
	if err != nil {
		return err
	}

	contentStr := string(content)
	err = json.Unmarshal([]byte(contentStr), &globalData)
	if err != nil {
		return nil
	}

	return nil
}

func TestDescribeParameters(t *testing.T) {
	err := populateData(t)
	if err != nil {
		t.Fatal("Failed to populate data")
	}

	api := &MockAWSSystemManagerParameterStore{}

	filterTags := []FilterTag{{Name: "Product", Value: "test"}}

	// Test DescribeParameters
	parameterKeys := FetchKeysOfParameters(
		context.Background(),
		*api,
		filterTags,
	)
	if len(parameterKeys) != 2 {
		// TODO(kompotkot): Extract length of parameters from data.json
		t.Logf("Length of parameter keys should be 2, but got %d", len(parameterKeys))
		t.Fail()
	}

	for _, p := range parameterKeys {
		globalParameterKeys = append(globalParameterKeys, p)
	}
}

func TestGetParameters(t *testing.T) {
	parameterKeyChunks := GenerateChunks(globalParameterKeys, 10)

	api := &MockAWSSystemManagerParameterStore{}

	parameters := FetchParameters(
		context.Background(),
		*api,
		parameterKeyChunks,
	)
	if len(parameters) != 2 {
		// TODO(kompotkot): Extract length of parameters from data.json
		t.Logf("Length of parameters should be 2, but got %d", len(parameters))
		t.Fail()
	}
	fmt.Println(parameters)
}
