package adapters

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
)

type dummyClient struct{}

func (c dummyClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	key := params.SecretId

	if *key == "fail" {
		return &secretsmanager.GetSecretValueOutput{}, errors.New("failure to fetch secret")
	}

	v := "wooper"
	return &secretsmanager.GetSecretValueOutput{
		SecretString: &v,
	}, nil
}

func TestGetSecretValue(t *testing.T) {
	client = dummyClient{}
	t.Cleanup(func() { client = nil })

	assert := assert.New(t)

	s, err := GetSecretValue("fail")
	assert.Empty(s, "string should be empty")
	assert.NotNil(err, "error should not be nil")

	s, err = GetSecretValue("something")
	assert.Equal("wooper", *s, "secret value should be returned")
	assert.Nil(err, "error should be nil")
}
