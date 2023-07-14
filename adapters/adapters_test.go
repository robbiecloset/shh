package adapters

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
)

var secretValue string = "wooper"

type dummyClient struct{}

func (c dummyClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	key := params.SecretId

	if *key == "fail" {
		return &secretsmanager.GetSecretValueOutput{}, errors.New("failure to fetch secret")
	}

	return &secretsmanager.GetSecretValueOutput{
		SecretString: &secretValue,
	}, nil
}

func TestGetSecretValue(t *testing.T) {
	assert := assert.New(t)
	client = dummyClient{}

	s, err := GetSecretValue("fail")

	assert.Empty(s, "string should be empty")
	assert.NotNil(err, "error should not be nil")

	s, err = GetSecretValue("something")

	assert.Equal(*s, secretValue, "secret value should be returned")
	assert.Nil(err, "error should be nil")
}
