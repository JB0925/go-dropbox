package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignup(t *testing.T) {
	// stopServer := setupServerForTests()  // uncomment these lines to run tests individually
	// defer stopServer(t)
	defer clearTestUser(t, signupManager.db)

	type args struct {
		name        string
		username    string
		password    string
		expectToken bool
		want        int // the http status code
	}

	tests := []args{
		{
			name:        "Test should return 400 when username is too short",
			username:    "abc",
			password:    "defghijkl",
			expectToken: false,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Test should return 400 when password is too short",
			username:    "abcde",
			password:    "fgh",
			expectToken: false,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Test should return 400 when username is empty",
			username:    "",
			password:    "abcdefghi",
			expectToken: false,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Test should return 400 when password is empty",
			username:    "abcdefghi",
			password:    "",
			expectToken: false,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Test should return 201 when data is valid and user does not exist",
			username:    testUsername,
			password:    testPassword,
			expectToken: true,
			want:        http.StatusCreated,
		},
		{
			name:        "Test should return 409 when user already exists",
			username:    testUsername,
			password:    testPassword,
			expectToken: false,
			want:        http.StatusConflict,
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
		name        string
		username    string
		password    string
		expectToken bool
		want        int // http status code
	}

	tests := []args{
		{
			name:        "Test should return 400 when username is too short",
			username:    "abc",
			password:    "defghijkl",
			expectToken: false,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Test should return 400 when password is too short",
			username:    "abcde",
			password:    "fgh",
			expectToken: false,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Test should return 400 when username is empty",
			username:    "",
			password:    "abcdefghi",
			expectToken: false,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Test should return 400 when password is empty",
			username:    "abcdefghi",
			password:    "",
			expectToken: false,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Test should return 200 when user data is valid",
			username:    testUsername,
			password:    testPassword,
			expectToken: true,
			want:        http.StatusOK,
		},
		{
			name:        "Test should return 401 when data is valid but user does not exist",
			username:    "helloworldz",
			password:    testPassword,
			expectToken: false,
			want:        http.StatusUnauthorized,
		},
		{
			name:        "Test should return 401 when username is correct but password is not",
			username:    testUsername,
			password:    testPassword + "foobar",
			expectToken: false,
			want:        http.StatusUnauthorized,
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
		name        string
		requestBody io.Reader
		token       string
		want        int // http status code
	}

	tests := []args{
		{
			name:        "Test should return 201 when project doesn't exist and details are valid",
			token:       token,
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foo", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`)),
			want:        http.StatusCreated,
		},
		{
			name:        "Test should return 409 when the project already exists for a user",
			token:       token,
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foo", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`)),
			want:        http.StatusConflict,
		},
		{
			name:        "Test should return 201 when user provides no directories to start a new project",
			token:       token,
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar"}`)),
			want:        http.StatusCreated,
		},
		{
			name:        "Test should return 401 when user does not provide a token",
			token:       "",
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar"}`)),
			want:        http.StatusUnauthorized,
		},
		{
			name:        "Test should return 401 when user provides a malformed token",
			token:       token + "foobar123",
			requestBody: bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar"}`)),
			want:        http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := makeRequest(t, "POST", createProjectEndpoint, defaultContentType, tt.token, tt.requestBody)
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
	r := makeRequest(t, http.MethodPost, createProjectEndpoint, defaultContentType, token, requestBody)
	defer r.Body.Close()
	got := r.StatusCode
	assert.Equal(t, http.StatusCreated, got, fmt.Sprintf("expected %d but got %d", http.StatusCreated, got))

	type args struct {
		name        string
		token       string
		queryString string
		want        int // http status code
	}

	tests := []args{
		{
			name:        "Test should return 200 when request has a valid token and has a project with the correct name",
			token:       token,
			queryString: "?project_name=foobar",
			want:        http.StatusOK,
		},
		{
			name:        "Test should return 404 when request is valid but project does not exist for user",
			token:       token,
			queryString: "?project_name=baz",
			want:        http.StatusNotFound,
		},
		{
			name:        "Test should return 401 when token is invalid",
			token:       token + "foobar123",
			queryString: "?project_name=foobar",
			want:        http.StatusUnauthorized,
		},
		{
			name:        "Test should return 401 when token is empty",
			token:       "",
			queryString: "?project_name=foobar",
			want:        http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := makeRequest(t, http.MethodGet, viewProjectEndpoint+tt.queryString, defaultContentType, tt.token, bytes.NewReader([]byte{}))
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

func TestDeleteProject(t *testing.T) {
	// create a user and then a project that belongs to the user
	token, _ := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, true)
	defer clearTestUser(t, signupManager.db)

	requestBody := bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`))
	r := makeRequest(t, http.MethodPost, createProjectEndpoint, defaultContentType, token, requestBody)
	defer r.Body.Close()
	got := r.StatusCode
	assert.Equal(t, http.StatusCreated, got, fmt.Sprintf("expected %d but got %d", http.StatusCreated, got))

	type args struct {
		name        string
		queryString string
		token       string
		want        int // http status code
	}

	tests := []args{
		{
			name:        "Test should return 204 when deleting a project that belongs to the user",
			queryString: "?project_name=foobar",
			token:       token,
			want:        http.StatusNoContent,
		},
		{
			name:        "Test should return 404 when deleting a project that no longer exists",
			queryString: "?project_name=foobar",
			token:       token,
			want:        http.StatusNotFound,
		},
		{
			name:        "Test should return 401 when token is malformed.",
			queryString: "?project_name=bar",
			token:       token + "foobar123",
			want:        http.StatusUnauthorized,
		},
		{
			name:        "Test should return 401 when token is missing",
			queryString: "?project_name=baz",
			token:       "",
			want:        http.StatusUnauthorized,
		},
		{
			name:        "Test should return 400 when request params are invalid",
			queryString: "?foo=bar",
			token:       token,
			want:        http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewReader([]byte("")) // no body required, but need to provide a default
			r := makeRequest(t, http.MethodDelete, deleteProjectEndpoint+tt.queryString, defaultContentType, tt.token, body)
			defer r.Body.Close()

			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected %d when deleting project, but got %d", tt.want, got))
		})
	}
}

func TestUploadFile(t *testing.T) {
	// Setup test by:
	// 1. Writing a test file to disk
	// 2. Creating a user
	// 3. Creating a project for that user
	defer clearTestUser(t, signupManager.db)
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	deleteFunc := writeFileForTesting(t, "This is a file used for testing.")
	defer deleteFunc(t, cwd+"/foo.txt")
	token, statusCode := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, true)
	assert.Equal(t, http.StatusCreated, statusCode, "Signup status code should be 201 when successful")

	requestBody := bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`))
	r := makeRequest(t, http.MethodPost, createProjectEndpoint, defaultContentType, token, requestBody)
	defer r.Body.Close()
	got := r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "expected 201 when creating a valid project")

	type args struct {
		name        string
		token       string
		fileName    string
		projectName string
		filePath    string
		want        int // http status code
	}

	tests := []args{
		{
			name:        "Test should return 201 with valid user and existing project",
			token:       token,
			fileName:    "foo.txt",
			projectName: "foobar",
			filePath:    "/root/bar",
			want:        http.StatusCreated,
		},
		{
			name:        "Test should return 400 when the path is not valid ( doesn't start with /root )",
			token:       token,
			fileName:    "bar.txt",
			projectName: "foobar",
			filePath:    "/foo/bar",
			want:        http.StatusBadRequest,
		},
		{
			name:        "Test should return 409 when the file is already uploaded",
			token:       token,
			fileName:    "foo.txt",
			projectName: "foobar",
			filePath:    "/root/bar",
			want:        http.StatusConflict,
		},
		{
			name:        "Test should return 404 when the project the file belongs to does not exist",
			token:       token,
			fileName:    "foo.txt",
			projectName: "barbaz",
			filePath:    "/root/bar",
			want:        http.StatusNotFound,
		},
		{
			name:        "Test should return 401 with a malformed token",
			token:       token + "foobar123",
			fileName:    "foo.txt",
			projectName: "foobar",
			filePath:    "/root/bar",
			want:        http.StatusUnauthorized,
		},
		{
			name:        "Test should return 401 with a nonexistent token",
			token:       "",
			fileName:    "foo.txt",
			projectName: "foobar",
			filePath:    "/root/bar",
			want:        http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := uploadOrUpdateTestFile(t, tt.token, tt.fileName, tt.projectName, uploadFilesEndpoint, tt.filePath)
			defer r.Body.Close()
			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected %d on file upload, but got %d", tt.want, got))
		})
	}
}

func TestUpdateFile(t *testing.T) {
	// Setup test by:
	// 1. Writing a test file to disk
	// 2. Creating a user
	// 3. Creating a project for that user
	defer clearTestUser(t, signupManager.db)
	deleteFunc := writeFileForTesting(t, "This file is used for testing.\n")
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer deleteFunc(t, cwd+"/foo.txt")
	token, statusCode := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, true)
	assert.Equal(t, http.StatusCreated, statusCode, "Signup status code should be 201 when successful")

	requestBody := bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`))
	r := makeRequest(t, http.MethodPost, createProjectEndpoint, defaultContentType, token, requestBody)
	defer r.Body.Close()
	got := r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "expected 201 when creating a valid project")

	// upload the file - later we will try to modify it
	r = uploadOrUpdateTestFile(t, token, "foo.txt", "foobar", uploadFilesEndpoint, "/root/bar")
	defer r.Body.Close()
	got = r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "Expected 201 when uploading a file")
	
	type args struct {
		name string
		token string
		fileName string
		projectName string
		newFileData []byte
		want int // http status code
	}

	tests := []args{
		{
			name: "File update should return 204 when token is valid, file belongs to given project, and project exists",
			token: token,
			fileName: "foo.txt",
			projectName: "foobar",
			newFileData: []byte("This is extra data for testing.\n"),
			want: http.StatusNoContent,
		},
		{
			name: "File update should return 404 when the project does not exist",
			token: token,
			fileName: "foo.txt",
			projectName: "barbaz",
			newFileData: []byte("This is extra data for testing.\n"),
			want: http.StatusNotFound,
		},
		{
			name: "File update should return 404 when the file does not exist in the project",
			token: token,
			fileName: "bar.jpg",
			projectName: "foobar",
			newFileData: []byte("This is extra data for testing.\n"),
			want: http.StatusNotFound,
		},
		{
			name: "File update should return 401 with a malformed token",
			token: token+"foobar123",
			fileName: "foo.txt",
			projectName: "foobar",
			newFileData: []byte("This is extra data for testing.\n"),
			want: http.StatusUnauthorized,
		},
		{
			name: "File update should return 401 with a nonexistent token",
			token: "",
			fileName: "foo.txt",
			projectName: "foobar",
			newFileData: []byte("This is extra data for testing.\n"),
			want: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := uploadOrUpdateTestFile(t, tt.token, tt.fileName, tt.projectName, updateFilesEndpoint, "")
			defer r.Body.Close()
			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected %d from update file call, but got %d", tt.want, got))
		})
	}
}

func TestDownloadFile(t *testing.T) {
	// Setup test by:
	// 1. Writing a test file to disk
	// 2. Creating a user
	// 3. Creating a project for that user
	defer clearTestUser(t, signupManager.db)
	deleteFunc := writeFileForTesting(t, "This is a file used for testing.")
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer deleteFunc(t, cwd+"/foo.txt")
	token, statusCode := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, true)
	assert.Equal(t, http.StatusCreated, statusCode, "Signup status code should be 201 when successful")

	// create the project for the user
	requestBody := bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`))
	r := makeRequest(t, http.MethodPost, createProjectEndpoint, defaultContentType, token, requestBody)
	defer r.Body.Close()
	got := r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "expected 201 when creating a valid project")

	// upload the file - later we will try to download it in different scenarios
	r = uploadOrUpdateTestFile(t, token, "foo.txt", "foobar", uploadFilesEndpoint, "/root/bar")
	defer r.Body.Close()
	got = r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "Expected 201 when uploading a file")

	type args struct {
		name string
		token string
		fileName string
		projectName string
		endpoint string
		want int // http status code
	}

	tests := []args{
		{
			name: "Test file download returns 200 with existing file, existing project, and valid token",
			token: token,
			fileName: "foo.txt",
			projectName: "foobar",
			endpoint: downloadFilesEndpoint+"?name=%s&project_name=%s",
			want: http.StatusOK,
		},
		{
			name: "Test file download returns 404 with nonexistent file",
			token: token,
			fileName: "bar.txt",
			projectName: "foobar",
			endpoint: downloadFilesEndpoint+"?name=%s&project_name=%s",
			want: http.StatusNotFound,
		},
		{
			name: "Test file download returns 404 with nonexistent project",
			token: token,
			fileName: "foo.txt",
			projectName: "barbaz",
			endpoint: downloadFilesEndpoint+"?name=%s&project_name=%s",
			want: http.StatusNotFound,
		},
		{
			name: "Test file download returns 400 with missing project_name field",
			token: token,
			fileName: "foo.txt",
			projectName: "",
			endpoint: downloadFilesEndpoint+"?name=%s&%s",
			want: http.StatusBadRequest,
		},
		{
			name: "Test file download returns 401 with malformed token",
			token: token+"foobar123",
			fileName: "foo.txt",
			projectName: "foobar",
			endpoint: downloadFilesEndpoint+"?name=%s&project_name=%s",
			want: http.StatusUnauthorized,
		},
		{
			name: "Test file download returns 401 with nonexistent token",
			token: "",
			fileName: "foo.txt",
			projectName: "foobar",
			endpoint: downloadFilesEndpoint+"?name=%s&project_name=%s",
			want: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formattedEndpoint := fmt.Sprintf(tt.endpoint, tt.fileName, tt.projectName)
			r := makeRequest(
				t,
				http.MethodGet,
				formattedEndpoint,
				defaultContentType,
				tt.token,
				bytes.NewReader([]byte("")),
			)

			defer r.Body.Close()
			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected %d on file download, but got %d", tt.want, got))

			if got == http.StatusOK {
				// file is 32 bytes
				contentLength := r.Header.Get("Content-Length")
				assert.NotEmpty(t, contentLength)
				assert.Equal(t, "32", contentLength, fmt.Sprintf("expected content length 32, but got %s", contentLength))
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	// Setup test by:
	// 1. Writing a test file to disk
	// 2. Creating a user
	// 3. Creating a project for that user
	defer clearTestUser(t, signupManager.db)
	deleteFunc := writeFileForTesting(t, "This is a file used for testing.")
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer deleteFunc(t, cwd+"/foo.txt")
	token, statusCode := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, true)
	assert.Equal(t, http.StatusCreated, statusCode, "Signup status code should be 201 when successful")

	// create the project for the user
	requestBody := bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`))
	r := makeRequest(t, http.MethodPost, createProjectEndpoint, defaultContentType, token, requestBody)
	defer r.Body.Close()
	got := r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "expected 201 when creating a valid project")

	// upload the file - later we will try to delete it in different scenarios
	r = uploadOrUpdateTestFile(t, token, "foo.txt", "foobar", uploadFilesEndpoint, "/root/bar")
	defer r.Body.Close()
	got = r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "Expected 201 when uploading a file")

	type args struct {
		name string
		token string
		fileName string
		filePath string
		projectName string
		body string
		want int // http status code
	}

	tests := []args{
		{
			name: "Test delete file should return 404 when the file does not exist at the given path",
			token: token,
			fileName: "foo.txt",
			filePath: "/root/boo", // file exists in project, but not at this path
			projectName: "foobar",
			body: `{"name": "%s", "path": "%s", "project_name": "%s"}`,
			want: http.StatusNotFound,
		},
		{
			name: "Test delete file should return 404 when the file does not exist at all",
			token: token,
			fileName: "baz.txt", // no such file at all
			filePath: "/root/bat",
			projectName: "foobar",
			body: `{"name": "%s", "path": "%s", "project_name": "%s"}`,
			want: http.StatusNotFound,
		},
		{
			name: "Test delete file should return 204 with a valid token, existing project, and existing file",
			token: token,
			fileName: "foo.txt",
			filePath: "/root/bar",
			projectName: "foobar",
			body: `{"name": "%s", "path": "%s", "project_name": "%s"}`,
			want: http.StatusNoContent,
		},
		{
			name: "Test delete file should return 400 when not all params are provided in request body",
			token: token,
			fileName: "foo.txt",
			filePath: "/root/bar",
			projectName: "foobar",
			body: `{"name": "%s", "path": "%s"}`,
			want: http.StatusBadRequest,
		},
		{
			name: "Test delete file should return 401 when token is malformed",
			token: token+"foobar123",
			fileName: "foo.txt",
			filePath: "/root/bar",
			projectName: "foobar",
			body: `{"name": "%s", "path": "%s", "project_name": "%s"}`,
			want: http.StatusUnauthorized,
		},
		{
			name: "Test delete file should return 401 when token does not exist",
			token: "",
			fileName: "foo.txt",
			filePath: "/root/bar",
			projectName: "foobar",
			body: `{"name": "%s", "path": "%s", "project_name": "%s"}`,
			want: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := fmt.Sprintf(tt.body, tt.fileName, tt.filePath, tt.projectName)
			body := bytes.NewReader([]byte(b))
			r := makeRequest(t, http.MethodDelete, deleteFilesEndpoint, defaultContentType, tt.token, body)
			defer r.Body.Close()

			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expected %d on delete file call, but got %d", tt.want, got))

			if got == http.StatusNoContent {
				queryString := "?project_name=foobar"
				r := makeRequest(t, http.MethodGet, viewProjectEndpoint+queryString, defaultContentType, tt.token, bytes.NewReader([]byte{}))
				defer r.Body.Close()

				b, err := io.ReadAll(r.Body)
				assert.Nil(t, err)

				assert.NotContains(t, string(b), tt.fileName)
			}
		})
	}
}

func TestFileSharingByUser(t *testing.T) {
	// Setup test by:
	// 1. Writing a test file to disk
	// 2. Creating a user
	// 3. Creating a project for that user
	defer clearTestUser(t, signupManager.db)
	deleteFunc := writeFileForTesting(t, "This is a file used for testing.")
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer deleteFunc(t, cwd+"/foo.txt")
	token, statusCode := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, true)
	assert.Equal(t, http.StatusCreated, statusCode, "Signup status code should be 201 when successful")

	// create the project for the user
	requestBody := bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`))
	r := makeRequest(t, http.MethodPost, createProjectEndpoint, defaultContentType, token, requestBody)
	defer r.Body.Close()
	got := r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "expected 201 when creating a valid project")

	// upload the file - later we will try to delete it in different scenarios
	r = uploadOrUpdateTestFile(t, token, "foo.txt", "foobar", uploadFilesEndpoint, "/root/bar")
	defer r.Body.Close()
	got = r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "Expected 201 when uploading a file")
	
	// get the project id for the file sharing call
	projectId := getProjectIdForTesting(t, signupManager.db, "foobar")

	type args struct {
		name string
		token string
		fileName string
		projectName string
		projectId int
		body string
		want int // http status code
	}

	tests := []args{
		{
			name: "Test sharing a file should return 200 on successful call",
			token: token,
			fileName: "foo.txt",
			projectName: "foobar",
			projectId: projectId,
			body: `{"name": "%s", "project_name": "%s", "project_id": %d}`,
			want: http.StatusOK,
		},
		{
			name: "Test sharing a file should return 404 on a nonexistent file",
			token: token,
			fileName: "bar.txt",
			projectName: "foobar",
			projectId: projectId,
			body: `{"name": "%s", "project_name": "%s", "project_id": %d}`,
			want: http.StatusNotFound,
		},
		{
			name: "Test sharing a file should return 404 on a nonexistent project",
			token: token,
			fileName: "foo.txt",
			projectName: "barbar",
			projectId: projectId,
			body: `{"name": "%s", "project_name": "%s", "project_id": %d}`,
			want: http.StatusNotFound,	
		},
		{
			name: "Test sharing a file should return 400 when request body is invalid",
			token: token,
			fileName: "foo.txt",
			projectName: "foobar",
			projectId: projectId,
			body: `{"name": "%s", "project_name": "%s"}`,
			want: http.StatusBadRequest,
		},
		{
			name: "Test sharing a file should return 401 with a malformed token",
			token: token+"foobar123",
			fileName: "foo.txt",
			projectName: "foobar",
			projectId: projectId,
			body: `{"name": "%s", "project_name": "%s", "project_id": %d}`,
			want: http.StatusUnauthorized,
		},
		{
			name: "Test sharing a file should return 401 with a nonexistent token",
			token: "",
			fileName: "foo.txt",
			projectName: "foobar",
			projectId: projectId,
			body: `{"name": "%s", "project_name": "%s", "project_id": %d}`,
			want: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bytes.NewReader([]byte(fmt.Sprintf(tt.body, tt.fileName, tt.projectName, tt.projectId)))
			r := makeRequest(t, http.MethodPost, sharingFilesEndpoint, defaultContentType, tt.token, b)
			defer r.Body.Close()

			got := r.StatusCode
			assert.Equal(t, tt.want, got, fmt.Sprintf("expecting %d on file sharing call, but got %d", tt.want, got))
			if got == http.StatusOK {
				b, err := io.ReadAll(r.Body)
				assert.Nil(t, err)
				payload := unMarshalResponseBody(b, t)

				hash, ok := payload["hash"].(string)
				assert.True(t, ok)
				assert.Len(t, hash, 64, "hash should be a 64 character string")
			}
		})
	}
}

