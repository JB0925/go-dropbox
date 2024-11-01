#!/bin/bash

# Run the Go application
echo "Starting Go application..."
if [[ $HOSTNAME != *.local ]]; then
    # for CI environment
    export DATABASE_URL=postgresql:///postgres?sslmode=disable
fi

export MAX_REQUESTS=5000; go run . &
app_pid=$!  # Save the PID of the go process

# Wait a moment to ensure the application starts
sleep 1

# Run tests with coverage
echo "Running tests with coverage..."
go test --cover ./...

# Kill the running application
echo "Killing the Go application..."
kill -9 $app_pid

# Ensure all processes named go-dropbox are terminated (optional)
pkill -f go-dropbox
unset MAX_REQUESTS

echo "Script completed."