package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func main() {
	if len(os.Args[1:]) == 0 {
		fmt.Println("provide a command to run with secrets")
		return
	}

	f, err := os.Open("./.env.secrets")
	if err != nil {
		log.Fatalf("error reading .env.secrets file: %s\n", err)
	}
	defer f.Close()

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	asmClient := secretsmanager.NewFromConfig(cfg)

	var kv []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		kv = strings.Split(scanner.Text(), "=")

		// Comment
		if string(kv[0][0]) == "#" {
			continue
		}

		secretValue, err := asmClient.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(kv[1]),
		})
		if err != nil {
			fmt.Printf("error retriving secret with id %s; skipping...\n", kv[1])
			continue
		}

		var js map[string]string
		err = json.Unmarshal([]byte(*secretValue.SecretString), &js)
		if err != nil {
			// Probabloy just a string then
			err = os.Setenv(kv[0], *secretValue.SecretString)
			if err != nil {
				log.Fatalf("error writing env var: %s\n", err)
			}
		} else {
			for k, v := range js {
				err = os.Setenv(k, v)
				if err != nil {
					log.Fatalf("error writing env var: %s\n", err)
				}
			}
		}
	}

	if err = scanner.Err(); err != nil {
		log.Fatalf("error reading env vars: %s\n", err)
	}

	command := os.Args[1]

	var sysCommandArgs []string
	if len(os.Args) > 2 {
		sysCommandArgs = os.Args[2:]
	}

	sysCommand := exec.Command(command, sysCommandArgs...)
	sysCommand.Env = os.Environ()

	sysCommand.Stdin = os.Stdin
	sysCommand.Stdout = os.Stdout
	sysCommand.Stderr = os.Stderr

	err = sysCommand.Run()
	if err != nil {
		log.Fatalf("error executing command: %s\n", err)
	}
}

func isJSONString(s string) bool {
	var js string
	return json.Unmarshal([]byte(s), &js) == nil

}