func TestGettingSharedFile(t *testing.T) {
	// Setup test by:
	// 1. Writing a test file to disk
	// 2. Creating a user
	// 3. Creating a project for that user
	defer clearTestUser(t, signupManager.db)
	deleteFunc := writeFileForTesting(t, "This is a file used for testing.")
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer deleteFunc(t, cwd+"/foo.txt")
	token, statusCode := signupOrLoginTestUser(t, testUsername, testPassword, signupEndpoint, true)
	assert.Equal(t, http.StatusCreated, statusCode, "Signup status code should be 201 when successful")

	// create the project for the user
	requestBody := bytes.NewReader([]byte(`{"username": "helloworld", "name": "foobar", "directories": {"root": {"files": [], "bar": {"files": []}}}}'}`))
	r := makeRequest(t, http.MethodPost, createProjectEndpoint, defaultContentType, token, requestBody)
	defer r.Body.Close()
	got := r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "expected 201 when creating a valid project")

	// upload the file - next we will share it and later try to get the shared file in different scenarios
	r = uploadOrUpdateTestFile(t, token, "foo.txt", "foobar", uploadFilesEndpoint, "/root/bar")
	defer r.Body.Close()
	got = r.StatusCode
	assert.Equal(t, http.StatusCreated, got, "Expected 201 when uploading a file")

	// share the file
	projectId := getProjectIdForTesting(t, signupManager.db, "foobar")
	body := bytes.NewReader([]byte(fmt.Sprintf(`{"name": "foo.txt", "project_name": "foobar", "project_id": %d}`, projectId)))
	r = makeRequest(t, http.MethodPost, sharingFilesEndpoint, defaultContentType, token, body)
	defer r.Body.Close()

	// get the hash from the response payload
	p, err := io.ReadAll(r.Body)
	assert.Nil(t, err)
	payload := unMarshalResponseBody(p, t)
	hash, ok := payload["hash"].(string)
	assert.True(t, ok)

	// try to get the file via sharing - this one should succeed
	c := http.Client{}
	req, err := http.NewRequest(http.MethodGet, baseUrl+sharedFilesEndpoint, bytes.NewReader([]byte("")))
	assert.Nil(t, err)
	req.Header.Set(sharerHeaderKey, "helloworld")
	req.Header.Set(sharedHashHeaderKey, hash)
	r, err = c.Do(req)
	assert.Nil(t, err)
	defer r.Body.Close()
	assert.Equal(t, http.StatusOK, r.StatusCode, fmt.Sprintf("expected %d when getting shared file, but got %d", http.StatusOK, r.StatusCode))

	// change file - this will generate a different hash
	_ = writeFileForTesting(t, "Some more new data.")
	r = uploadOrUpdateTestFile(t, token, "foo.txt", "foobar", updateFilesEndpoint, "/root/bar")
	defer r.Body.Close()

	// try to get the shared file again - this should not succeed and should result in 410 Gone
	req, err = http.NewRequest(http.MethodGet, baseUrl+sharedFilesEndpoint, bytes.NewReader([]byte("")))
	assert.Nil(t, err)
	req.Header.Set(sharerHeaderKey, "helloworld")
	req.Header.Set(sharedHashHeaderKey, hash)
	r, err = c.Do(req)
	assert.Nil(t, err)
	defer r.Body.Close()
	assert.Equal(t, http.StatusGone, r.StatusCode, fmt.Sprintf("expected %d when getting shared file, but got %d", http.StatusOK, r.StatusCode))	
}
