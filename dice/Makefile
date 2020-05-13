
# Image URL to use all building/pushing
DOCKER_IMAGE ?= herochinese/dice:latest
# Docker Hub username / password
DOCKER_PASS =
DOCKER_USER =

all: manager

test: fmt vet
	M_WORK_HOME=./ M_MODE=dev M_LOCAL_TILE_REPO=./repo go test ./... -v -cover

# Build manager binary
manager: fmt vet
	go build -o dice main.go

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
docker-build: test
	docker build \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		-t ${DOCKER_IMAGE} .

# Push the docker image
docker-push:
	echo $DOCKER_PASS | docker login --username $DOCKER_USER --password-stdin
	docker push ${IMG}

