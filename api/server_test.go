package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignup(t *testing.T) {
	// stopServer := setupServerForTests()  // uncomment these lines to run tests individually
	// defer stopServer(t)
	defer clearTestUser(t, signupManager.db)

	type args struct {
		name string
		username string
		password string
		want int // the http status code
	}

	tests := []args{
		{
			name: "Test should return 400 when username is too short",
			username: "abc",
			password: "defghijkl",
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when password is too short",
			username: "abcde",
			password: "fgh",
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when username is empty",
			username: "",
			password: "abcdefghi",
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when password is empty",
			username: "abcdefghi",
			password: "",
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 409 when user already exists",
			username: "jesseb",
			password: "testing123",
			want: http.StatusConflict,
		},
		{
			name: "Test should return 401 when data is valid and user does not exist",
			username: testUsername,
			password: testPassword,
			want: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := signupOrLoginTestUser(t, tt.username, tt.password, signupEndpoint)

			defer r.Body.Close()
			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("error code should be %d", tt.want))

			if tt.want == http.StatusCreated {
				body, err := io.ReadAll(r.Body)
				assert.Nil(t, err)
				b := unMarshalResponseBody(body, t)
				assert.Contains(t, b, "token")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	r := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint)
	defer r.Body.Close()
	defer clearTestUser(t, signupManager.db)

	type args struct {
		name string
		username string
		password string
		want int // http status code
	}

	tests := []args{
		{
			name: "Test should return 400 when username is too short",
			username: "abc",
			password: "defghijkl",
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when password is too short",
			username: "abcde",
			password: "fgh",
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when username is empty",
			username: "",
			password: "abcdefghi",
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when password is empty",
			username: "abcdefghi",
			password: "",
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 200 when user data is valid",
			username: testUsername,
			password: testPassword,
			want: http.StatusOK,
		},
		{
			name: "Test should return 401 when data is valid but user does not exist",
			username: "helloworldz",
			password: testPassword,
			want: http.StatusUnauthorized,
		},
		{
			name: "Test should return 401 when username is correct but password is not",
			username: testUsername,
			password: testPassword+"foobar",
			want: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := signupOrLoginTestUser(t, tt.username, tt.password, loginEndpoint)
			defer r.Body.Close()

			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected status code %d, but got %d", tt.want, got))

			if tt.want == http.StatusOK {
				body, err := io.ReadAll(r.Body)
				assert.Nil(t, err)
				b := unMarshalResponseBody(body, t)
				assert.Contains(t, b, "token")
			}
		})
	}
}

func TestCreateProject(t *testing.T) {
	r := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint)
	defer r.Body.Close()
	defer clearTestUser(t, signupManager.db)
	b, err := io.ReadAll(r.Body)
	assert.Nil(t, err)
	m := unMarshalResponseBody(b, t)

	token, ok := m["token"].(string)
	assert.True(t, ok, "response body should have a token")

	type args struct {
		name string
		requestBody io.Reader
		token string
		want int
	}

	tests := []args{
		{
			name: "Test should return 201 when project doesn't exist and details are valid",
			token: token,
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foo", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`)),
			want: http.StatusCreated,
		},
		{
			name: "Test should return 409 when the project already exists for a user",
			token: token,
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foo", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`)),
			want: http.StatusConflict,
		},
		{
			name: "Test should return 201 when user provides no directories to start a new project",
			token: token,
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar"}`)),
			want: http.StatusCreated,
		},
		{
			name: "Test should return 401 when user does not provide a token",
			token: "",
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar"}`)),
			want: http.StatusUnauthorized,	
		},
		{
			name: "Test should return 401 when user provides a malformed token",
			token: token+"foobar123",
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar"}`)),
			want: http.StatusUnauthorized,	
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := http.Client{}
			url := baseUrl+createProjectEndpoint
			body := tt.requestBody
			contentType := defaultContentType
			req, err := http.NewRequest("POST", url, body)
			assert.Nil(t, err)
			req.Header.Set(contentTypeHeaderKey, contentType)
			req.Header.Set(authHeaderKey, tt.token)
			r, err := c.Do(req)
			assert.Nil(t, err)
			defer r.Body.Close()
			assert.Nil(t, err)
			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected %d but got %d", tt.want, got))
		})
	}
}
