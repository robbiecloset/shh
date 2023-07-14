package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var cmd string = "wooper"
var arg1 string = "face"
var arg2 string = "something"

func TestNewSysCommand(t *testing.T) {
	assert := assert.New(t)
	c := newSysCommand([]string{"/path/to/exe", cmd})
	assert.Equal(c.Path, cmd, "command should be equal")

	c = newSysCommand([]string{"/path/to/exe", cmd, arg1, arg2})
	assert.Equal(c.Path, cmd, "command should be equal")
	assert.Contains(c.Args, arg1, "should contain arg1")
	assert.Contains(c.Args, arg2, "should contain arg2")
}
