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
		expectToken bool
		want int // the http status code
	}

	tests := []args{
		{
			name: "Test should return 400 when username is too short",
			username: "abc",
			password: "defghijkl",
			expectToken: false,
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when password is too short",
			username: "abcde",
			password: "fgh",
			expectToken: false,
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when username is empty",
			username: "",
			password: "abcdefghi",
			expectToken: false,
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when password is empty",
			username: "abcdefghi",
			password: "",
			expectToken: false,
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 409 when user already exists",
			username: "jesseb",
			password: "testing123",
			expectToken: false,
			want: http.StatusConflict,
		},
		{
			name: "Test should return 401 when data is valid and user does not exist",
			username: testUsername,
			password: testPassword,
			expectToken: true,
			want: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, statusCode := signupOrLoginTestUser(t, tt.username, tt.password, signupEndpoint, tt.expectToken)
			got := statusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("error code should be %d", tt.want))

			if tt.want == http.StatusCreated {
				assert.NotEmpty(t, token, "token should not be empty after successful call")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, false)
	defer clearTestUser(t, signupManager.db)

	type args struct {
		name string
		username string
		password string
		expectToken bool
		want int // http status code
	}

	tests := []args{
		{
			name: "Test should return 400 when username is too short",
			username: "abc",
			password: "defghijkl",
			expectToken: false,
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when password is too short",
			username: "abcde",
			password: "fgh",
			expectToken: false,
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when username is empty",
			username: "",
			password: "abcdefghi",
			expectToken: false,
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 400 when password is empty",
			username: "abcdefghi",
			password: "",
			expectToken: false,
			want: http.StatusBadRequest,
		},
		{
			name: "Test should return 200 when user data is valid",
			username: testUsername,
			password: testPassword,
			expectToken: true,
			want: http.StatusOK,
		},
		{
			name: "Test should return 401 when data is valid but user does not exist",
			username: "helloworldz",
			password: testPassword,
			expectToken: false,
			want: http.StatusUnauthorized,
		},
		{
			name: "Test should return 401 when username is correct but password is not",
			username: testUsername,
			password: testPassword+"foobar",
			expectToken: false,
			want: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, statusCode := signupOrLoginTestUser(t, tt.username, tt.password, loginEndpoint, tt.expectToken)
			got := statusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected status code %d, but got %d", tt.want, got))

			if tt.want == http.StatusOK {
				assert.NotEmpty(t, token, "token should not be empty after successful call")
			}
		})
	}
}

func TestCreateProject(t *testing.T) {
	token, _ := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, true)
	defer clearTestUser(t, signupManager.db)
	assert.NotEmpty(t, token, "response body should have a token")

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
			r := makeRequest(t, "POST", createProjectEndpoint, tt.token, tt.requestBody)
			defer r.Body.Close()
			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected %d but got %d", tt.want, got))
		})
	}
}

func TestViewProject(t *testing.T) {
	token, _ := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, true)
	defer clearTestUser(t, signupManager.db)

	requestBody := bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`))
	r := makeRequest(t, httpVerbPost, createProjectEndpoint, token, requestBody)
	defer r.Body.Close()
	got := r.StatusCode
	assert.Equal(t, http.StatusCreated, got, fmt.Sprintf("expected %d but got %d", http.StatusCreated, got))
	
	type args struct {
		name string
		token string
		queryString string
		want int
	}

	tests := []args{
		{
			name: "Test should return 200 when request has a valid token and has a project with the correct name",
			token: token,
			queryString: "?project_name=foobar",
			want: http.StatusOK,
		},
		{
			name: "Test should return 404 when request is valid but project does not exist for user",
			token: token,
			queryString: "?project_name=baz",
			want: http.StatusNotFound,
		},
		{
			name: "Test should return 401 when token is invalid",
			token: token+"foobar123",
			queryString: "?project_name=foobar",
			want: http.StatusUnauthorized,
		},
		{
			name: "Test should return 401 when token is empty",
			token: "",
			queryString: "?project_name=foobar",
			want: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := makeRequest(t, httpVerbGet, viewProjectEndpoint+tt.queryString, tt.token, bytes.NewReader([]byte{}))
			defer r.Body.Close()
			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected %d status code but got %d", tt.want, got))

			if got == http.StatusOK {
				b, err := io.ReadAll(r.Body)
				assert.Nil(t, err)
				m := unMarshalResponseBody(b, t)
				p, ok := m["project"].(string)
				assert.True(t, ok)

				m = unMarshalResponseBody([]byte(p), t)
				_, exists := m["root"].(map[string]interface{})["bar"]
				assert.True(t, exists)
			}
		})
	}
}
