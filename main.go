package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/robbiecloset/shh/adapters"
)

var f *os.File

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

	injectSecrets(f, adapters.GetSecretValue, os.Setenv)
	sysCommand := newSysCommand(os.Args)

	err = sysCommand.Run()
	if err != nil {
		log.Fatalf("error executing command: %s\n", err)
	}
}

type getSecretValueFunc func(string) (*string, error)
type setEnvFunc func(string, string) error

func injectSecrets(r io.Reader, getSecretValue getSecretValueFunc, setEnv setEnvFunc) error {
	var kv []string
	scanner := bufio.NewScanner(r)
	fmt.Println(r)
	for scanner.Scan() {
		kv = strings.Split(scanner.Text(), "=")
		fmt.Println(kv)

		// Comment
		if string(kv[0][0]) == "#" {
			continue
		}

		secretValue, err := getSecretValue(kv[1])
		if err != nil {
			fmt.Printf("error retriving secret with id %s; skipping...\n", kv[1])
			continue
		}

		var js map[string]string
		err = json.Unmarshal([]byte(*secretValue), &js)
		if err == nil {
			for k, v := range js {
				err = setEnv(k, v)
				if err != nil {
					return fmt.Errorf("error writing env var: %w\n", err)
				}
			}
		} else {
			// Probably just a string then
			err = setEnv(kv[0], *secretValue)
			if err != nil {
				return fmt.Errorf("error writing env var: %w\n", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading secrets file: %w\n", err)
	}

	return nil
}

func newSysCommand(osArgs []string) *exec.Cmd {
	command := osArgs[1]

	var sysCommandArgs []string
	if len(osArgs) > 2 {
		sysCommandArgs = osArgs[2:]
	}

	sysCommand := exec.Command(command, sysCommandArgs...)
	sysCommand.Env = os.Environ()

	sysCommand.Stdin = os.Stdin
	sysCommand.Stdout = os.Stdout
	sysCommand.Stderr = os.Stderr

	return sysCommand
}
