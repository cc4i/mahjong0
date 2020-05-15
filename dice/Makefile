
# Image URL to use all building/pushing
DOCKER_IMAGE ?= herochinese/dice
DOCKER_TAG ?= $(shell git rev-parse --abbrev-ref HEAD)
GOX ?= $(shell go env GOPATH)/bin/gox
VERSION ?= ${tag}
GIT_COMMIT ?= $(shell git rev-list -1 HEAD)
BUILT ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Docker Hub username / password
DOCKER_PASS =
DOCKER_USER =


all: manager

test: fmt vet
	M_WORK_HOME=./ M_MODE=dev M_LOCAL_TILE_REPO=./repo go test ./... -v -cover

# Build manager binary
manager: fmt vet
	GO111MODULE=off go get github.com/mitchellh/gox
	${GOX} -osarch=linux/386 \
		  -osarch=linux/amd64 \
		  -osarch=darwin/386 \
		  -osarch=darwin/amd64 \
		  -osarch=windows/amd64
		  -output=dist/mctl_{{.OS}}_{{.Arch}} \
		  -ldflags="-X 'dice/utils.ServerVersion=${VERSION}' -X 'dice/utils.GitCommit=${GIT_COMMIT}' -X 'dice/utils.Built=${BUILT}'"
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
	docker build \
		--build-arg VERSION=${VERSION} \
		--build-arg GIT_COMMIT=`git rev-list -1 HEAD` \
		--build-arg BUILT=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		-t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest

# Push the docker image
docker-push:
	#echo $DOCKER_PASS | docker login --username $DOCKER_USER --password-stdin
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}
	docker push ${DOCKER_IMAGE}:latest

