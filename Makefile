# These are all run as part of local git pre-commit hooks,
# but can obviously also be run separately.

# run all tests
test:
	./run-tests.sh

# run linting
lint:
	golangci-lint run ./...

# run gofmt
fmt:
	gofmt -l .

# run go vet
vet:
	go vet ./...