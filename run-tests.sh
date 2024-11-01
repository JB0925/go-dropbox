#!/bin/bash

# Run the Go application
echo "Starting Go application..."
if [[ $HOSTNAME != *.local ]]; then
    # for CI environment
    export DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
fi

export MAX_REQUESTS=5000
go run . &
app_pid=$!  # Save the PID of the go process

# Wait a moment to ensure the application starts
sleep 1

# Run tests with coverage
echo "Running tests with coverage..."
go test --cover ./...
test_exit_code=$?  # Capture the exit code of go test

# Kill the running application (handled by trap on exit)
echo "Killing the Go application..."
kill -9 $app_pid

# Ensure all processes named go-dropbox are terminated (optional)
pkill -f go-dropbox

# Unset the MAX_REQUESTS environment variable
unset MAX_REQUESTS

# Exit with the test result code to signal pass/fail status
exit $test_exit_code
