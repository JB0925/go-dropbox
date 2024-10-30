package api

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func startServer() {
	cmd := exec.Command("go", "run", "..")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func killServer(pid string) {
	cmd := exec.Command("sh", "-c", "kill -9 "+pid)
	if err := cmd.Run(); err != nil {
        // Check if the error is an ExitError due to the process being killed - this is expected
        exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.Sys().(syscall.WaitStatus).ExitStatus() == 9 {
            log.Default().Printf("Process %s killed successfully\n", pid)
			return
        }

		panic(fmt.Errorf("server did not stop"))
	}
}

func stopServerAfterTest(t *testing.T) {
	var buf bytes.Buffer
    var pid string

    // Retry until we find the process or reach a timeout
    for retries := 5; retries > 0; retries-- {
        cmd := exec.Command("pgrep", "go-dropbox")
        cmd.Stdout = &buf

        if err := cmd.Run(); err == nil {
            pid = buf.String()
            break
        }

        // Clear the buffer and retry after a short delay
        buf.Reset()
        time.Sleep(500 * time.Millisecond)
    }

    if pid == "" {
        t.Fatal("could not find server PID")
    }

    log.Default().Printf("PID: %s", pid)
	killServer(pid)
}