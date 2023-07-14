package adapters

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var client fetchClient

type fetchClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func GetSecretValue(k string) (*string, error) {
	if client == nil {
		initClient()
	}

	secretValue, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(k),
	})

	if err != nil {
		return nil, err
	}

	return secretValue.SecretString, nil
}

func initClient() {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	client = secretsmanager.NewFromConfig(cfg)
}
