
# Image URL to use all building/pushing
VERSION ?= ${tag}
GIT_COMMIT ?= $(shell git rev-list -1 HEAD)
BUILT ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")



all: manager

test: fmt vet
	go test ./... -v -cover

# Build manager binary
manager: fmt vet
	go build -a -installsuffix cgo -o probe

run: fmt vet
	go run ./main.go

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...


