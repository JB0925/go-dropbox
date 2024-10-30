package api

import (
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
