
# Image URL to use all building/pushing
IMG ?= mctl:latest
GOX ?= $(shell go env GOPATH)/bin/gox
VERSION ?= ${tag}
GIT_COMMIT ?= $(shell git rev-list -1 HEAD)
BUILT ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")


all: manager

test: fmt vet
	go test ./... -v -cover

# Build manager binary
manager: fmt vet
	GO111MODULE=off go get github.com/mitchellh/gox
	${GOX} -osarch=linux/386 \
		  -osarch=linux/amd64 \
		  -osarch=darwin/386 \
		  -osarch=darwin/amd64 \
		  -osarch=windows/amd64 \
		  -output=dist/mctl_{{.OS}}_{{.Arch}} \
		  -ldflags="-X 'mctl/cmd/version.ClientVersion=${VERSION}' -X 'mctl/cmd/version.GitCommit=${GIT_COMMIT}' -X 'mctl/cmd/version.Built=${BUILT}'"
	cd dist/ && gzip *

run: fmt vet
	go run ./main.go

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

#generate:
#	go run github.com/99designs/gqlgen generate ./...

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	#echo $DOCKER_PASS | docker login --username $DOCKER_USER --password-stdin
	docker push ${IMG}

