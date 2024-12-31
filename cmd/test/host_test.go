package test

import (
	"testing"
	"wsystemd/cmd/process"
)

func TestHost(t *testing.T) {
	t.Log(process.GetHostName())
}
