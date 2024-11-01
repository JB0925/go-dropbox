package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	baseUrl = "http://localhost:8080"
	signupEndpoint = "/signup"
	loginEndpoint = "/login"
	createProjectEndpoint = "/projects/create"
	viewProjectEndpoint = "/projects/view"
	defaultContentType = "application/json"
	testUsername = "helloworld"
	testPassword = "Testing123!"
	contentTypeHeaderKey = "content-type"
	authHeaderKey = "Authorization"
	httpVerbGet = "GET"
	httpVerbPost = "POST"
)

//lint:ignore U1000 Ignore unused function in case of potential future use
func startServer() {
	cmd := exec.Command("go", "run", "..")
	if err := cmd.Run(); err != nil {
		msg := fmt.Sprintf("got error: %s", err.Error())
		panic(msg)
	}
}

//lint:ignore U1000 Ignore unused function in case of potential future use
func killServer(pid string) {
	cmd := exec.Command("sh", "-c", "kill -9 $(pgrep go-dropbox)")
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

//lint:ignore U1000 Ignore unused function in case of potential future use
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

//lint:ignore U1000 Ignore unused function in case of potential future use
func setupServerForTests() func(*testing.T) {
	go startServer()
	time.Sleep(500 * time.Millisecond)
	return stopServerAfterTest
}

func marshalRequestBody(m map[string]string, t *testing.T) []byte {
	b, err := json.Marshal(m)
	assert.Nil(t, err)
	return b
}

func unMarshalResponseBody(b []byte, t *testing.T) map[string]interface{} {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	assert.Nil(t, err)
	return m
}

func signupOrLoginTestUser(t *testing.T, username, password, mode string, expectToken bool) (string, int) {
	c := http.Client{}
	url := baseUrl+mode
	body := marshalRequestBody(map[string]string{
		"username": username,
		"password": password,
	}, t)
	contentType := defaultContentType
	r, err := c.Post(url, contentType, bytes.NewReader(body))
	assert.Nil(t, err)
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	assert.Nil(t, err)

	var m map[string]interface{}
	var token string
	if expectToken {
		m = unMarshalResponseBody(b, t)
		tk, ok := m["token"].(string)
		assert.True(t, ok, "response body should have a token")
		token = tk
	}

	return token, r.StatusCode
}

func clearTestUser(t *testing.T, db *sql.DB) {
	query := `DELETE FROM users WHERE username = $1`
	username := "helloworld"
	_, err := db.Exec(query, username)
	assert.Nil(t, err)
}

// makeRequest used to make a generic http request
// where we can inject the method, path, body, etc.
func makeRequest(
	t *testing.T,
	method,
	path,
	token string,
	body io.Reader,
) *http.Response {
	c := http.Client{}
	url := baseUrl+path
	contentType := defaultContentType
	req, err := http.NewRequest(method, url, body)
	assert.Nil(t, err)
	req.Header.Set(contentTypeHeaderKey, contentType)
	if token != "" {
		req.Header.Set(authHeaderKey, token)
	}
	r, err := c.Do(req)
	assert.Nil(t, err)
	return r
}