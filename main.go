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

func main() {
	if len(os.Args[1:]) == 0 {
		fmt.Println("provide a command to run with secrets")
		return
	}

	f, err := os.Open("./.env.secrets")
	if err != nil {
		log.Fatalf("error reading .env.secrets file: %s", err)
	}
	defer f.Close()

	if err := injectSecrets(f, adapters.GetSecretValue, os.Setenv); err != nil {
		log.Fatalf("error injecting secrets: %s", err)
	}

	sysCommand := newSysCommand(os.Args)
	if err := sysCommand.Run(); err != nil {
		log.Fatalf("error executing command: %s", err)
	}
}

type getSecretValueFunc func(string) (*string, error)
type setEnvFunc func(string, string) error

// injectSecrets reads key=secretID pairs from r, fetches each secret, and sets
// the corresponding environment variable via setEnv. JSON map secrets are
// expanded into one env var per key. Empty lines and lines starting with '#'
// are skipped. Fetch failures are logged and skipped; a setEnv failure is fatal.
func injectSecrets(r io.Reader, getSecretValue getSecretValueFunc, setEnv setEnvFunc) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			fmt.Printf("skipping malformed line: %q\n", line)
			continue
		}
		envKey, secretID := parts[0], parts[1]

		secretValue, err := getSecretValue(secretID)
		if err != nil {
			fmt.Printf("error retrieving secret with id %s; skipping...\n", secretID)
			continue
		}

		var js map[string]string
		if err = json.Unmarshal([]byte(*secretValue), &js); err == nil {
			for k, v := range js {
				if err = setEnv(k, v); err != nil {
					return fmt.Errorf("error writing env var: %w", err)
				}
			}
		} else {
			if err = setEnv(envKey, *secretValue); err != nil {
				return fmt.Errorf("error writing env var: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading secrets file: %w", err)
	}

	return nil
}

// newSysCommand builds an exec.Cmd from osArgs, wiring stdin/stdout/stderr to
// the current process so the child behaves like a normal shell child.
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
