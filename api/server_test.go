package api

import (
	"testing"
	"time"
)

func TestServerSetupAndTeardown(t *testing.T) {
	go startServer()
	time.Sleep(3 * time.Second)
	stopServerAfterTest(t)
}
