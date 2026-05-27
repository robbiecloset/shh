// Package adapters provides a wrapper around AWS Secrets Manager.
package adapters

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var (
	mu     sync.Mutex
	client fetchClient
)

type fetchClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// GetSecretValue fetches the secret string stored at the given secret ID from
// AWS Secrets Manager. The AWS client is initialized lazily on first call.
func GetSecretValue(k string) (*string, error) {
	c, err := ensureClient()
	if err != nil {
		return nil, err
	}

	out, err := c.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(k),
	})
	if err != nil {
		return nil, fmt.Errorf("fetching secret %q: %w", k, err)
	}
	if out.SecretString == nil {
		return nil, fmt.Errorf("secret %q has no string value (binary secrets are not supported)", k)
	}

	return out.SecretString, nil
}

func ensureClient() (fetchClient, error) {
	mu.Lock()
	defer mu.Unlock()
	if client != nil {
		return client, nil
	}
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	client = secretsmanager.NewFromConfig(cfg)
	return client, nil
}
