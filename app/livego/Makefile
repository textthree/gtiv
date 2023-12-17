GOCMD ?= go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
BINARY_NAME = livego
BINARY_UNIX = $(BINARY_NAME)_unix
DOCKER_ACC ?= gwuhaolin
DOCKER_REPO ?= livego
TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null)

FFMPEG = /Users/t3/workspace/projects/gtiv/newframe/kit/ffmpeg/ffmpeg

default: all

all: test build dockerize
build:
	$(GOBUILD) -o $(BINARY_NAME) -v -ldflags="-X main.VERSION=$(TAG)"

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run: build
	./$(BINARY_NAME)

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

dockerize:
	docker build -t $(DOCKER_ACC)/$(DOCKER_REPO):$(TAG) .
	docker push $(DOCKER_ACC)/$(DOCKER_REPO):$(TAG)

push:
	$(FFMPEG) -re -stream_loop -1 -i videos/japan.flv -c copy -f flv rtmp://localhost:2011/live/rfBd56ti2SMtYvSgD5xAV0YU99zampta7Z7S575KLkIZ9PYk

push2:
	$(FFMPEG) -re -stream_loop -1 -i videos/demo4m.mp4 -c copy -f flv rtmp://localhost:2011/live/L17LTlsVqMNTZyLKMIFSD2x28MlgPJ0SDZVHnHJPxMKi0tWx、

push3:
	$(FFMPEG) -re -stream_loop -1 -i videos/japan.flv -c copy -f flv rtmp://192.168.1.200:2011/live/NUxhZjX5vD1WNYbkBBesIoEtkMR0uFWbUAiVpuXf3dq9xN3I