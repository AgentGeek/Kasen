PLATFORMS=linux
ARCHITECTURES=386 amd64
LDFLAGS=-ldflags="-s -w"

default: build

all: vet test build build-view

vet:
	go vet

test:
	go test ./... -v -timeout 10m

pprof-profile:
	go tool pprof -http=:41001 http://localhost:42072/debug/pprof/profile

pprof-heap:
	go tool pprof -http=:41002 http://localhost:42072/debug/pprof/heap

build:
	$(foreach GOARCH,$(ARCHITECTURES),\
		$(shell GOOS=$(GOOS) GOARCH=$(GOARCH))\
		$(shell go build $(LDFLAGS) -o bin/$(GOARCH)/kasen)\
		$(shell go build $(LDFLAGS) -o bin/$(GOARCH)/kasen-image internal/cmd/image/image.go)\
	)\

build-web:
	cd web && yarn && yarn prod

run:
	cd bin && ./kasen_linux-amd64

dev:
	cd bin && ./kasen_linux-amd64 -mode=development

dev-web:
	cd web && yarn && yarn dev

.EXPORT_ALL_VARIABLES:
MALLOC_ARENA_MAX=2