package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSysCommand(t *testing.T) {
	assert := assert.New(t)
	c := newSysCommand([]string{"/path/to/exe", "wooper"})
	assert.Equal(c.Path, "wooper", "command should be equal")

	c = newSysCommand([]string{"/path/to/exe", "wooper", "face", "something"})
	assert.Equal(c.Path, "wooper", "command should be equal")
	assert.Contains(c.Args, "face", "should contain arg1")
	assert.Contains(c.Args, "something", "should contain arg2")
}

func TestInjectSecrets(t *testing.T) {
	plainSecret := func(id string) (*string, error) {
		v := "plain-value"
		return &v, nil
	}
	jsonSecret := func(id string) (*string, error) {
		v := `{"KEY1":"val1","KEY2":"val2"}`
		return &v, nil
	}
	failSecret := func(id string) (*string, error) {
		return nil, errors.New("fetch failed")
	}

	collectEnv := func() (map[string]string, setEnvFunc) {
		env := map[string]string{}
		return env, func(k, v string) error { env[k] = v; return nil }
	}

	t.Run("plain string secret sets env var", func(t *testing.T) {
		env, setEnv := collectEnv()
		err := injectSecrets(strings.NewReader("FOO=my/secret/id\n"), plainSecret, setEnv)
		assert.NoError(t, err)
		assert.Equal(t, "plain-value", env["FOO"])
	})

	t.Run("json map secret expands into multiple env vars", func(t *testing.T) {
		env, setEnv := collectEnv()
		err := injectSecrets(strings.NewReader("FOO=my/secret/id\n"), jsonSecret, setEnv)
		assert.NoError(t, err)
		assert.Equal(t, "val1", env["KEY1"])
		assert.Equal(t, "val2", env["KEY2"])
	})

	t.Run("empty lines are skipped", func(t *testing.T) {
		env, setEnv := collectEnv()
		err := injectSecrets(strings.NewReader("\n\nFOO=my/secret/id\n"), plainSecret, setEnv)
		assert.NoError(t, err)
		assert.Len(t, env, 1)
	})

	t.Run("comment lines are skipped", func(t *testing.T) {
		env, setEnv := collectEnv()
		err := injectSecrets(strings.NewReader("# comment\nFOO=my/secret/id\n"), plainSecret, setEnv)
		assert.NoError(t, err)
		assert.Len(t, env, 1)
	})

	t.Run("malformed line without = is skipped", func(t *testing.T) {
		env, setEnv := collectEnv()
		err := injectSecrets(strings.NewReader("MALFORMED\nFOO=my/secret/id\n"), plainSecret, setEnv)
		assert.NoError(t, err)
		assert.Equal(t, "plain-value", env["FOO"])
		_, had := env["MALFORMED"]
		assert.False(t, had)
	})

	t.Run("line with empty secret ID is skipped", func(t *testing.T) {
		env, setEnv := collectEnv()
		err := injectSecrets(strings.NewReader("FOO=\n"), plainSecret, setEnv)
		assert.NoError(t, err)
		assert.Empty(t, env)
	})

	t.Run("fetch error causes skip not failure", func(t *testing.T) {
		env, setEnv := collectEnv()
		err := injectSecrets(strings.NewReader("FOO=my/secret/id\n"), failSecret, setEnv)
		assert.NoError(t, err)
		assert.Empty(t, env)
	})

	t.Run("setEnv error is returned", func(t *testing.T) {
		failSetEnv := func(k, v string) error { return errors.New("setenv failed") }
		err := injectSecrets(strings.NewReader("FOO=my/secret/id\n"), plainSecret, failSetEnv)
		assert.Error(t, err)
	})

	t.Run("secret id with = in value is parsed correctly", func(t *testing.T) {
		var capturedID string
		getSecret := func(id string) (*string, error) {
			capturedID = id
			v := "val"
			return &v, nil
		}
		env, setEnv := collectEnv()
		err := injectSecrets(strings.NewReader("FOO=arn:aws:secretsmanager:us-east-1:123=suffix\n"), getSecret, setEnv)
		assert.NoError(t, err)
		assert.Equal(t, "arn:aws:secretsmanager:us-east-1:123=suffix", capturedID)
		assert.Equal(t, "val", env["FOO"])
	})
}
